package service

import (
	handlrdto "gocars-api/internal/articles/handler/dto"
	repo "gocars-api/internal/articles/repository/postgresql"
	"gocars-api/internal/shared/utils"
)

type OemService struct {
	repo *repo.OemRepository
}

func NewOemService(r *repo.OemRepository) *OemService {
	return &OemService{repo: r}
}

func (s *OemService) GetOEMs(ids []*int, articleIDs []*int) ([]handlrdto.OEMResponse, error) {
	var dbIDs []uint
	var extIDs []uint

	for _, id := range ids {
		if id != nil {
			dbIDs = append(dbIDs, uint(*id))
		}
	}

	for _, a := range articleIDs {
		if a != nil {
			extIDs = append(extIDs, uint(*a))
		}
	}

	articles, err := s.repo.GetWithOEMs(dbIDs, extIDs)
	if err != nil {
		return nil, err
	}

	result := make([]handlrdto.OEMResponse, len(articles))

	for i, a := range articles {
		oems := make([]handlrdto.OemDTO, 0, len(a.AllOems))

		for _, ao := range a.AllOems {
			oems = append(oems, handlrdto.OemDTO{
				Brand:     ao.Oem.Brand,
				DisplayNo: ao.Oem.DisplayNo,
			})
		}

		result[i] = handlrdto.OEMResponse{
			ID:        utils.UintToIntPtr(a.ID),
			ArticleID: utils.UintPtrToIntPtr(a.ArticleID),
			OEMs:      oems,
		}
	}

	return result, nil
}
