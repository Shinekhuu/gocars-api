package handler

import (
	"net/http"

	vehiclesvc "gocars-api/internal/vehicle/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ManufacturerHandler struct {
	svc *vehiclesvc.ManufacturerService
}

func NewManufacturerHandler(svc *vehiclesvc.ManufacturerService) *ManufacturerHandler {
	return &ManufacturerHandler{svc: svc}
}

func (h *ManufacturerHandler) GetManufacturers(c *gin.Context) {
	manufacturers, err := h.svc.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch manufacturers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"CountManufactures": len(manufacturers),
		"manufacturers":     manufacturers,
	})
}

func (h *ManufacturerHandler) FillData(c *gin.Context) {
	count, err := h.svc.FillFromFile("/home/api/data/manufactures.json")
	if err != nil {
		zap.L().Error("FillData failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All manufacturers inserted/updated successfully!",
		"count":   count,
	})
}
