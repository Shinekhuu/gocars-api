package repository

import (
	"gocars-api/internal/vehicle/repository/postgresql/model"

	"gorm.io/gorm"
)

type ManufacturerRepository struct {
	db *gorm.DB
}

func NewManufacturerRepository(db *gorm.DB) *ManufacturerRepository {
	return &ManufacturerRepository{db: db}
}

func (r *ManufacturerRepository) GetAll() ([]model.Manufacturer, error) {
	var manufacturers []model.Manufacturer
	err := r.db.Find(&manufacturers).Error
	return manufacturers, err
}

func (r *ManufacturerRepository) GetByName(name string) (*model.Manufacturer, error) {
	var m model.Manufacturer
	if err := r.db.Where("manufacturer_name = ?", name).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ManufacturerRepository) Upsert(m *model.Manufacturer) error {
	return r.db.Where(model.Manufacturer{ManufacturerID: m.ManufacturerID}).Assign(m).FirstOrCreate(m).Error
}
