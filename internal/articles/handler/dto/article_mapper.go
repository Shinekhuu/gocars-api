package dto

import (
	articles "gocars-api/internal/articles/repository/postgresql/model"
	"gocars-api/internal/shared/utils"
	vehicle "gocars-api/internal/vehicle/repository/postgresql/model"
)

func buildArticleDTO(a articles.ArticleItem) ArticleDTO {
	return ArticleDTO{
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
	article articles.ArticleItem,
	engines []vehicle.Engine,
	total int64,
	page, limit int,
) *ArticleResponse {

	oems := make([]OemDTO, len(article.AllOems))
	for i, link := range article.AllOems {
		oems[i] = OemDTO{
			Brand:     link.Oem.Brand,
			DisplayNo: link.Oem.DisplayNo,
		}
	}

	return &ArticleResponse{
		Article:           buildArticleDTO(article),
		Oems:              oems,
		AllSpecifications: article.AllSpecifications,
		CompatibleCars: PaginationDTO{
			Page:  page,
			Limit: limit,
			Total: total,
			Data:  engines,
		},
	}
}

func ToAPIResponse(
	article articles.ArticleItem,
	page, limit, offset int,
) *ArticleResponse {

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

	oems := make([]OemDTO, len(article.OemResponses))
	for i, o := range article.OemResponses {
		oems[i] = OemDTO{
			Brand:     o.Brand,
			DisplayNo: o.DisplayNo,
		}
	}

	return &ArticleResponse{
		Article:           buildArticleDTO(article),
		Oems:              oems,
		AllSpecifications: article.AllSpecifications,
		CompatibleCars: PaginationDTO{
			Page:  page,
			Limit: limit,
			Total: int64(total),
			Data:  paged,
		},
	}
}
