package repository

import (
	"errors"
	"time"

	articles "gocars-api/internal/articles/repository/postgresql/model"

	"gorm.io/gorm"
)

type APIFetchLogRepository struct {
	db *gorm.DB
}

func NewAPIFetchLogRepository(db *gorm.DB) *APIFetchLogRepository {
	return &APIFetchLogRepository{db: db}
}

func (r *APIFetchLogRepository) EnsureAPIFetchLog(vehicleID uint, categoryID uint) error {
	return r.db.
		FirstOrCreate(
			&articles.APIFetchLog{},
			articles.APIFetchLog{
				VehicleID:     vehicleID,
				CategoryID:    categoryID,
				LastFetchedAt: time.Now(),
			},
		).Error
}

func (r *APIFetchLogRepository) GetAPIFetchLog(vehicleID uint, categoryID uint) (*articles.APIFetchLog, error) {
	var log articles.APIFetchLog

	err := r.db.
		Where("vehicle_id=? AND category_id=?", vehicleID, categoryID).
		First(&log).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &log, nil
}

func (r *APIFetchLogRepository) TouchAPIFetchLog(tx *gorm.DB, vehicleID uint, categoryID uint) error {
	return tx.
		Where(articles.APIFetchLog{VehicleID: vehicleID, CategoryID: categoryID}).
		Assign(articles.APIFetchLog{LastFetchedAt: time.Now()}).
		FirstOrCreate(&articles.APIFetchLog{}).Error
}
