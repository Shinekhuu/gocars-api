package model

import "gorm.io/gorm"

type Dictionary struct {
	gorm.Model
	DictKey string `gorm:"type:varchar(255);uniqueIndex"`

	Translations []DictionaryTranslation `gorm:"foreignKey:DictionaryID;references:ID;constraint:OnDelete:CASCADE"`
}

type DictionaryTranslation struct {
	gorm.Model
	DictionaryID uint   `gorm:"not null;uniqueIndex:idx_dict_lang"`
	LanguageCode string `gorm:"type:varchar(5);not null;uniqueIndex:idx_dict_lang"`
	Value        string `gorm:"type:varchar(255);not null;index"`
}
