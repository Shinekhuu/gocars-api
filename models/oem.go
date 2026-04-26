package models

import "gorm.io/gorm"

type OemArticleResponse struct {
	Articles []ArticleItem `json:"articles"` // must match JSON
}

type Oem struct {
	gorm.Model
	Brand     string `json:"brand" gorm:"type:varchar(255);uniqueIndex:idx_oem"`
	DisplayNo string `json:"displayNo" gorm:"type:varchar(255);uniqueIndex:idx_oem"`

	// 🔥 NEW: cleaned version (for fast search)
	DisplayNoClean string `json:"-" gorm:"type:varchar(255);index"`
}

