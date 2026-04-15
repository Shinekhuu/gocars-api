package controllers

import (
	"gocars-api/repositories"
	"gocars-api/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler(s *services.ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	search := c.Query("search")
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	categoryStr := c.Query("category_id")
	var categoryPtr *uint
	if categoryStr != "" {
		if v, err := strconv.Atoi(categoryStr); err == nil {
			tmp := uint(v)
			categoryPtr = &tmp
		}
	}

	filter := repositories.ProductFilter{
		Search:     searchPtr,
		CategoryID: categoryPtr,
		Page:       page,
		Limit:      limit,
	}

	data, total, err := h.service.GetProducts(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  data,
		"total": total,
	})
}
