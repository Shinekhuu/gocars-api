package repository

import (
	"gocars-api/internal/vehicle/repository/postgresql/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ModelRepository struct {
	db *gorm.DB
}

func NewModelRepository(db *gorm.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

func (r *ModelRepository) GetByManufacturerID(manufacturerID uint) ([]model.Model, error) {
	var models []model.Model
	err := r.db.Where("manufacturer_id = ?", manufacturerID).Find(&models).Error
	return models, err
}

func (r *ModelRepository) GetByName(manufacturerID uint, modelName, buildDate string) (*model.Model, error) {
	var m model.Model
	err := r.db.
		Where("manufacturer_id = ?", manufacturerID).
		Where("model_name LIKE ?", modelName+"%").
		Where("model_year_from <= ?", buildDate).
		Where("(model_year_to >= ? OR (model_year_from <= ? AND model_year_to IS NULL))", buildDate, buildDate).
		Order("model_year_from DESC").
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ModelRepository) UpsertMany(models []model.Model) error {
	if len(models) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "model_id"}},
		UpdateAll: true,
	}).Create(&models).Error
}
