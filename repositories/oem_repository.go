package repositories

import (
	"gocars-api/database"
	"gocars-api/models"
)

func GetWithOEMs(
	ids []uint,
	articleIDs []uint,
) ([]models.ArticleItem, error) {

	var articles []models.ArticleItem

	if len(ids) == 0 && len(articleIDs) == 0 {
		return articles, nil
	}

	query := database.DB.Preload("AllOems.Oem")

	if len(ids) > 0 && len(articleIDs) > 0 {
		query = query.Where("id IN ? OR article_id IN ?", ids, articleIDs)
	} else if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	} else {
		query = query.Where("article_id IN ?", articleIDs)
	}

	err := query.Find(&articles).Error
	return articles, err
}
