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
	manufacturerID := utils.AtoiUint(c.DefaultQuery("manufacturer_id", "0"))
	if manufacturerID == 0 {
		manufacturerID = 100260
	}
	modelID := utils.AtoiUint(c.DefaultQuery("model_id", "0"))
	if modelID == 0 {
		modelID = 100260
	}

	engineResponse, err := h.svc.GetByModelID(modelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	if engineResponse.CountModelTypes > 0 {
		c.JSON(http.StatusOK, gin.H{
			"total":   engineResponse.CountModelTypes,
			"engines": engineResponse.Engines,
			"source":  "database",
		})
		return
	}

	engineResponse, err = h.svc.FetchFromAPI(manufacturerID, modelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch engines"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   engineResponse.CountModelTypes,
		"engines": engineResponse.Engines,
		"source":  "api",
	})
}
