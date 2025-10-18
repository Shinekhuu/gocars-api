package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

var visited = make(map[string]bool)

// Crawl fetches the page content
func crawl(pageURL string, depth int) (string, error) {
	if depth <= 0 {
		return "", nil
	}

	if visited[pageURL] {
		return "", nil
	}
	visited[pageURL] = true

	fmt.Println("Crawling:", pageURL)

	res, err := http.Get(pageURL)
	if err != nil {
		log.Println("Error fetching page:", err)
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Println("Non-OK HTTP status:", res.StatusCode)
		return "", fmt.Errorf("status code %d", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return "", err
	}

	return string(bodyBytes), nil
}

// Resolve relative URL to absolute
func resolveURL(base, href string) string {
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(u).String()
}

// Gin handler
func Garage(c *gin.Context) {
	plate := c.DefaultQuery("plate", "")

	if plate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required 'plate' parameter"})
		return
	}

	startURL := "https://apiweb.garage.mn/api/platenew?platenumber=" + plate
	_, err := crawl(startURL, 1) // Crawl 1 level deep
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	endUrl := "https://apiweb.garage.mn/api/plate?platenumber=" + plate

	endBody, err := crawl(endUrl, 1) // Crawl 2 level deep
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"body": endBody,
	})
}
