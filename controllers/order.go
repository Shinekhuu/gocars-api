package controllers

import (
	"gocars-api/database"
	"gocars-api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Inputs: list of article IDs
type OrderItemsInput struct {
	ArticleIDs []uint `form:"article_ids[]" binding:"required"`
}

func CreateOrder(c *gin.Context) {
	// 1. Get user email from context
	email, _ := c.Get("email")

	// Pagination query params
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "40")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 40
	}

	// 2. Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 3. Parse form input (x-www-form-urlencoded)
	var input OrderItemsInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.ArticleIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "article_ids cannot be empty"})
		return
	}

	// 4. Create order transaction
	createdOrder, createdInvoice, err := models.CreateTransaction(user.ID, input.ArticleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Pagination logic
	offset := (page - 1) * limit

	// Load order items with pagination
	var orderItems []models.OrderItem
	if err := database.DB.
		Where("order_id = ?", createdOrder.ID).
		Preload("Article").
		Offset(offset).
		Limit(limit).
		Find(&orderItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 6. Build clean response
	type OrderItemResponse struct {
		ArticleID uint    `json:"article_id"`
		Name      string  `json:"name"`
		Price     float64 `json:"price"`
		Quantity  int     `json:"quantity"`
	}

	items := make([]OrderItemResponse, 0)
	for _, oi := range orderItems {
		if oi.Article != nil { // safety check
			items = append(items, OrderItemResponse{
				ArticleID: oi.ArticleID,
				Name:      oi.Article.ArticleProductName,
				Price:     oi.Price,
				Quantity:  oi.Quantity,
			})
		}
	}

	// Return JSON
	c.JSON(http.StatusOK, gin.H{
		"order_id":       createdOrder.ID,
		"invoice_number": createdInvoice.InvoiceNumber,
		"total":          len(items),
		"order_items":    items,
		"paid":           createdInvoice.Paid,
		"page":           page,
		"limit":          limit,
	})
}
