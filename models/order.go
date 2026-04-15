package models

import (
	"errors"
	"fmt"
	"gocars-api/database"
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	UserID     uint
	User       *User
	Status     string `gorm:"size:50;default:'pending'"`
	Total      float64
	OrderItems []OrderItem
	Invoice    *Invoice `gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	gorm.Model
	OrderID   uint
	ArticleID uint `gorm:"column:article_id;index"`
	Quantity  int
	Price     float64
	Article   *ArticleItem `gorm:"foreignKey:ArticleID;references:ArticleID"`
}

func CreateTransaction(userID uint, articleIDs []uint) (Order, Invoice, error) {
	var createdOrder Order
	var createdInvoice Invoice

	err := database.DB.Transaction(func(tx *gorm.DB) error {

		// 1. Create Order
		order := Order{
			UserID: userID,
			Status: "pending",
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		createdOrder = order

		// 2. Fetch articles
		var articles []ArticleItem
		if err := tx.Where("article_id IN ?", articleIDs).Find(&articles).Error; err != nil {
			return err
		}

		if len(articles) == 0 {
			return errors.New("no products found for given article_ids")
		}

		// 3. Insert OrderItems
		total := 0.0
		for _, a := range articles {
			item := OrderItem{
				OrderID:   order.ID,
				ArticleID: *a.ArticleID,
				Quantity:  1,
				Price:     a.Price,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
			total += a.Price
		}

		// 4. Update total
		if err := tx.Model(&order).Update("total", total).Error; err != nil {
			return err
		}

		// 5. Create Invoice
		invoice := Invoice{
			InvoiceNumber: fmt.Sprintf("INV-%d", time.Now().Unix()),
			OrderID:       order.ID,
			UserID:        userID,
			Amount:        total,
			Paid:          false,
		}

		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}
		createdInvoice = invoice

		return nil
	})

	return createdOrder, createdInvoice, err
}
