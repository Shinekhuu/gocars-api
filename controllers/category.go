package controllers

import (
	"gocars-api/models"

	"github.com/gin-gonic/gin"
)

func FillCategories(c *gin.Context) {
	file := "/home/ubuntu/project-go/gocars-api/data/categories.json"

	if err := models.SeedCategories(file); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Categories imported successfully!"})
}
