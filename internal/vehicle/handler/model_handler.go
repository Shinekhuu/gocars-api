package handler

import (
	"net/http"

	"gocars-api/internal/shared/utils"
	vehiclesvc "gocars-api/internal/vehicle/service"

	"github.com/gin-gonic/gin"
)

type ModelHandler struct {
	svc *vehiclesvc.ModelService
}

func NewModelHandler(svc *vehiclesvc.ModelService) *ModelHandler {
	return &ModelHandler{svc: svc}
}

func (h *ModelHandler) GetModels(c *gin.Context) {
	manufacturerID := utils.AtoiUint(c.DefaultQuery("manufacturer_id", "100260"))

	modelResponse, err := h.svc.GetOrFetchByManufacturerID(manufacturerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch models"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":  modelResponse.CountModels,
		"models": modelResponse.Models,
	})
}
