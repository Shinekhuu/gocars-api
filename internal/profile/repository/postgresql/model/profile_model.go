package model

import "time"

// Profile stores app-specific user data in public.profiles.
// The ID is the UUID from Supabase auth.users — set it from the JWT sub claim.
type Profile struct {
	ID          string    `gorm:"type:uuid;primaryKey"          json:"id"`
	DisplayName string    `gorm:"type:text"                     json:"display_name"`
	AvatarURL   string    `gorm:"type:text"                     json:"avatar_url"`
	Phone       string    `gorm:"type:varchar(30)"              json:"phone"`
	CreatedAt   time.Time `                                     json:"created_at"`
	UpdatedAt   time.Time `                                     json:"updated_at"`
}

func (Profile) TableName() string { return "profiles" }
