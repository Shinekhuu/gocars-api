package repositories

import (
	"gocars-api/database"
	"gocars-api/models"
)

func FindArticle(id, articleID int) (models.ArticleItem, error) {
	var article models.ArticleItem

	db := database.DB.
		Preload("AllSpecifications").
		Preload("AllOems.Oem")

	if id != 0 {
		db = db.Where("id = ?", id)
	} else {
		db = db.Where("article_id = ?", articleID)
	}

	err := db.First(&article).Error
	return article, err
}

func GetEngines(articleItemID uint, offset, limit int) ([]models.Engine, int64, error) {
	var engines []models.Engine
	var total int64

	database.DB.Model(&models.ArticleVehicles{}).
		Where("article_item_id = ?", articleItemID).
		Count(&total)

	err := database.DB.
		Table("article_vehicles av").
		Select("e.*").
		Joins("LEFT JOIN engines e ON av.vehicle_id = e.vehicle_id").
		Where("av.article_item_id = ?", articleItemID).
		Offset(offset).
		Limit(limit).
		Find(&engines).Error

	return engines, total, err
}
