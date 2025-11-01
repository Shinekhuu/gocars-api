package controllers

import (
	"bytes"
	"encoding/json"
	"gocars-api/database"
	"gocars-api/models"
	"gocars-api/scraper"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

// RequestBody represents your incoming POST request
type RequestBody struct {
	PlateNumber string `json:"plateNumber" binding:"required"`
}

func GetXyrData(c *gin.Context) {
	var reqBody RequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plateNumber is required"})
		return
	}

	plate := reqBody.PlateNumber

	// 1. Fetch data from the API
	url := "https://xyp-api.smartcar.mn/xyp-api/v1/xyp/get-data-public"
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

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
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

	// 2. Parse JSON into Xyr struct
	var xyr models.Xyr
	if err := json.Unmarshal(respBody, &xyr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal JSON", "details": err.Error()})
		return
	}

	xyr.MarkName = strings.ToUpper(xyr.MarkName)
	xyr.ModelName = strings.ToUpper(xyr.ModelName)

	// 3. Insert or update on conflict
	if result := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "plate_number"}},
		UpdateAll: true,
	}).Create(&xyr); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert record", "details": result.Error.Error()})
		return
	}

	// 4. Fetch Toyota/Lexus VIN info using goquery scraper
	var market, year, makeVal, model, frame string

	if xyr.MarkName == "TOYOTA" || xyr.MarkName == "LEXUS" {
		market, year, makeVal, model, frame, err = scraper.FetchVehicleInfo(xyr.CabinNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch vehicle info",
				"details": err.Error(),
			})
			return
		}
	}

	// 5. Get Manufacturer ID
	dbManufacturer, err := models.GetManufacturerByName(xyr.MarkName)
	if err != nil {
		// handle not found
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Manufacture not found",
			"details": err,
		})
		return
	}

	// 6. Get Model ID
	dbModel, err := models.GetModelByName(dbManufacturer.ManufacturerID, xyr.ModelName, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Model not found",
			"details": err,
		})
		return
	}

	// 7. Get Vehicle ID
	dbEngine, err := models.GetEngineByName(dbManufacturer.ManufacturerID, dbModel.ModelID, frame)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Engine not found",
			"details": err,
		})
		return
	}

	// 5. Return JSON response
	c.JSON(http.StatusOK, gin.H{
		"plateNumber": xyr.PlateNumber,
		"cabinNumber": xyr.CabinNumber,
		"market":      market,
		"year":        year,
		"make":        makeVal,
		"model":       model,
		"frame":       frame,
		"engine":      dbEngine,
	})
}
