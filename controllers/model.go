package controllers

import (
	"gocars-api/database"
	"gocars-api/models"
	"gocars-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetModels(c *gin.Context) {
	// vehicle_id := c.DefaultQuery("vehicle_id", "10538")
	manufacturerID := c.DefaultQuery("manufacturer_id", "100260")

	// 1️⃣ Try to load from database
	var dbModels []models.Model
	if err := database.DB.Where("manufacturer_id = ?", manufacturerID).Find(&dbModels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	// 2️⃣ If found in DB, return immediately
	if len(dbModels) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"total":  len(dbModels),
			"models": dbModels,
		})
		return
	}

	modelResponse, err := models.GetModelsFromRapidAPI(utils.Atoi(manufacturerID))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch models"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":  modelResponse.CountModels,
		"models": modelResponse.Models,
	})
}
