package service

import (
	"gocars-api/internal/profile/repository/postgresql/model"
	profilerepo "gocars-api/internal/profile/repository/postgresql"
)

type ProfileService struct {
	repo *profilerepo.ProfileRepository
}

func NewProfileService(r *profilerepo.ProfileRepository) *ProfileService {
	return &ProfileService{repo: r}
}

func (s *ProfileService) GetProfile(userID string) (*model.Profile, error) {
	return s.repo.GetByID(userID)
}

func (s *ProfileService) UpsertProfile(p *model.Profile) error {
	return s.repo.Upsert(p)
}
