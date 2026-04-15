package mappers

import (
	"gocars-api/dto"
	"gocars-api/models"
	"gocars-api/utils"
)

func buildArticleDTO(a models.ArticleItem) dto.ArticleDTO {
	return dto.ArticleDTO{
		ID:                   utils.UintPtrOrNilFromUint(a.ID),
		ArticleID:            a.ArticleID,
		ArticleSearchNo:      a.ArticleSearchNo,
		ArticleNo:            a.ArticleNo,
		ArticleProductName:   a.ArticleProductName,
		ManufacturerID:       uint(a.ManufacturerID),
		ManufacturerName:     a.ManufacturerName,
		SupplierID:           uint(a.SupplierID),
		SupplierName:         a.SupplierName,
		ArticleMediaType:     a.ArticleMediaType,
		ArticleMediaFileName: a.ArticleMediaFileName,
		S3Image:              a.S3Image,
	}
}

func ToDBResponse(
	article models.ArticleItem,
	engines []models.Engine,
	total int64,
	page, limit int,
) *dto.ArticleResponse {

	oems := make([]dto.OemDTO, len(article.AllOems))
	for i, link := range article.AllOems {
		oems[i] = dto.OemDTO{
			Brand:     link.Oem.Brand,
			DisplayNo: link.Oem.DisplayNo,
		}
	}

	return &dto.ArticleResponse{
		Article:           buildArticleDTO(article),
		Oems:              oems,
		AllSpecifications: article.AllSpecifications,
		CompatibleCars: dto.PaginationDTO{
			Page:  page,
			Limit: limit,
			Total: total,
			Data:  engines,
		},
	}
}

func ToAPIResponse(
	article models.ArticleItem,
	page, limit, offset int,
) *dto.ArticleResponse {

	total := len(article.CompatibleCarsResponse)

	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paged := article.CompatibleCarsResponse[start:end]

	oems := make([]dto.OemDTO, len(article.OemResponses))
	for i, o := range article.OemResponses {
		oems[i] = dto.OemDTO{
			Brand:     o.Brand,
			DisplayNo: o.DisplayNo,
		}
	}

	return &dto.ArticleResponse{
		Article:           buildArticleDTO(article),
		Oems:              oems,
		AllSpecifications: article.AllSpecifications,
		CompatibleCars: dto.PaginationDTO{
			Page:  page,
			Limit: limit,
			Total: int64(total),
			Data:  paged,
		},
	}
}
