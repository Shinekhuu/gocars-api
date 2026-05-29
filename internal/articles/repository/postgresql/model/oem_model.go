package model

import "gorm.io/gorm"

type OemArticleResponse struct {
	Articles []ArticleItem `json:"articles"`
}

type Oem struct {
	gorm.Model
	Brand          string `json:"brand" gorm:"type:varchar(255);uniqueIndex:idx_oem"`
	DisplayNo      string `json:"displayNo" gorm:"type:varchar(255);uniqueIndex:idx_oem"`
	DisplayNoClean string `json:"-" gorm:"type:varchar(255);index"`
}
