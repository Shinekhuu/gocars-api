package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gocars-api/internal/vehicle/repository/postgresql/model"
	vehiclerepo "gocars-api/internal/vehicle/repository/postgresql"

	"go.uber.org/zap"
)

type ManufacturerService struct {
	repo *vehiclerepo.ManufacturerRepository
}

func NewManufacturerService(repo *vehiclerepo.ManufacturerRepository) *ManufacturerService {
	return &ManufacturerService{repo: repo}
}

func (s *ManufacturerService) GetAll() ([]model.Manufacturer, error) {
	return s.repo.GetAll()
}

func (s *ManufacturerService) GetByName(name string) (*model.Manufacturer, error) {
	return s.repo.GetByName(name)
}

func (s *ManufacturerService) FillFromFile(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	var data model.ManufacturerData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return 0, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	for i := range data.Manufacturers {
		if err := s.repo.Upsert(&data.Manufacturers[i]); err != nil {
			zap.L().Error("failed to upsert manufacturer",
				zap.String("name", data.Manufacturers[i].ManufacturerName),
				zap.Error(err),
			)
		}
	}

	return len(data.Manufacturers), nil
}
