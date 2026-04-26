// internal/service/oem_service.go
package services

import (
	"gocars-api/dto"
	"gocars-api/repositories"
	"gocars-api/utils"
)

func GetOEMs(
	ids []*int,
	articleIDs []*int,
) ([]dto.OEMResponse, error) {

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

	articles, err := repositories.GetWithOEMs(dbIDs, extIDs)
	if err != nil {
		return nil, err
	}

	result := make([]dto.OEMResponse, len(articles))

	for i, a := range articles {
		oems := make([]dto.OemDTO, 0, len(a.AllOems))

		for _, ao := range a.AllOems {
			oems = append(oems, dto.OemDTO{
				Brand:     ao.Oem.Brand,
				DisplayNo: ao.Oem.DisplayNo,
			})
		}

		result[i] = dto.OEMResponse{
			ID:        utils.UintToIntPtr(a.ID),
			ArticleID: utils.UintPtrToIntPtr(a.ArticleID),
			OEMs:      oems,
		}
	}

	return result, nil
}
