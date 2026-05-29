package repository

import (
	roder "gocars-api/internal/roder/repository/postgresql/model"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetOrderForSuccess(id int) (*roder.Order, error) {
	var order roder.Order

	err := r.db.
		Preload("Invoice").
		Preload("OrderItems").
		Preload("OrderItems.ArticleItem").
		Preload("OrderItems.ArticleItem.AllSpecifications").
		First(&order, id).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) GetOrderForPDF(id int) (*roder.Order, error) {
	var order roder.Order

	err := r.db.
		Preload("Invoice").
		Preload("OrderItems").
		Preload("OrderItems.ArticleItem").
		Preload("OrderItems.ArticleItem.AllOems", func(db *gorm.DB) *gorm.DB {
			return db.Limit(100)
		}).
		Preload("OrderItems.ArticleItem.AllOems.Oem").
		Preload("OrderItems.ArticleItem.AllSpecifications").
		First(&order, id).Error

	return &order, err
}

func (r *OrderRepository) CreateOrderTransaction(
	order *roder.Order,
	invoice *roder.Invoice,
	items []roder.OrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(order).Error; err != nil {
			return err
		}

		invoice.OrderID = &order.ID

		if err := tx.Create(invoice).Error; err != nil {
			return err
		}

		for _, item := range items {
			item.OrderID = order.ID

			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
