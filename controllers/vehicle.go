package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gocars-api/database"
	"gocars-api/models"
	"gocars-api/scraper"
	"gocars-api/utils"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func FetchData(c *gin.Context) {
	plateNumber := c.DefaultQuery("plate_number", "")
	if plateNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required 'plate_number' parameter"})
		return
	}

	// ----------------- Step 1: Load or fetch Xyr ----------------- //
	xyr, err := getOrFetchXyr(plateNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ----------------- Step 2: Load engine or vehicle ----------------- //
	engine, vehicle, err := getOrFetchEngineVehicle(xyr, plateNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// ----------------- Step 3: Build response ----------------- //
	response := gin.H{
		"plate_number": plateNumber,
		"manufacturer": xyr.MarkName,
		"model":        fmt.Sprintf("%s %d", xyr.ModelName, xyr.BuildYear),
	}

	// Engine available
	if engine != nil && engine.VehicleID != 0 {
		response["vehicle_id"] = engine.VehicleID
		response["engine"] = fmt.Sprintf("%s %s %s", engine.FuelType, engine.TypeEngineName, engine.EngineCodes)
		c.JSON(http.StatusOK, response)
		return
	}

	// Vehicle available from scraper
	if vehicle != nil && vehicle.CarID != 0 {
		response["vehicle_id"] = vehicle.CarID
		if vehicle.MotorType != nil {
			response["engine"] = fmt.Sprintf("%s %s %s", utils.SafeString(vehicle.MotorType), vehicle.CarName, utils.SafeString(vehicle.MotorCode))
		} else {
			response["engine"] = vehicle.CarName
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Neither engine nor vehicle available
	response["vehicle_id"] = nil
	response["engine"] = "unknown"
	c.JSON(http.StatusOK, response)
}

// ---------------- Helper Functions ---------------- //

func getOrFetchXyr(plateNumber string) (models.Xyr, error) {
	var xyr models.Xyr
	err := database.DB.Where("plate_number = ?", plateNumber).First(&xyr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fetchAndStoreXyr(plateNumber)
	}
	return xyr, err
}

func getOrFetchEngineVehicle(xyr models.Xyr, plateNumber string) (*models.Engine, *models.Vehicle, error) {
	var xyrVehicle models.XyrVehicle

	// Try join first
	err := database.DB.Where("xyr_id = ?", xyr.ID).First(&xyrVehicle).Error
	if err == nil {
		var engine models.Engine
		err2 := database.DB.Where("vehicle_id = ?", xyrVehicle.VehicleID).First(&engine).Error
		if err2 == nil {
			return &engine, nil, nil
		}
	}

	// Either join missing or engine missing → fetch from scraper
	vehicle, err := fetchAndStoreVehicle(plateNumber)
	if err != nil {
		return nil, nil, err
	}

	// Ensure join exists
	if err := insertXyrVehicle(xyr.ID, vehicle.CarID); err != nil {
		return nil, vehicle, fmt.Errorf("failed to insert xyr engine join: %w", err)
	}

	return nil, vehicle, nil
}

func fetchAndStoreXyr(plateNumber string) (models.Xyr, error) {
	var xyr models.Xyr
	url := "https://xyp-api.smartcar.mn/xyp-api/v1/xyp/get-data-public"
	reqBody := map[string]interface{}{
		"serviceCode": "WS100401_getVehicleInfo",
		"customFields": map[string]string{
			"plateNumber": plateNumber,
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return xyr, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return xyr, fmt.Errorf("failed to call XYP API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return xyr, fmt.Errorf("failed to read response: %w", err)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return xyr, errors.New("API did not return JSON")
	}

	if err := json.Unmarshal(respBytes, &xyr); err != nil {
		return xyr, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	xyr.MarkName = strings.ToUpper(xyr.MarkName)
	xyr.ModelName = strings.ToUpper(xyr.ModelName)

	if err := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "plate_number"}},
		UpdateAll: true,
	}).Create(&xyr).Error; err != nil {
		return xyr, fmt.Errorf("failed to insert xyr: %w", err)
	}

	return xyr, nil
}

func fetchAndStoreVehicle(plateNumber string) (*models.Vehicle, error) {
	vehicle, err := scraper.FetchVehicleData(plateNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vehicle: %w", err)
	}

	if vehicle != nil && vehicle.ManuID != nil && vehicle.ModelID != nil {
		GetModelsWithEngines(*vehicle.ManuID, *vehicle.ModelID)
	}

	return vehicle, nil
}

func insertXyrVehicle(xyrID, vehicleID uint) error {
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "xyr_id"}, {Name: "vehicle_id"}},
		DoNothing: true,
	}).Create(&models.XyrVehicle{
		XyrID:     xyrID,
		VehicleID: vehicleID,
	}).Error
}

func GetModelsWithEngines(manufacturerID, modelID uint) {
	if modelResp, err := models.GetModelsByManufacturerId(manufacturerID); err != nil || modelResp.CountModels < 1 {
		models.GetModelsFromRapidAPI(manufacturerID)
	}
	if engineResp, err := models.GetEnginesByModelId(modelID); err != nil || engineResp.CountModelTypes < 1 {
		models.GetEnginesFromRapidAPI(manufacturerID, modelID)
	}
}
