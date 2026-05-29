package handler

import (
	"fmt"
	"net/http"

	"gocars-api/internal/shared/utils"
	vehiclesvc "gocars-api/internal/vehicle/service"

	"github.com/gin-gonic/gin"
)

type VehicleHandler struct {
	svc       *vehiclesvc.VehicleService
	engineSvc *vehiclesvc.EngineService
	modelSvc  *vehiclesvc.ModelService
}

func NewVehicleHandler(svc *vehiclesvc.VehicleService, engineSvc *vehiclesvc.EngineService, modelSvc *vehiclesvc.ModelService) *VehicleHandler {
	return &VehicleHandler{svc: svc, engineSvc: engineSvc, modelSvc: modelSvc}
}

func (h *VehicleHandler) FetchData(c *gin.Context) {
	plateNumber := c.DefaultQuery("plate_number", "")
	if plateNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required 'plate_number' parameter"})
		return
	}

	xyr, err := h.svc.GetOrFetchXyr(plateNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	engine, veh, err := h.svc.GetOrFetchEngineVehicle(xyr, plateNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"plate_number": plateNumber,
		"manufacturer": xyr.MarkName,
		"model":        fmt.Sprintf("%s %d", xyr.ModelName, xyr.BuildYear),
	}

	if engine != nil && engine.VehicleID != 0 {
		response["vehicle_id"] = engine.VehicleID
		response["engine"] = fmt.Sprintf("%s %s %s", engine.FuelType, engine.TypeEngineName, engine.EngineCodes)
		c.JSON(http.StatusOK, response)
		return
	}

	if veh != nil && veh.CarID != 0 {
		response["vehicle_id"] = veh.CarID
		if veh.MotorType != nil {
			response["engine"] = fmt.Sprintf("%s %s %s", utils.SafeString(veh.MotorType), veh.CarName, utils.SafeString(veh.MotorCode))
		} else {
			response["engine"] = veh.CarName
		}
		c.JSON(http.StatusOK, response)
		return
	}

	response["vehicle_id"] = nil
	response["engine"] = "unknown"
	c.JSON(http.StatusOK, response)
}
