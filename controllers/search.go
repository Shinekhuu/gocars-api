package controllers

import (
	"gocars-api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SearchOEM(c *gin.Context) {
	oem := c.DefaultQuery("oem", "")

	// Pagination query params
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "40")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 40
	}

	if oem == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required 'oem' parameter"})
		return
	}

	// Try reading from DB first
	articles, total, err := models.GetArticleItemsByOem(oem, page, limit)
	if err == nil && total > 0 {
		c.JSON(http.StatusOK, gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"articles": articles,
		})
		return
	}

	// Otherwise, fetch from API
	var apiArticles []models.ArticleItem
	apiArticles, err = models.GetArticleItemsByOemFromRapidAPI(oem)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply pagination manually (API returns full list)
	start := (page - 1) * limit
	end := start + limit
	if start > len(apiArticles) {
		start = len(apiArticles)
	}
	if end > len(apiArticles) {
		end = len(apiArticles)
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"limit":    limit,
		"total":    len(apiArticles),
		"articles": apiArticles[start:end],
	})
}
