package dto

import (
	articles "gocars-api/internal/articles/repository/postgresql/model"
	roder "gocars-api/internal/roder/repository/postgresql/model"
)

func mapOrderBaseDTO(order *roder.Order) OrderBaseDTO {
	companyName := ""
	if order.CompanyName != nil {
		companyName = *order.CompanyName
	}

	invoiceNumber := ""
	if order.Invoice != nil {
		invoiceNumber = order.Invoice.InvoiceNumber
	}

	return OrderBaseDTO{
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
		CreatedAt:       order.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func MapOrderToSummaryDTO(order *roder.Order) *OrderSummaryDTO {
	response := &OrderSummaryDTO{
		OrderBaseDTO: mapOrderBaseDTO(order),
		Items:        make([]OrderItemSummaryDTO, 0, len(order.OrderItems)),
	}

	for _, item := range order.OrderItems {
		response.Items = append(response.Items, mapOrderItemSummaryDTO(item))
	}

	return response
}

func MapOrderToFitmentDTO(order *roder.Order) *OrderWithFitmentDataDTO {
	response := &OrderWithFitmentDataDTO{
		OrderBaseDTO: mapOrderBaseDTO(order),
		Items:        make([]OrderItemWithFitmentDataDTO, 0, len(order.OrderItems)),
	}

	for _, item := range order.OrderItems {
		response.Items = append(response.Items, mapOrderItemWithFitmentDTO(item))
	}

	return response
}

func mapOrderItemSummaryDTO(item roder.OrderItem) OrderItemSummaryDTO {
	dtoItem := OrderItemSummaryDTO{
		OrderItemDTO: mapOrderItemDTO(item),
	}

	if item.ArticleItem == nil {
		return dtoItem
	}

	dtoItem.AllSpecifications = mapSpecificationsToDTO(item.ArticleItem.AllSpecifications)

	return dtoItem
}

func mapOrderItemDTO(item roder.OrderItem) OrderItemDTO {
	dtoItem := OrderItemDTO{
		ID:            item.ID,
		ArticleItemID: item.ArticleItemID,
		Quantity:      item.Quantity,
	}

	if item.ArticleItem == nil {
		return dtoItem
	}

	dtoItem.ArticleProductName = item.ArticleItem.ArticleProductName
	dtoItem.SupplierName = item.ArticleItem.SupplierName
	dtoItem.S3Image = item.ArticleItem.S3Image

	return dtoItem
}

func mapOrderItemWithFitmentDTO(item roder.OrderItem) OrderItemWithFitmentDataDTO {
	dtoItem := OrderItemWithFitmentDataDTO{
		OrderItemDTO: mapOrderItemDTO(item),
	}

	if item.ArticleItem == nil {
		return dtoItem
	}

	dtoItem.FitmentDTO = FitmentDTO{
		OEMS:              mapOemsToDTO(item.ArticleItem.AllOems),
		AllSpecifications: mapSpecificationsToDTO(item.ArticleItem.AllSpecifications),
	}

	return dtoItem
}

func mapOemsToDTO(oems []articles.ArticleOem) []OemDTO {
	result := make([]OemDTO, 0, len(oems))

	for _, oem := range oems {
		result = append(result, OemDTO{
			Brand:     oem.Oem.Brand,
			DisplayNo: oem.Oem.DisplayNo,
		})
	}

	return result
}

func mapSpecificationsToDTO(specs []articles.ArticleAllSpecification) []AllSpecificationDTO {
	result := make([]AllSpecificationDTO, 0, len(specs))

	for _, s := range specs {
		result = append(result, AllSpecificationDTO{
			CriteriaName:  s.CriteriaName,
			CriteriaValue: s.CriteriaValue,
		})
	}

	return result
}
