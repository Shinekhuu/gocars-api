package handler

import (
	"net/http"
	"strconv"

	repository "gocars-api/internal/articles/repository/postgresql"
	"gocars-api/internal/articles/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(s *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: s}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 {
		limit = 20
	}

	search := c.DefaultQuery("search", "")
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	var categoryPtr *uint
	if categoryStr := c.DefaultQuery("category_id", ""); categoryStr != "" {
		if v, err := strconv.Atoi(categoryStr); err == nil {
			tmp := uint(v)
			categoryPtr = &tmp
		}
	}

	filter := repository.ProductFilter{
		Search:     searchPtr,
		CategoryID: categoryPtr,
		Page:       page,
		Limit:      limit,
	}

	data, total, err := h.svc.GetProducts(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  data,
		"total": total,
	})
}
