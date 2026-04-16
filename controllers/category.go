package controllers

import (
	"context"
	"gocars-api/database"
	"gocars-api/models"
	"log"

	"github.com/gin-gonic/gin"
)

func FillCategories(c *gin.Context) {
	jsonPath := "/home/api/data/categories.json"

	go func() {
		if err := models.SeedCategoriesFromFile(context.Background(), database.DB, jsonPath); err != nil {
			log.Printf("Error seeding categories: %v\n", err)
		} else {
			log.Println("Categories imported successfully")
		}
	}()

	c.JSON(200, gin.H{"message": "Seeding started in background"})
}

func SyncCategoryMn(c *gin.Context) {
	categoryMap, err := models.LoadCategoryMap("/home/ubuntu/project-go/gocars-api/data/categories_mn.json")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = models.UpdateCategoryNamesBatch(categoryMap)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Category names updated successfully"})
}
