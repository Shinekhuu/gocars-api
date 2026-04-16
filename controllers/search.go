package controllers

import (
	"gocars-api/models"
	"gocars-api/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Search(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	vehicle_id := c.DefaultQuery("vehicle_id", "")

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

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required 'query' parameter"})
		return
	}

	filter := models.ProductFilter{
		Search:    &query,
		VehicleID: utils.StringToUintPtr(vehicle_id),
		Page:      page,
		Limit:     limit,
	}

	products, total, err := models.GetProducts(filter)

	if err == nil && total > 0 {
		c.JSON(http.StatusOK, gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"articles": products,
		})
		return
	}

	// Otherwise, fetch from API
	var apiArticles []models.ArticleItem
	apiArticles, err = models.GetArticleItemsByOemFromRapidAPI(query)
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
