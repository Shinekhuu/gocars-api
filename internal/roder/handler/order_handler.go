package handler

import (
	"net/http"

	handlrdto "gocars-api/internal/roder/handler/dto"
	"gocars-api/internal/roder/service"
	"gocars-api/internal/shared/utils"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := utils.Atoi(c.Param("id"))

	order, err := h.svc.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetOrderPDF(c *gin.Context) {
	id := c.Param("id")
	orderID := utils.Atoi(id)

	order, err := h.svc.GetOrderWithFitmentDataByID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	pdfBytes, err := service.GenerateOrderPDF(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pdf generation failed"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\"quote-"+id+".pdf\"")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var input handlrdto.CreateOrderInputs
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "items required"})
		return
	}

	order, invoice, err := h.svc.CreateTransaction(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"order":   order,
		"invoice": invoice,
	})
}
