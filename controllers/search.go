package controllers

import (
	"encoding/json"
	"gocars-api/models"
	"io"
	"net/http"
	"os"
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

	url := "https://tecdoc-catalog.p.rapidapi.com/articles-oem/search-by-article-oem-no/lang-id/4/article-oem-no/" + oem

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-key", os.Getenv("X_RAPIDAPI_KEY"))
	req.Header.Set("x-rapidapi-host", os.Getenv("X_RAPIDAPI_HOST"))

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending request"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response"})
		return
	}

	var articles models.ArticleList
	if err := json.Unmarshal(body, &articles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing JSON", "raw": string(body)})
		return
	}

	// Pagination logic
	start := (page - 1) * limit
	end := start + limit
	if start > len(articles) {
		start = len(articles)
	}
	if end > len(articles) {
		end = len(articles)
	}
	paginatedArticles := articles[start:end]

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"limit":    limit,
		"total":    len(articles),
		"articles": paginatedArticles,
	})
}
