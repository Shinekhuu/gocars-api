package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	scraperrepo "gocars-api/internal/vehicle/repository"
	vehiclesvc "gocars-api/internal/vehicle/service"

	"github.com/gin-gonic/gin"
)

type RequestBody struct {
	PlateNumber string `json:"plateNumber"`
}

type TodoHandler struct {
	manufacturerSvc *vehiclesvc.ManufacturerService
	modelSvc        *vehiclesvc.ModelService
	engineSvc       *vehiclesvc.EngineService
}

func NewTodoHandler(manufacturerSvc *vehiclesvc.ManufacturerService, modelSvc *vehiclesvc.ModelService, engineSvc *vehiclesvc.EngineService) *TodoHandler {
	return &TodoHandler{
		manufacturerSvc: manufacturerSvc,
		modelSvc:        modelSvc,
		engineSvc:       engineSvc,
	}
}

func (h *TodoHandler) GetXyrData(c *gin.Context) {
	var reqBody RequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.PlateNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plateNumber is required"})
		return
	}

	plate := reqBody.PlateNumber
	requestBody := map[string]interface{}{
		"serviceCode": "WS100401_getVehicleInfo",
		"customFields": map[string]string{
			"plateNumber": plate,
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request body", "details": err.Error()})
		return
	}

	resp, err := http.Post("https://xyp-api.smartcar.mn/xyp-api/v1/xyp/get-data-public", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call API", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read API response", "details": err.Error()})
		return
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API did not return JSON", "body": string(respBody)})
		return
	}

	type xyrRaw struct {
		MarkName    string `json:"markName"`
		ModelName   string `json:"modelName"`
		CabinNumber string `json:"cabinNumber"`
		PlateNumber string `json:"plateNumber"`
	}
	var xyr xyrRaw
	if err := json.Unmarshal(respBody, &xyr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal JSON", "details": err.Error()})
		return
	}

	xyr.MarkName = strings.ToUpper(xyr.MarkName)
	xyr.ModelName = strings.ToUpper(xyr.ModelName)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Manufacture not found", "details": err})
		return
	}

	dbModel, err := h.modelSvc.GetByName(dbManufacturer.ManufacturerID, xyr.ModelName, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Model not found", "details": err})
		return
	}

	dbEngines, err := h.engineSvc.GetByName(dbManufacturer.ManufacturerID, dbModel.ModelID, frame)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Engine not found", "details": err})
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
