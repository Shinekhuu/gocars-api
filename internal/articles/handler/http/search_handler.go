package handler

import (
	"net/http"
	"strconv"

	articles "gocars-api/internal/articles/repository/postgresql/model"
	repo "gocars-api/internal/articles/repository/postgresql"
	service "gocars-api/internal/articles/service"
	"gocars-api/internal/search/meili"
	"gocars-api/internal/shared/utils"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	articleRepo *repo.ArticleRepository
	articleSvc  *service.ArticleService
}

func NewSearchHandler(articleRepo *repo.ArticleRepository, articleSvc *service.ArticleService) *SearchHandler {
	return &SearchHandler{
		articleRepo: articleRepo,
		articleSvc:  articleSvc,
	}
}

func (h *SearchHandler) Search(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	vehicleIDStr := c.DefaultQuery("vehicle_id", "")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "40"))
	if err != nil || limit < 1 {
		limit = 40
	}

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required 'query' parameter"})
		return
	}

	// Try Meilisearch first
	if meili.Default != nil {
		result, err := meili.Default.Search(query, nil, page, limit)
		if err == nil && len(result.Hits) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"page": page, "limit": limit, "total": result.Total,
				"articles": result.Hits, "source": "meili",
			})
			return
		}
	}

	// Fall back to DB
	filter := articles.ProductFilter{
		Search:    &query,
		VehicleID: utils.StringToUintPtr(vehicleIDStr),
		Page:      page,
		Limit:     limit,
	}
	products, total, err := h.articleRepo.SearchProducts(filter)
	if err == nil && total > 0 {
		c.JSON(http.StatusOK, gin.H{
			"page": page, "limit": limit, "total": total,
			"articles": products, "source": "db",
		})
		return
	}

	// Fall back to RapidAPI
	apiArticles, err := h.articleSvc.GetByOemFromRapidAPI(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	start := (page - 1) * limit
	end := start + limit
	if start > len(apiArticles) {
		start = len(apiArticles)
	}
	if end > len(apiArticles) {
		end = len(apiArticles)
	}

	c.JSON(http.StatusOK, gin.H{
		"page": page, "limit": limit, "total": len(apiArticles),
		"articles": apiArticles[start:end], "source": "api",
	})
}
