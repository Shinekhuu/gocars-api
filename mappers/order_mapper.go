package mappers

import (
	"gocars-api/dto"
	"gocars-api/models"
)

func mapOrderBaseDTO(
	order *models.Order,
) dto.OrderBaseDTO {

	companyName := ""
	if order.CompanyName != nil {
		companyName = *order.CompanyName
	}

	invoiceNumber := ""
	if order.Invoice != nil {
		invoiceNumber = order.Invoice.InvoiceNumber
	}

	return dto.OrderBaseDTO{
		ID:              order.ID,
		FirstName:       order.FirstName,
		LastName:        order.LastName,
		CompanyName:     companyName,
		Email:           order.Email,
		Phone:           order.Phone,
		ShippingAddress: order.ShippingAddress,
		Status:          order.Status,
		TotalItems:      order.TotalItems,
		InvoiceNumber:   invoiceNumber,
		CreatedAt: order.CreatedAt.Format(
			"2006-01-02 15:04:05",
		),
	}
}

//
// ---------- Public Mappers ----------
//

func MapOrderToSummaryDTO(
	order *models.Order,
) *dto.OrderSummaryDTO {

	response := &dto.OrderSummaryDTO{
		OrderBaseDTO: mapOrderBaseDTO(order),
		Items: make(
			[]dto.OrderItemSummaryDTO,
			0,
			len(order.OrderItems),
		),
	}

	for _, item := range order.OrderItems {
		response.Items = append(
			response.Items,
			mapOrderItemSummaryDTO(item),
		)
	}

	return response
}

func MapOrderToFitmentDTO(
	order *models.Order,
) *dto.OrderWithFitmentDataDTO {

	response := &dto.OrderWithFitmentDataDTO{
		OrderBaseDTO: mapOrderBaseDTO(order),
		Items: make(
			[]dto.OrderItemWithFitmentDataDTO,
			0,
			len(order.OrderItems),
		),
	}

	for _, item := range order.OrderItems {
		response.Items = append(
			response.Items,
			mapOrderItemWithFitmentDTO(item),
		)
	}

	return response
}

// ---------- Item Mappers ----------
func mapOrderItemSummaryDTO(
	item models.OrderItem,
) dto.OrderItemSummaryDTO {

	dtoItem := dto.OrderItemSummaryDTO{
		OrderItemDTO: mapOrderItemDTO(item),
	}

	if item.ArticleItem == nil {
		return dtoItem
	}

	dtoItem.AllSpecifications =
		mapSpecificationsToDTO(
			item.ArticleItem.AllSpecifications,
		)

	return dtoItem
}

func mapOrderItemDTO(
	item models.OrderItem,
) dto.OrderItemDTO {

	dtoItem := dto.OrderItemDTO{
		ID:            item.ID,
		ArticleItemID: item.ArticleItemID,
		Quantity:      item.Quantity,
	}

	if item.ArticleItem == nil {
		return dtoItem
	}

	dtoItem.ArticleProductName =
		item.ArticleItem.ArticleProductName

	dtoItem.SupplierName =
		item.ArticleItem.SupplierName

	dtoItem.S3Image =
		item.ArticleItem.S3Image

	return dtoItem
}

func mapOrderItemWithFitmentDTO(
	item models.OrderItem,
) dto.OrderItemWithFitmentDataDTO {

	dtoItem := dto.OrderItemWithFitmentDataDTO{
		OrderItemDTO: mapOrderItemDTO(item),
	}

	if item.ArticleItem == nil {
		return dtoItem
	}

	dtoItem.FitmentDTO = dto.FitmentDTO{
		OEMS: mapOemsToDTO(
			item.ArticleItem.AllOems,
		),
		AllSpecifications: mapSpecificationsToDTO(
			item.ArticleItem.AllSpecifications,
		),
	}

	return dtoItem
}

//
// ---------- Helpers ----------
//

func mapOemsToDTO(
	oems []models.ArticleOem,
) []dto.OemDTO {

	result := make(
		[]dto.OemDTO,
		0,
		len(oems),
	)

	for _, oem := range oems {
		result = append(
			result,
			dto.OemDTO{
				Brand:     oem.Oem.Brand,
				DisplayNo: oem.Oem.DisplayNo,
			},
		)
	}

	return result
}

func mapSpecificationsToDTO(
	specs []models.ArticleAllSpecification,
) []dto.AllSpecificationDTO {

	result := make(
		[]dto.AllSpecificationDTO,
		0,
		len(specs),
	)

	for _, s := range specs {
		result = append(
			result,
			dto.AllSpecificationDTO{
				CriteriaName:  s.CriteriaName,
				CriteriaValue: s.CriteriaValue,
			},
		)
	}

	return result
}
