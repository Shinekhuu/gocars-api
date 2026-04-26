package models

import (
	"time"

	"gorm.io/gorm"
)

type Invoice struct {
	gorm.Model
	InvoiceNumber string `gorm:"size:100;unique;not null"`
	OrderID       *uint
	UserID        *uint
	Amount        float64
	Paid          bool `gorm:"default:false"`
	PaidAt        *time.Time

	Order *Order `gorm:"constraint:OnDelete:CASCADE"` // pointer to avoid recursion
	User  *User  `gorm:"constraint:OnDelete:SET NULL"`
}
