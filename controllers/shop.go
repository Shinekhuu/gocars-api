package controllers

import (
	"encoding/json"
	"fmt"
	"gocars-api/database"
	"gocars-api/models"
	"gocars-api/utils"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

func Shop(c *gin.Context) {
	vehicleIDStr := c.DefaultQuery("vehicle_id", "10538")
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
	articles, total, err := models.GetArticleItemsByVehicleIdAndCategoryId(vehicleID, categoryID, page, limit)
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

func FillArticleItemData(c *gin.Context) {
	fileName := c.DefaultQuery("filename", "100001.json")
	filePath := "/home/api/data/" + fileName

	// 1️⃣ Read JSON file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open JSON file %s: %v", fileName, err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read JSON file %s: %v", fileName, err)
	}

	// 2️⃣ Parse JSON
	var data models.VehicleArticlesResponse
	if err := json.Unmarshal(bytes, &data); err != nil {
		log.Fatalf("failed to unmarshal JSON %s: %v", fileName, err)
	}

	fmt.Printf("Found %d articles in JSON\n", len(data.Articles))

	go func() {
		// 3️⃣ Ensure Vehicle exists
		var engine models.Engine
		if err := database.DB.
			Where(models.Engine{VehicleID: data.VehicleID}).
			Assign(engine).
			FirstOrCreate(&engine).Error; err != nil {
			log.Fatalf("failed to upsert vehicle %v: %v", data.VehicleID, err)
		}

		// 4️⃣ Upsert ArticleItems (1-by-1)
		for i := range data.Articles {
			article := &data.Articles[i]
			if err := database.DB.
				Where(models.ArticleItem{ArticleID: article.ArticleID}).
				Assign(article).
				FirstOrCreate(article).Error; err != nil {
				log.Printf("❌ Failed to upsert article %v: %v", article.ArticleID, err)
			}
		}

		// 5️⃣ Batch upsert ArticleVehicles (link articles to vehicle)
		var articleVehicles []models.ArticleVehicles
		for _, article := range data.Articles {
			articleVehicles = append(articleVehicles, models.ArticleVehicles{
				VehicleID: data.VehicleID,
				ArticleID: article.ArticleID,
			})
		}

		batchSize := 500
		for i := 0; i < len(articleVehicles); i += batchSize {
			end := i + batchSize
			if end > len(articleVehicles) {
				end = len(articleVehicles)
			}
			batch := articleVehicles[i:end]

			if err := database.DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "vehicle_id"}, {Name: "article_id"}},
				DoNothing: true, // skip duplicates
			}).Create(&batch).Error; err != nil {
				log.Printf("❌ Failed to batch upsert VehicleArticles: %v", err)
			}
		}

		// 6️⃣ Batch upsert ArticleCategory
		var articleCategories []models.ArticleCategory
		for _, article := range data.Articles {
			// Use VehicleArticlesResponse.CategoryID if exists, otherwise loop article.AllCategories
			if data.CategoryID != 0 {
				articleCategories = append(articleCategories, models.ArticleCategory{
					ArticleID:  article.ArticleID,
					CategoryID: data.CategoryID,
				})
			} else {
				for _, cat := range article.AllCategories {
					articleCategories = append(articleCategories, models.ArticleCategory{
						ArticleID:  article.ArticleID,
						CategoryID: cat.CategoryID,
					})
				}
			}
		}

		for i := 0; i < len(articleCategories); i += batchSize {
			end := i + batchSize
			if end > len(articleCategories) {
				end = len(articleCategories)
			}
			batch := articleCategories[i:end]

			if err := database.DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "article_id"}, {Name: "category_id"}},
				DoNothing: true, // skip duplicates
			}).Create(&batch).Error; err != nil {
				log.Printf("❌ Failed to batch upsert ArticleCategories: %v", err)
			}
		}

	}()

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("✅ %d articles associated with vehicle %d!", len(data.Articles), data.VehicleID),
	})
}
