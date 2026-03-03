package controllers

import (
	"gocars-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetResponse(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"responses": services.GetResponseOpenAi(),
	})
}
