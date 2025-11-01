package models

import "gorm.io/gorm"

type User struct {
	gorm.Model        // Embeds ID, CreatedAt, UpdatedAt, DeletedAt with default column types
	Name       string `gorm:"type:text" json:"name"`
	Email      string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password   string `gorm:"type:varchar(250);not null"`
	IsVerified bool   `gorm:"default:false" json:"is_verified"`
}
