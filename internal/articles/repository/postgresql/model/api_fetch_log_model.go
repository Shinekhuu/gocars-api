package model

import (
	"time"

	"gorm.io/gorm"
)

type APIFetchLog struct {
	gorm.Model
	VehicleID     uint `gorm:"uniqueIndex:idx_vehicle_category"`
	CategoryID    uint `gorm:"uniqueIndex:idx_vehicle_category"`
	LastFetchedAt time.Time
}
