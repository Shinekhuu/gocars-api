package repository

import (
	"gocars-api/internal/vehicle/repository/postgresql/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type XyrRepository struct {
	db *gorm.DB
}

func NewXyrRepository(db *gorm.DB) *XyrRepository {
	return &XyrRepository{db: db}
}

func (r *XyrRepository) GetByPlate(plate string) (*model.Xyr, error) {
	var xyr model.Xyr
	if err := r.db.Where("plate_number = ?", plate).First(&xyr).Error; err != nil {
		return nil, err
	}
	return &xyr, nil
}

func (r *XyrRepository) Upsert(xyr *model.Xyr) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "plate_number"}},
		UpdateAll: true,
	}).Create(xyr).Error
}

func (r *XyrRepository) GetXyrVehicle(xyrID uint) (*model.XyrVehicle, error) {
	var xv model.XyrVehicle
	if err := r.db.Where("xyr_id = ?", xyrID).First(&xv).Error; err != nil {
		return nil, err
	}
	return &xv, nil
}

func (r *XyrRepository) UpsertXyrVehicle(xyrID, vehicleID uint) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "xyr_id"}, {Name: "vehicle_id"}},
		DoNothing: true,
	}).Create(&model.XyrVehicle{XyrID: xyrID, VehicleID: vehicleID}).Error
}
