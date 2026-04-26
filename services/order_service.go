package services

import (
	"fmt"
	"gocars-api/dto"
	"gocars-api/mappers"
	"gocars-api/models"
	"gocars-api/repositories"
	"time"
)

func GetOrderByID(id int) (*dto.OrderSummaryDTO, error) {
	repo := repositories.NewOrderRepository()

	order, err := repo.GetOrderForSuccess(id)
	if err != nil {
		return nil, err
	}

	return mappers.MapOrderToSummaryDTO(order), nil
}

func GetOrderWithFitmentDataByID(id int) (*dto.OrderWithFitmentDataDTO, error) {
	repo := repositories.NewOrderRepository()

	order, err := repo.GetOrderForPDF(id)
	if err != nil {
		return nil, err
	}

	return mappers.MapOrderToFitmentDTO(order), nil
}

func CreateTransaction(input dto.CreateOrderInputs) (*models.Order, *models.Invoice, error) {
	var count int
	for _, item := range input.Items {
		count += item.Quantity
	}

	order := models.Order{
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		CompanyName:     input.CompanyName,
		Email:           input.Email,
		ShippingAddress: input.ShippingAddress,
		Status:          "pending",
		TotalItems:      count,
	}

	invoice := models.Invoice{
		InvoiceNumber: fmt.Sprintf("INV-%d", time.Now().Unix()),
		Paid:          false,
	}

	repo := repositories.NewOrderRepository()

	var orderItems []models.OrderItem

	for _, item := range input.Items {
		orderItems = append(
			orderItems,
			models.OrderItem{
				ArticleItemID: item.ID,
				Quantity:      item.Quantity,
			},
		)
	}

	err := repo.CreateOrderTransaction(
		&order,
		&invoice,
		orderItems,
	)

	if err != nil {
		return nil, nil, err
	}

	return &order, &invoice, nil
}
