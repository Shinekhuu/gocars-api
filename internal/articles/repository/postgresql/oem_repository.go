package repository

import (
	articles "gocars-api/internal/articles/repository/postgresql/model"

	"gorm.io/gorm"
)

type OemRepository struct {
	db *gorm.DB
}

func NewOemRepository(db *gorm.DB) *OemRepository {
	return &OemRepository{db: db}
}

func (r *OemRepository) GetWithOEMs(ids []uint, articleIDs []uint) ([]articles.ArticleItem, error) {
	var result []articles.ArticleItem

	if len(ids) == 0 && len(articleIDs) == 0 {
		return result, nil
	}

	query := r.db.Preload("AllOems.Oem")

	if len(ids) > 0 && len(articleIDs) > 0 {
		query = query.Where("id IN ? OR article_id IN ?", ids, articleIDs)
	} else if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	} else {
		query = query.Where("article_id IN ?", articleIDs)
	}

	err := query.Find(&result).Error
	return result, err
}
