package controllers

import (
	"gocars-api/database"
	"gocars-api/models"
	"gocars-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetEngines(c *gin.Context) {
	ManufacturerID := c.DefaultQuery("manufacturer_id", "100260")
	ModelID := c.DefaultQuery("model_id", "100260")

	// 1️⃣ Try to load from database
	var dbEngines []models.Engine
	if err := database.DB.Where("model_id = ?", ModelID).Find(&dbEngines).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	// 2️⃣ If found in DB, return immediately
	if len(dbEngines) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"total":   len(dbEngines),
			"engines": dbEngines,
		})
		return
	}

	engineResponse, err := models.GetEnginesFromRapidAPI(utils.Atoi(ManufacturerID), utils.Atoi(ModelID))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch engines"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   engineResponse.CountModelTypes,
		"engines": engineResponse.Engines,
	})
}
