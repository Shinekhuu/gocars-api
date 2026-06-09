package handler

import (
	"net/http"

	scraperrepo "gocars-api/internal/vehicle/repository"
	vehiclesvc "gocars-api/internal/vehicle/service"

	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	vehicleSvc      *vehiclesvc.VehicleService
	manufacturerSvc *vehiclesvc.ManufacturerService
	modelSvc        *vehiclesvc.ModelService
	engineSvc       *vehiclesvc.EngineService
}

func NewTodoHandler(
	vehicleSvc *vehiclesvc.VehicleService,
	manufacturerSvc *vehiclesvc.ManufacturerService,
	modelSvc *vehiclesvc.ModelService,
	engineSvc *vehiclesvc.EngineService,
) *TodoHandler {
	return &TodoHandler{
		vehicleSvc:      vehicleSvc,
		manufacturerSvc: manufacturerSvc,
		modelSvc:        modelSvc,
		engineSvc:       engineSvc,
	}
}

func (h *TodoHandler) GetXyrData(c *gin.Context) {
	var reqBody struct {
		PlateNumber string `json:"plateNumber"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.PlateNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plateNumber is required"})
		return
	}

	xyr, err := h.vehicleSvc.FetchXyr(reqBody.PlateNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vehicle info", "details": err.Error()})
		return
	}

	var market, year, makeVal, model, frame string
	if xyr.MarkName == "TOYOTA" || xyr.MarkName == "LEXUS" {
		market, year, makeVal, model, frame, err = scraperrepo.FetchVehicleInfo(xyr.CabinNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vehicle info", "details": err.Error()})
			return
		}
	}

	dbManufacturer, err := h.manufacturerSvc.GetByName(xyr.MarkName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Manufacturer not found", "details": err.Error()})
		return
	}

	dbModel, err := h.modelSvc.GetByName(dbManufacturer.ManufacturerID, xyr.ModelName, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Model not found", "details": err.Error()})
		return
	}

	dbEngines, err := h.engineSvc.GetByName(dbManufacturer.ManufacturerID, dbModel.ModelID, frame)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Engine not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plateNumber": xyr.PlateNumber,
		"cabinNumber": xyr.CabinNumber,
		"market":      market,
		"year":        year,
		"make":        makeVal,
		"model":       model,
		"frame":       frame,
		"engine":      dbEngines,
	})
}
