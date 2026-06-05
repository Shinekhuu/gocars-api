package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gocars-api/internal/shared/utils"
	"gocars-api/internal/vehicle/repository/postgresql/model"
	vehiclerepo "gocars-api/internal/vehicle/repository/postgresql"

	"go.uber.org/zap"
)

type ModelService struct {
	repo *vehiclerepo.ModelRepository
}

func NewModelService(repo *vehiclerepo.ModelRepository) *ModelService {
	return &ModelService{repo: repo}
}

func (s *ModelService) GetByManufacturerID(manufacturerID uint) (*model.ModelResponse, error) {
	models, err := s.repo.GetByManufacturerID(manufacturerID)
	if err != nil {
		return nil, err
	}
	return &model.ModelResponse{Models: models, CountModels: len(models)}, nil
}

func (s *ModelService) FetchFromAPI(manufacturerID uint) (*model.ModelResponse, error) {
	url := fmt.Sprintf(
		"https://%s/models/list/type-id/1/manufacturer-id/%d/lang-id/4/country-filter-id/125",
		os.Getenv("X_RAPIDAPI_HOST"), manufacturerID,
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

	var result model.ModelResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	for i := range result.Models {
		result.Models[i].ManufacturerID = manufacturerID
	}
	result.CountModels = len(result.Models)

	if err := s.repo.UpsertMany(result.Models); err != nil {
		zap.L().Error("failed to persist models from API", zap.Error(err))
	}

	return &result, nil
}

// GetByName tries the DB first, falls back to RapidAPI, then filters by name and build date.
func (s *ModelService) GetByName(manufacturerID uint, modelName, buildDate string) (*model.Model, error) {
	modelName = strings.ToUpper(utils.SplitModelName(modelName))

	m, err := s.repo.GetByName(manufacturerID, modelName, buildDate)
	if err == nil {
		return m, nil
	}

	apiResp, apiErr := s.FetchFromAPI(manufacturerID)
	if apiErr != nil {
		return nil, fmt.Errorf("DB lookup failed: %v; RapidAPI fetch failed: %v", err, apiErr)
	}

	for i := range apiResp.Models {
		m := apiResp.Models[i]
		if strings.HasPrefix(m.ModelName, modelName) {
			if m.ModelYearFrom <= buildDate && (m.ModelYearTo == nil || *m.ModelYearTo >= buildDate) {
				return &m, nil
			}
		}
	}

	return nil, fmt.Errorf("model not found for manufacturer %d and name %s", manufacturerID, modelName)
}

// GetOrFetchByManufacturerID returns models from DB, falling back to RapidAPI if not found.
func (s *ModelService) GetOrFetchByManufacturerID(manufacturerID uint) (*model.ModelResponse, error) {
	resp, err := s.GetByManufacturerID(manufacturerID)
	if err != nil {
		return nil, err
	}
	if resp.CountModels > 0 {
		return resp, nil
	}
	return s.FetchFromAPI(manufacturerID)
}
