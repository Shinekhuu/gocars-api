package models

import (
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	FirstName       string  `gorm:"type:varchar(255);"`
	LastName        string  `gorm:"type:varchar(255);"`
	CompanyName     *string `gorm:"type:varchar(255)"`
	Email           string  `gorm:"type:varchar(255);index"`
	Phone           string  `gorm:"type:varchar(255);index"`
	ShippingAddress string  `gorm:"type:varchar(255)"`
	Status          string  `gorm:"size:50;default:'pending'"`
	TotalItems      int
	Total           float64

	OrderItems []OrderItem
	Invoice    *Invoice `gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	gorm.Model
	OrderID       uint
	ArticleItemID uint `gorm:"column:article_item_id;index"`
	Quantity      int
	Price         float64

	ArticleItem *ArticleItem `gorm:"foreignKey:ArticleItemID;references:ID"`
}
