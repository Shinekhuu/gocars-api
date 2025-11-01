package controllers

import (
	"encoding/json"
	"fmt"
	"gocars-api/database"
	"gocars-api/models"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func GetManufacturers(c *gin.Context) {
	var manufacturers []models.Manufacturer

	// Fetch all manufacturers
	if err := database.DB.Find(&manufacturers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch manufacturers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"CountManufactures": len(manufacturers),
		"manufacturers":     manufacturers,
	})
}

func FillData(c *gin.Context) {
	// 1. Read JSON file
	file, err := os.Open("/home/ubuntu/project-go/gocars-api/data/manufactures.json")
	if err != nil {
		log.Fatal("failed to open JSON file:", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("failed to read JSON file:", err)
	}

	// 2. Parse JSON data
	var data models.ManufacturerData
	if err := json.Unmarshal(bytes, &data); err != nil {
		log.Fatal("failed to unmarshal JSON:", err)
	}

	fmt.Printf("Found %d manufacturers in JSON\n", data.CountManufactures)

	// 3. Insert or update records
	for _, m := range data.Manufacturers {
		err := database.DB.Where(models.Manufacturer{ManufacturerID: m.ManufacturerID}).Assign(m).FirstOrCreate(&m).Error
		if err != nil {
			log.Printf("❌ Failed to insert manufacturer %v: %v", m.ManufacturerName, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "✅ All manufacturers inserted/updated successfully!",
	})
}
