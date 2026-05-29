package model

import "gorm.io/gorm"

type Manufacturer struct {
	gorm.Model
	ManufacturerID   uint   `gorm:"uniqueIndex" json:"manufacturerId"`
	ManufacturerName string `json:"manufacturerName"`
}

type ManufacturerData struct {
	CountManufactures int            `json:"countManufactures"`
	Manufacturers     []Manufacturer `json:"manufacturers"`
}
