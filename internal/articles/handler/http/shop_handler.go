package handler

import (
	"net/http"

	repo "gocars-api/internal/articles/repository/postgresql"
	service "gocars-api/internal/articles/service"
	"gocars-api/internal/shared/utils"

	"github.com/gin-gonic/gin"
)

type ShopHandler struct {
	articleRepo  *repo.ArticleRepository
	fetchLogRepo *repo.APIFetchLogRepository
	articleSvc   *service.ArticleService
}

func NewShopHandler(
	articleRepo *repo.ArticleRepository,
	fetchLogRepo *repo.APIFetchLogRepository,
	articleSvc *service.ArticleService,
) *ShopHandler {
	return &ShopHandler{
		articleRepo:  articleRepo,
		fetchLogRepo: fetchLogRepo,
		articleSvc:   articleSvc,
	}
}

func (h *ShopHandler) Shop(c *gin.Context) {
	vehicleID := utils.AtoiUint(c.DefaultQuery("vehicle_id", "0"))
	categoryID := utils.AtoiUint(c.DefaultQuery("category_id", "100260"))
	page := utils.AtoiDefault(c.DefaultQuery("page", "1"), 1)
	limit := utils.AtoiDefault(c.DefaultQuery("limit", "40"), 40)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 40
	}

	articles, total, err := h.articleRepo.GetArticleItemsByVehicleIdAndCategoryId(vehicleID, categoryID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if total > 0 {
		if vehicleID != 0 {
			go func() { _ = h.fetchLogRepo.EnsureAPIFetchLog(vehicleID, categoryID) }()
			if h.articleSvc.ShouldRefetch(vehicleID, categoryID) {
				h.articleSvc.RefreshArticlesAsync(vehicleID, categoryID)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"page": page, "limit": limit, "total": total,
			"articles": articles, "api": "db",
		})
		return
	}

	if vehicleID == 0 {
		c.JSON(http.StatusOK, gin.H{
			"page": page, "limit": limit, "total": 0,
			"articles": []any{}, "api": "db",
		})
		return
	}

	apiData, err := h.articleSvc.GetArticleItemsFromRapidAPI(vehicleID, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	start := (page - 1) * limit
	end := start + limit
	if start > len(apiData.Articles) {
		start = len(apiData.Articles)
	}
	if end > len(apiData.Articles) {
		end = len(apiData.Articles)
	}

	c.JSON(http.StatusOK, gin.H{
		"page": page, "limit": limit,
		"total": len(apiData.Articles), "articles": apiData.Articles[start:end], "api": "api",
	})
}
