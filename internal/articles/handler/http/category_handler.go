package handler

import (
	"context"
	"net/http"

	articles "gocars-api/internal/articles/repository/postgresql/model"
	repo "gocars-api/internal/articles/repository/postgresql"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CategoryHandler struct {
	categoryRepo *repo.CategoryRepository
}

func NewCategoryHandler(categoryRepo *repo.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

func (h *CategoryHandler) FillCategories(c *gin.Context) {
	jsonPath := "/home/api/data/categories.json"
	go func() {
		if err := h.categoryRepo.SeedFromFile(context.Background(), jsonPath); err != nil {
			zap.L().Error("error seeding categories", zap.Error(err))
		} else {
			zap.L().Info("categories imported successfully")
		}
	}()
	c.JSON(http.StatusOK, gin.H{"message": "seeding started in background"})
}

func (h *CategoryHandler) SyncCategoryMn(c *gin.Context) {
	categoryMap, err := articles.LoadCategoryMap("/home/ubuntu/project-go/gocars-api/data/categories_mn.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.categoryRepo.UpdateCategoryNamesBatch(categoryMap); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category names updated successfully"})
}
