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
