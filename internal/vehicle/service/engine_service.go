package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gocars-api/internal/vehicle/repository/postgresql/model"
	vehiclerepo "gocars-api/internal/vehicle/repository/postgresql"

	"go.uber.org/zap"
)

type EngineService struct {
	repo *vehiclerepo.EngineRepository
}

func NewEngineService(repo *vehiclerepo.EngineRepository) *EngineService {
	return &EngineService{repo: repo}
}

func (s *EngineService) GetByModelID(modelID uint) (*model.EngineResponse, error) {
	engines, err := s.repo.GetByModelID(modelID)
	if err != nil {
		return nil, err
	}
	return &model.EngineResponse{Engines: engines, CountModelTypes: len(engines)}, nil
}

func (s *EngineService) FetchFromAPI(manufacturerID, modelID uint) (*model.EngineResponse, error) {
	url := fmt.Sprintf(
		"https://%s/types/type-id/1/list-vehicles-types/%d/lang-id/4/country-filter-id/125",
		os.Getenv("X_RAPIDAPI_HOST"), modelID,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-key", os.Getenv("X_RAPIDAPI_KEY"))
	req.Header.Set("x-rapidapi-host", os.Getenv("X_RAPIDAPI_HOST"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var result model.EngineResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	for i := range result.Engines {
		result.Engines[i].ManufacturerID = manufacturerID
		result.Engines[i].ModelID = modelID
		result.Engines[i].IsFetched = true
	}
	result.CountModelTypes = len(result.Engines)

	if err := s.repo.UpsertMany(result.Engines); err != nil {
		zap.L().Error("failed to persist engines from API", zap.Error(err))
	}

	return &result, nil
}

// GetByName tries the DB first, falls back to RapidAPI, then filters by frame.
func (s *EngineService) GetByName(manufacturerID, modelID uint, frame string) ([]model.Engine, error) {
	frame = strings.ToUpper(frame)

	engines, err := s.repo.GetByNameLike(manufacturerID, modelID, frame)
	if err == nil && len(engines) > 0 {
		return engines, nil
	}

	apiResp, apiErr := s.FetchFromAPI(manufacturerID, modelID)
	if apiErr != nil {
		return nil, fmt.Errorf("DB lookup failed: %v; RapidAPI fetch failed: %v", err, apiErr)
	}

	var matched []model.Engine
	for _, e := range apiResp.Engines {
		if strings.Contains(strings.ToUpper(e.TypeEngineName), frame) {
			matched = append(matched, e)
		}
	}

	if len(matched) > 0 {
		return matched, nil
	}

	return nil, fmt.Errorf("engine not found for manufacturer %d, model %d, frame %s", manufacturerID, modelID, frame)
}

func (s *EngineService) GetByTypeEngineNames(names []string) ([]model.Engine, error) {
	return s.repo.GetByTypeEngineNames(names)
}
