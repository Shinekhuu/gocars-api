package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CrawlerHandler struct{}

func NewCrawlerHandler() *CrawlerHandler {
	return &CrawlerHandler{}
}

func (h *CrawlerHandler) Garage(c *gin.Context) {
	plate := c.DefaultQuery("plate", "")
	if plate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required 'plate' parameter"})
		return
	}

	base := os.Getenv("GARAGE_HOST")
	startURL := base + "platenew?platenumber=" + plate
	if _, err := crawlPage(startURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	endURL := base + "plate?platenumber=" + plate
	endBody, err := crawlPage(endURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"body": endBody})
}

func crawlPage(pageURL string) (string, error) {
	res, err := http.Get(pageURL)
	if err != nil {
		zap.L().Error("Error fetching page", zap.String("url", pageURL), zap.Error(err))
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code %d", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

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
