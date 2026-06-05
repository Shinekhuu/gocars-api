package handler

import (
	"net/http"

	vehiclesvc "gocars-api/internal/vehicle/service"

	"github.com/gin-gonic/gin"
)

type VinHandler struct {
	svc *vehiclesvc.VinService
}

func NewVinHandler(svc *vehiclesvc.VinService) *VinHandler {
	return &VinHandler{svc: svc}
}

func (h *VinHandler) Decode(c *gin.Context) {
	vin := c.DefaultQuery("vin", "MHU382076138")
	info := h.svc.Decode(vin)
	c.JSON(http.StatusOK, info)
}
