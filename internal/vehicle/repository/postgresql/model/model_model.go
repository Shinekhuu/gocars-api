package model

import "gorm.io/gorm"

type ModelResponse struct {
	CountModels int     `json:"countModels"`
	Models      []Model `json:"models"`
}

type Model struct {
	gorm.Model
	ManufacturerID uint          `json:"manufacturerId"`
	ModelID        uint          `gorm:"uniqueIndex" json:"modelId"`
	ModelName      string        `json:"modelName"`
	Families       []ModelFamily `gorm:"foreignKey:ModelID" json:"families"`
	ModelYearFrom  string        `json:"modelYearFrom"`
	ModelYearTo    *string       `json:"modelYearTo"`
}

type ModelFamily struct {
	gorm.Model
	ModelID    uint   `gorm:"column:model_id" json:"modelId"`
	FamilyCode string `gorm:"column:family_code;type:varchar(50)" json:"familyCode"`
}
