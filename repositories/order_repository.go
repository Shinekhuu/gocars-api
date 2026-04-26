package repositories

import (
	"gocars-api/database"
	"gocars-api/models"

	"gorm.io/gorm"
)

type OrderRepository struct{}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{}
}

func (r *OrderRepository) GetOrderForSuccess(id int) (*models.Order, error) {
	var order models.Order

	err := database.DB.
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

func (r *OrderRepository) GetOrderForPDF(id int) (*models.Order, error) {
	var order models.Order

	err := database.DB.
		Preload("Invoice").
		Preload("OrderItems").
		Preload("OrderItems.ArticleItem").
		Preload("OrderItems.ArticleItem.AllOems").
		Preload("OrderItems.ArticleItem.AllOems.Oem"). // IMPORTANT
		Preload("OrderItems.ArticleItem.AllSpecifications").
		First(&order, id).Error

	return &order, err
}

func (r *OrderRepository) CreateOrderTransaction(
	order *models.Order,
	invoice *models.Invoice,
	items []models.OrderItem) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {

		// 1. Create order
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 2. Create invoice
		invoice.OrderID = &order.ID

		if err := tx.Create(invoice).Error; err != nil {
			return err
		}

		// 3. Create order items
		for _, item := range items {

			item.OrderID = order.ID

			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
