package service

import (
	articles "gocars-api/internal/articles/repository/postgresql/model"
	repo "gocars-api/internal/articles/repository/postgresql"
)

type ProductService struct {
	repo *repo.ProductRepository
}

func NewProductService(r *repo.ProductRepository) *ProductService {
	return &ProductService{repo: r}
}

func (s *ProductService) GetProducts(filter repo.ProductFilter) ([]articles.Product, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	return s.repo.GetProducts(filter)
}
