package controllers

import (
	"gocars-api/dto"
	"gocars-api/services"
	"gocars-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetOrder(c *gin.Context) {
	id := c.Param("id")

	// optional uint convert
	// orderID, _ := strconv.ParseUint(id, 10, 64)

	orderID := utils.Atoi(id)

	order, err := services.GetOrderByID(orderID)

	if err != nil {
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"error": "Order not found",
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		order,
	)
}

func GetOrderPDF(c *gin.Context) {

	id := c.Param("id")
	orderID := utils.Atoi(id)

	// data
	order, err :=
		services.GetOrderWithFitmentDataByID(
			orderID,
		)

	if err != nil {
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"error": "order not found",
			},
		)
		return
	}

	// pdf
	pdfBytes, err :=
		services.GenerateOrderPDF(order)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "pdf generation failed",
			},
		)
		return
	}

	filename :=
		"quote-" + id + ".pdf"

	c.Header(
		"Content-Type",
		"application/pdf",
	)

	c.Header(
		"Content-Disposition",
		`attachment; filename="`+
			filename+
			`"`,
	)

	c.Data(
		http.StatusOK,
		"application/pdf",
		pdfBytes,
	)
}

func CreateOrder(c *gin.Context) {
	var input dto.CreateOrderInputs

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		return
	}

	if len(input.Items) == 0 {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "items required"},
		)
		return
	}

	order, invoice, err := services.CreateTransaction(input)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"order":   order,
			"invoice": invoice,
		},
	)
}
