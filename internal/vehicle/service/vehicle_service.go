package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	scraperrepo "gocars-api/internal/vehicle/repository"
	"gocars-api/internal/vehicle/repository/postgresql/model"
	vehiclerepo "gocars-api/internal/vehicle/repository/postgresql"

	"gorm.io/gorm"
)

const xypAPIURL = "https://xyp-api.smartcar.mn/xyp-api/v1/xyp/get-data-public"

type VehicleService struct {
	xyrRepo    *vehiclerepo.XyrRepository
	engineRepo *vehiclerepo.EngineRepository
}

func NewVehicleService(xyrRepo *vehiclerepo.XyrRepository, engineRepo *vehiclerepo.EngineRepository) *VehicleService {
	return &VehicleService{xyrRepo: xyrRepo, engineRepo: engineRepo}
}

// GetOrFetchXyr returns a cached Xyr or calls the XYP API and stores the result.
func (s *VehicleService) GetOrFetchXyr(plate string) (*model.Xyr, error) {
	xyr, err := s.xyrRepo.GetByPlate(plate)
	if err == nil {
		return xyr, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return s.fetchAndStoreXyr(plate)
}

func (s *VehicleService) fetchAndStoreXyr(plate string) (*model.Xyr, error) {
	reqBody := map[string]interface{}{
		"serviceCode": "WS100401_getVehicleInfo",
		"customFields": map[string]string{
			"plateNumber": plate,
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(xypAPIURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to call XYP API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return nil, errors.New("XYP API did not return JSON")
	}

	var xyr model.Xyr
	if err := json.Unmarshal(respBytes, &xyr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	xyr.MarkName = strings.ToUpper(xyr.MarkName)
	xyr.ModelName = strings.ToUpper(xyr.ModelName)

	if err := s.xyrRepo.Upsert(&xyr); err != nil {
		return nil, fmt.Errorf("failed to store xyr: %w", err)
	}

	return &xyr, nil
}

// GetOrFetchEngineVehicle finds a cached engine for the xyr, or falls back to scraping.
func (s *VehicleService) GetOrFetchEngineVehicle(xyr *model.Xyr, plate string) (*model.Engine, *model.Vehicle, error) {
	if xv, err := s.xyrRepo.GetXyrVehicle(xyr.ID); err == nil {
		engines, dbErr := s.engineRepo.GetByModelID(xv.VehicleID)
		if dbErr == nil && len(engines) > 0 {
			return &engines[0], nil, nil
		}
	}

	veh, err := scraperrepo.FetchVehicleData(plate)
	if err != nil {
		return nil, nil, err
	}

	if err := s.xyrRepo.UpsertXyrVehicle(xyr.ID, veh.CarID); err != nil {
		return nil, veh, fmt.Errorf("failed to save xyr vehicle: %w", err)
	}

	return nil, veh, nil
}

func (s *VehicleService) UpsertXyrVehicle(xyrID, vehicleID uint) error {
	return s.xyrRepo.UpsertXyrVehicle(xyrID, vehicleID)
}
