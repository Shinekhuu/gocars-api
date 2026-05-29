package repository

import (
	"gocars-api/internal/profile/repository/postgresql/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) GetByID(id string) (*model.Profile, error) {
	var p model.Profile
	if err := r.db.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProfileRepository) Upsert(p *model.Profile) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"display_name", "avatar_url", "phone", "updated_at"}),
	}).Create(p).Error
}
