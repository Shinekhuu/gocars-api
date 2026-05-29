package repository

import (
	articles "gocars-api/internal/articles/repository/postgresql/model"

	"gorm.io/gorm"
)

type DictionaryRepository struct {
	db *gorm.DB
}

func NewDictionaryRepository(db *gorm.DB) *DictionaryRepository {
	return &DictionaryRepository{db: db}
}

func (r *DictionaryRepository) FindDictionaryByKey(dictKey string) (articles.Dictionary, error) {
	var item articles.Dictionary
	err := r.db.Where("dict_key = ?", dictKey).First(&item).Error
	return item, err
}

func (r *DictionaryRepository) InsertDictionaryTranslation(dictionaryID uint, value string) error {
	translation := articles.DictionaryTranslation{
		DictionaryID: dictionaryID,
		LanguageCode: "mn",
		Value:        value,
	}
	return r.db.Create(&translation).Error
}
