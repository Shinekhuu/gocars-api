package controllers

import (
	"gocars-api/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetResponse(c *gin.Context) {
	oem := c.DefaultQuery("oem", "")
	name := c.DefaultQuery("name", "")
	brand := c.DefaultQuery("brand", "")
	modelOrSpecifications := c.DefaultQuery("specifications", "")

	car, err := services.GetResponseOpenAi(oem, name, brand, modelOrSpecifications)
	if err != nil {
		log.Println(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"responses": car,
	})
}

func GetMapper(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	oem := c.DefaultQuery("oem", "")
	brand := c.DefaultQuery("brand", "")
	modelStr := c.DefaultQuery("input", "")

	result, err := services.MapWithAI(name, oem, modelStr, brand)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get mapping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mapping": result,
	})
}
