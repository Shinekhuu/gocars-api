package handler

import (
	"net/http"

	"gocars-api/internal/articles/service"
	"gocars-api/internal/shared/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ArticleHandler struct {
	svc *service.ArticleService
}

func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

func (h *ArticleHandler) Article(c *gin.Context) {
	id := utils.Atoi(c.DefaultQuery("id", ""))
	articleID := utils.Atoi(c.DefaultQuery("article_id", ""))
	page := utils.AtoiDefault(c.DefaultQuery("page", ""), 1)
	limit := utils.AtoiDefault(c.DefaultQuery("limit", ""), 20)

	res, err := h.svc.GetArticleDetail(id, articleID, page, limit)
	if err != nil {
		zap.L().Error("error fetching article detail", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
