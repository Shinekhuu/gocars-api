package controllers

import (
	"gocars-api/models"
	"gocars-api/repositories"
	"gocars-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Shop(c *gin.Context) {
	vehicleIDStr := c.DefaultQuery("vehicle_id", "0")
	categoryIDStr := c.DefaultQuery("category_id", "100260")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "40")

	vehicleID := utils.AtoiUint(vehicleIDStr)
	categoryID := utils.AtoiUint(categoryIDStr)
	page := utils.Atoi(pageStr)
	limit := utils.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 40
	}

	// Try reading from DB first
	articles, total, err := repositories.GetArticleItemsByVehicleIdAndCategoryId(vehicleID, categoryID, page, limit)
	if err == nil && total > 0 {
		c.JSON(http.StatusOK, gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"articles": articles,
			"api":      "db",
		})
		return
	}

	// Otherwise, fetch from API
	apiData, err := models.GetArticleItemsFromRapidAPI(vehicleID, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply pagination manually (API returns full list)
	start := (page - 1) * limit
	end := start + limit
	if start > len(apiData.Articles) {
		start = len(apiData.Articles)
	}
	if end > len(apiData.Articles) {
		end = len(apiData.Articles)
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"limit":    limit,
		"total":    len(apiData.Articles),
		"articles": apiData.Articles[start:end],
		"api":      "api",
	})
}

// func FillArticleItemData(c *gin.Context) {
// 	fileName := c.DefaultQuery("filename", "100001.json")
// 	filePath := "/home/api/data/" + fileName

// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer file.Close()

// 	bytes, err := io.ReadAll(file)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	var data models.VehicleArticlesResponse
// 	if err := json.Unmarshal(bytes, &data); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	go func(data models.VehicleArticlesResponse) {

// 		tx := database.DB

// 		// =========================
// 		// 1. Upsert Articles (bulk)
// 		// =========================

// 		var articles []models.ArticleItem
// 		for _, a := range data.Articles {
// 			articles = append(articles, a)
// 		}

// 		if len(articles) > 0 {
// 			tx.Clauses(clause.OnConflict{
// 				Columns: []clause.Column{{Name: "article_id"}},
// 				DoUpdates: clause.AssignmentColumns([]string{
// 					"article_no", "brand_name",
// 				}),
// 			}).Create(&articles)
// 		}

// 		// =========================
// 		// 2. Build ArticleVehicles
// 		// =========================

// 		vehicleMap := map[uint]struct{}{}
// 		var articleVehicles []models.ArticleVehicles

// 		for _, a := range articles {
// 			if a.ID == 0 {
// 				continue
// 			}

// 			key := a.ID
// 			if _, ok := vehicleMap[key]; ok {
// 				continue
// 			}
// 			vehicleMap[key] = struct{}{}

// 			articleVehicles = append(articleVehicles, models.ArticleVehicles{
// 				ArticleItemID: a.ID,
// 				VehicleID:     data.VehicleID,
// 			})
// 		}

// 		// bulk insert
// 		if len(articleVehicles) > 0 {
// 			tx.Clauses(clause.OnConflict{
// 				Columns: []clause.Column{
// 					{Name: "article_item_id"},
// 					{Name: "vehicle_id"},
// 				},
// 				DoNothing: true,
// 			}).Create(&articleVehicles)
// 		}

// 		// =========================
// 		// 3. Build ArticleCategory
// 		// =========================

// 		catMap := map[string]struct{}{}
// 		var articleCategories []models.ArticleCategory

// 		for _, a := range articles {
// 			if a.ID == 0 {
// 				continue
// 			}

// 			// global category
// 			if data.CategoryID != 0 {
// 				key := fmt.Sprintf("%d-%d", a.ID, data.CategoryID)
// 				if _, ok := catMap[key]; !ok {
// 					catMap[key] = struct{}{}
// 					articleCategories = append(articleCategories, models.ArticleCategory{
// 						ArticleItemID: a.ID,
// 						CategoryID:    data.CategoryID,
// 					})
// 				}
// 			}

// 			// per-article categories
// 			for _, cat := range a.AllCategories {
// 				if cat.CategoryID == 0 {
// 					continue
// 				}

// 				key := fmt.Sprintf("%d-%d", a.ID, cat.CategoryID)
// 				if _, ok := catMap[key]; ok {
// 					continue
// 				}
// 				catMap[key] = struct{}{}

// 				articleCategories = append(articleCategories, models.ArticleCategory{
// 					ArticleItemID: a.ID,
// 					CategoryID:    cat.CategoryID,
// 				})
// 			}
// 		}

// 		if len(articleCategories) > 0 {
// 			tx.Clauses(clause.OnConflict{
// 				Columns: []clause.Column{
// 					{Name: "article_item_id"},
// 					{Name: "category_id"},
// 				},
// 				DoNothing: true,
// 			}).Create(&articleCategories)
// 		}

// 	}(data)

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": fmt.Sprintf("✅ %d articles queued for vehicle %d", len(data.Articles), data.VehicleID),
// 	})
// }
