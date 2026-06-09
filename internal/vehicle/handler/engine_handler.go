package handler

import (
	"net/http"

	"gocars-api/internal/shared/utils"
	vehiclesvc "gocars-api/internal/vehicle/service"

	"github.com/gin-gonic/gin"
)

type EngineHandler struct {
	svc *vehiclesvc.EngineService
}

func NewEngineHandler(svc *vehiclesvc.EngineService) *EngineHandler {
	return &EngineHandler{svc: svc}
}

func (h *EngineHandler) GetEngines(c *gin.Context) {
	manufacturerID := utils.AtoiUint(c.DefaultQuery("manufacturer_id", "100260"))
	modelID := utils.AtoiUint(c.DefaultQuery("model_id", "100260"))

	engineResponse, err := h.svc.GetOrFetchByModelID(manufacturerID, modelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch engines"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   engineResponse.CountModelTypes,
		"engines": engineResponse.Engines,
	})
}
