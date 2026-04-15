package controllers

import (
	"gocars-api/services"
	"gocars-api/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Article(c *gin.Context) {
	id := utils.Atoi(c.DefaultQuery("id", ""))
	articleID := utils.Atoi(c.DefaultQuery("article_id", ""))

	page := utils.AtoiDefault(c.DefaultQuery("page", ""), 1)
	limit := utils.AtoiDefault(c.DefaultQuery("limit", ""), 20)

	log.Printf("Received request for article detail. ID: %d, ArticleID: %d, Page: %d, Limit: %d", id, articleID, page, limit)

	res, err := services.GetArticleDetail(id, articleID, page, limit)
	if err != nil {
		log.Printf("Error fetching article detail: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// func ArticleOld(c *gin.Context) {
// 	// ==========================
// 	// READ PARAMS
// 	// ==========================
// 	idStr := c.DefaultQuery("id", "")
// 	articleIDStr := c.DefaultQuery("article_id", "")

// 	ID := utils.Atoi(idStr)
// 	articleID := utils.Atoi(articleIDStr)

// 	// Pagination params
// 	pageStr := c.DefaultQuery("page", "1")
// 	limitStr := c.DefaultQuery("limit", "20")

// 	page, err := strconv.Atoi(pageStr)
// 	if err != nil || page < 1 {
// 		page = 1
// 	}

// 	limit, err := strconv.Atoi(limitStr)
// 	if err != nil || limit < 1 {
// 		limit = 20
// 	}
// 	offset := (page - 1) * limit

// 	// ==========================
// 	// 1. CHECK DB FIRST
// 	// ==========================
// 	var articleDetail models.ArticleItem
// 	result := database.DB.
// 		Preload("AllSpecifications").
// 		Preload("AllOems.Oem").
// 		Where("id = ?", ID).
// 		Or("article_id", articleID).
// 		First(&articleDetail)

// 	if *articleDetail.ArticleID == 0 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Article not found"})
// 		return
// 	}

// 	// ==========================
// 	// DB HIT → RETURN PAGINATED FROM DB
// 	// ==========================
// 	if result.Error == nil && articleDetail.IsFetched {
// 		// Count total engines
// 		var total int64
// 		database.DB.Model(&models.ArticleVehicles{}).
// 			Where("article_item_id = ?", ID).
// 			Count(&total)

// 		// Paginate engines
// 		var engines []models.Engine
// 		database.DB.
// 			Table("article_vehicles av").
// 			Select("e.*").
// 			Joins("LEFT JOIN engines e ON av.vehicle_id = e.vehicle_id").
// 			Where("av.article_item_id = ?", ID).
// 			Offset(offset).
// 			Limit(limit).
// 			Find(&engines)

// 		// Build OEM response
// 		oems := make([]gin.H, len(articleDetail.AllOems))
// 		for i, link := range articleDetail.AllOems {
// 			oems[i] = gin.H{
// 				"brand":     link.Oem.Brand,
// 				"displayNo": link.Oem.DisplayNo,
// 			}
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"article": gin.H{
// 				"id":                   articleDetail.ID,
// 				"articleId":            articleDetail.ArticleID,
// 				"articleSearchNo":      articleDetail.ArticleSearchNo,
// 				"articleNo":            articleDetail.ArticleNo,
// 				"articleProductName":   articleDetail.ArticleProductName,
// 				"productId":            articleDetail.ProductID,
// 				"manufacturerId":       articleDetail.ManufacturerID,
// 				"manufacturerName":     articleDetail.ManufacturerName,
// 				"supplierId":           articleDetail.SupplierID,
// 				"supplierName":         articleDetail.SupplierName,
// 				"articleMediaType":     articleDetail.ArticleMediaType,
// 				"articleMediaFileName": articleDetail.ArticleMediaFileName,
// 				"s3image":              articleDetail.S3Image,
// 			},
// 			"oems":              oems,
// 			"allSpecifications": articleDetail.AllSpecifications,
// 			"compatibleCars": gin.H{
// 				"page":  page,
// 				"limit": limit,
// 				"total": total,
// 				"data":  engines,
// 			},
// 		})
// 		return
// 	}

// 	// ==========================
// 	// NO DB → FETCH FROM API
// 	// ==========================
// 	if articleID == 0 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
// 		return
// 	}

// 	articlePtr, err := models.GetArticleCompleteDetailFromRapidAPI(articleID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	articleDetail = *articlePtr

// 	// ===== PAGINATE API slice =====
// 	totalAPI := len(articleDetail.CompatibleCarsResponse)

// 	start := offset
// 	end := offset + limit

// 	if start > totalAPI {
// 		start = totalAPI
// 	}
// 	if end > totalAPI {
// 		end = totalAPI
// 	}

// 	pagedAPI := articleDetail.CompatibleCarsResponse[start:end]

// 	// Build OEM response (from API)
// 	oems := make([]gin.H, len(articleDetail.OemResponses))
// 	for i, oemResp := range articleDetail.OemResponses {
// 		oems[i] = gin.H{
// 			"brand":     oemResp.Brand,
// 			"displayNo": oemResp.DisplayNo,
// 		}
// 	}

// 	// ==========================
// 	// RETURN API RESULT
// 	// ==========================
// 	c.JSON(http.StatusOK, gin.H{
// 		"article": gin.H{
// 			"articleId":            articleDetail.ArticleID,
// 			"articleSearchNo":      articleDetail.ArticleSearchNo,
// 			"articleNo":            articleDetail.ArticleNo,
// 			"articleProductName":   articleDetail.ArticleProductName,
// 			"productId":            articleDetail.ProductID,
// 			"manufacturerId":       articleDetail.ManufacturerID,
// 			"manufacturerName":     articleDetail.ManufacturerName,
// 			"supplierId":           articleDetail.SupplierID,
// 			"supplierName":         articleDetail.SupplierName,
// 			"articleMediaType":     articleDetail.ArticleMediaType,
// 			"articleMediaFileName": articleDetail.ArticleMediaFileName,
// 			"s3image":              articleDetail.S3Image,
// 		},
// 		"oems":              oems,
// 		"allSpecifications": articleDetail.AllSpecifications,
// 		"compatibleCars": gin.H{
// 			"page":  page,
// 			"limit": limit,
// 			"total": totalAPI,
// 			"data":  pagedAPI,
// 		},
// 	})
// }
