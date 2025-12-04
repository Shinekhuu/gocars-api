package models

import (
	"gocars-api/database"

	"gorm.io/gorm"
)

type Manufacturer struct {
	gorm.Model
	ManufacturerID   uint   `gorm:"uniqueIndex" json:"manufacturerId"`
	ManufacturerName string `json:"manufacturerName"`
}

// Struct for the whole JSON file
type ManufacturerData struct {
	CountManufactures int            `json:"countManufactures"`
	Manufacturers     []Manufacturer `json:"manufacturers"`
}

func GetManufacturerByName(manufacturerName string) (*Manufacturer, error) {
	var dbManufacturer Manufacturer
	if err := database.DB.Where("manufacturer_name = ?", manufacturerName).First(&dbManufacturer).Error; err != nil {
		return nil, err
	}
	return &dbManufacturer, nil
}
