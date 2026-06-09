package model

import "gorm.io/gorm"

type EngineResponse struct {
	ModelType       string   `json:"modelType"`
	CountModelTypes int      `json:"countModelTypes"`
	Engines         []Engine `json:"ModelTypes"`
}

type Engine struct {
	gorm.Model
	ManufacturerID            uint           `json:"manufacturerId"`
	ModelID                   uint           `json:"modelId"`
	VehicleID                 uint           `gorm:"uniqueIndex" json:"vehicleId"`
	ManufacturerName          string         `json:"manufacturerName"`
	ModelName                 string         `json:"modelName"`
	TypeEngineName            string         `json:"typeEngineName"`
	ConstructionIntervalStart string         `json:"constructionIntervalStart"`
	ConstructionIntervalEnd   string         `json:"constructionIntervalEnd"`
	PowerKw                   string         `json:"powerKw"`
	PowerPs                   string         `json:"powerPs"`
	CapacityTax               *string        `json:"capacityTax"`
	FuelType                  string         `json:"fuelType"`
	BodyType                  string         `json:"bodyType"`
	NumberOfCylinders         int            `json:"numberOfCylinders"`
	CapacityLt                string         `json:"capacityLt"`
	CapacityTech              string         `json:"capacityTech"`
	EngineCodes               string         `json:"engineCodes"`
	EngID                     uint           `json:"engId"`
	IsFetched                 bool           `gorm:"type:boolean;default:false"`
	Families                  []EngineFamily `gorm:"foreignKey:EngineID;constraint:OnDelete:CASCADE;" json:"families"`
}

type EngineFamily struct {
	gorm.Model
	EngineID   uint   `gorm:"column:engine_id" json:"engineId"`
	FamilyCode string `gorm:"column:family_code;type:varchar(50)" json:"familyCode"`
}
