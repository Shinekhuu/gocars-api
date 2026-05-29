package jobs

import (
	articles "gocars-api/internal/articles/repository/postgresql/model"
	db "gocars-api/internal/database/mysql"
	"gocars-api/internal/search/meili"
	"gocars-api/internal/search/models"

	"go.uber.org/zap"
)

func processArticle(a articles.ArticleItem) {
	if err := saveMain(a); err != nil {
		zap.L().Error("saveMain failed", zap.Error(err))
		return
	}

	var dbArticle articles.ArticleItem
	if err := db.DB.
		Preload("AllCategories.Category").
		Where("article_id = ?", *a.ArticleID).
		First(&dbArticle).Error; err != nil {
		zap.L().Error("failed to reload article", zap.Error(err))
		return
	}

	a.ID = dbArticle.ID

	go saveEngines(a)
	go indexToMeili(dbArticle)
}

func indexToMeili(a articles.ArticleItem) {
	if meili.Default == nil {
		return
	}

	var categoryID uint
	var categoryName, categoryNameMN string
	if len(a.AllCategories) > 0 {
		cat := a.AllCategories[0].Category
		categoryID = cat.CategoryID
		categoryName = cat.CategoryName
		categoryNameMN = cat.CategoryNameMn
	}

	doc := models.MeiliArticle{
		ID:             a.ID,
		ArticleNo:      a.ArticleNo,
		SearchNo:       a.ArticleSearchNo,
		ProductName:    a.ArticleProductName,
		Supplier:       a.SupplierName,
		Image:          a.S3Image,
		CategoryID:     categoryID,
		CategoryName:   categoryName,
		CategoryNameMN: categoryNameMN,
	}

	if err := meili.Default.IndexDocuments([]models.MeiliArticle{doc}); err != nil {
		zap.L().Error("meili index failed", zap.Uint("id", a.ID), zap.Error(err))
	}
}
