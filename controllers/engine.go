package controllers

import (
	"gocars-api/models"
	"gocars-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetEngines(c *gin.Context) {
	ManufacturerID := utils.AtoiUint(c.DefaultQuery("manufacturer_id", "100260"))
	ModelID := utils.AtoiUint(c.DefaultQuery("model_id", "100260"))

	// 1️⃣ Try to load from database
	engineResponse, err := models.GetEnginesByModelId(ModelID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	// 2️⃣ If found in DB, return immediately
	if engineResponse.CountModelTypes > 0 {
		c.JSON(http.StatusOK, gin.H{
			"total":   engineResponse.CountModelTypes,
			"engines": engineResponse.Engines,
			"source":  "database",
		})
		return
	}

	engineResponse, err = models.GetEnginesFromRapidAPI(ManufacturerID, ModelID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch engines"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   engineResponse.CountModelTypes,
		"engines": engineResponse.Engines,
		"source":  "api",
	})
}
