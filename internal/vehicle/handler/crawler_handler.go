package handler

import (
	"net/http"

	vehiclesvc "gocars-api/internal/vehicle/service"

	"github.com/gin-gonic/gin"
)

type CrawlerHandler struct {
	svc *vehiclesvc.CrawlerService
}

func NewCrawlerHandler(svc *vehiclesvc.CrawlerService) *CrawlerHandler {
	return &CrawlerHandler{svc: svc}
}

func (h *CrawlerHandler) Garage(c *gin.Context) {
	plate := c.DefaultQuery("plate", "")
	if plate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required 'plate' parameter"})
		return
	}

	body, err := h.svc.FetchGarageByPlate(plate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"body": body})
}
