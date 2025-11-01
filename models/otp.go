package models

import (
	"time"

	"gorm.io/gorm"
)

type Otp struct {
	gorm.Model                 // Embeds ID, CreatedAt, UpdatedAt, DeletedAt with default column types
	Email            string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	VerificationCode string    `gorm:"type:varchar(50);not null" json:"verification_code"`
	ExpiresAt        time.Time `gorm:"not null"`
}
