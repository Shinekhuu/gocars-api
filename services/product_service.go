package services

import (
	"gocars-api/models"
	"gocars-api/repositories"
)

type ProductService struct {
	repo *repositories.ProductRepository
}

func NewProductService(r *repositories.ProductRepository) *ProductService {
	return &ProductService{repo: r}
}

func (s *ProductService) GetProducts(filter repositories.ProductFilter) ([]models.Product, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	return s.repo.GetProducts(filter)
}
