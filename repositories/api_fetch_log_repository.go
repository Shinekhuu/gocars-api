package repositories

import (
	"errors"
	"time"

	"gocars-api/database"
	"gocars-api/models"

	"gorm.io/gorm"
)

func EnsureAPIFetchLog(
	vehicleID uint,
	categoryID uint,
) error {

	return database.DB.
		FirstOrCreate(
			&models.APIFetchLog{},
			models.APIFetchLog{
				VehicleID:     vehicleID,
				CategoryID:    categoryID,
				LastFetchedAt: time.Now(),
			},
		).Error
}

func GetAPIFetchLog(
	vehicleID uint,
	categoryID uint,
) (*models.APIFetchLog, error) {

	var log models.APIFetchLog

	err := database.DB.
		Where(
			"vehicle_id=? AND category_id=?",
			vehicleID,
			categoryID,
		).
		First(&log).Error

	if errors.Is(
		err,
		gorm.ErrRecordNotFound,
	) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &log, nil
}

func TouchAPIFetchLog(
	tx *gorm.DB,
	vehicleID uint,
	categoryID uint,
) error {

	return tx.
		Where(
			models.APIFetchLog{
				VehicleID:  vehicleID,
				CategoryID: categoryID,
			},
		).
		Assign(
			models.APIFetchLog{
				LastFetchedAt: time.Now(),
			},
		).
		FirstOrCreate(
			&models.APIFetchLog{},
		).Error
}
