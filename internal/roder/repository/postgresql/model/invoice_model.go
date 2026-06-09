package model

import (
	"time"

	profile "gocars-api/internal/profile/repository/postgresql/model"

	"gorm.io/gorm"
)

type Invoice struct {
	gorm.Model
	InvoiceNumber string  `gorm:"size:100;unique;not null"`
	OrderID       *uint
	UserID        *string `gorm:"type:uuid"` // references public.profiles (Supabase user UUID)
	Amount        float64
	Paid          bool `gorm:"default:false"`
	PaidAt        *time.Time

	Order *Order        `gorm:"constraint:OnDelete:CASCADE"`
	User  *profile.Profile `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}
