package controllers

import (
	"encoding/json"
	"fmt"
	"gocars-api/models"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Shop(c *gin.Context) {
	// vehicle_id := c.DefaultQuery("vehicle_id", "10538")
	category_id := c.DefaultQuery("category_id", "100260")

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

	filePath := fmt.Sprintf("data/%s.json", category_id)

	// Read local JSON file
	body, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read file: %s", filePath),
		})
		return
	}

	// url := fmt.Sprintf(
	// 	"https://tecdoc-catalog.p.rapidapi.com/articles/list/type-id/1/vehicle-id/%s/category-id/%s/lang-id/4",
	// 	vehicle_id,
	// 	category_id,
	// )

	// req, err := http.NewRequest("GET", url, nil)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
	// 	return
	// }

	// // Set headers
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("x-rapidapi-key", os.Getenv("X_RAPIDAPI_KEY"))
	// req.Header.Set("x-rapidapi-host", os.Getenv("X_RAPIDAPI_HOST"))

	// // Send request
	// resp, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending request"})
	// 	return
	// }
	// defer resp.Body.Close()

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response"})
	// 	return
	// }

	var vehicleArticles models.VehicleArticles
	if err := json.Unmarshal(body, &vehicleArticles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing JSON", "raw": string(body)})
		return
	}

	// Pagination logic
	start := (page - 1) * limit
	end := start + limit
	if start > len(vehicleArticles.Articles) {
		start = len(vehicleArticles.Articles)
	}
	if end > len(vehicleArticles.Articles) {
		end = len(vehicleArticles.Articles)
	}
	paginatedArticles := vehicleArticles.Articles[start:end]

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"limit":    limit,
		"total":    vehicleArticles.CountArticles,
		"articles": paginatedArticles,
	})
}
