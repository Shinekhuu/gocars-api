package service

import (
	"fmt"
	"time"

	handlrdto "gocars-api/internal/roder/handler/dto"
	roder "gocars-api/internal/roder/repository/postgresql/model"
	"gocars-api/internal/roder/repository"
)

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(r *repository.OrderRepository) *OrderService {
	return &OrderService{repo: r}
}

func (s *OrderService) GetOrderByID(id int) (*handlrdto.OrderSummaryDTO, error) {
	order, err := s.repo.GetOrderForSuccess(id)
	if err != nil {
		return nil, err
	}

	return handlrdto.MapOrderToSummaryDTO(order), nil
}

func (s *OrderService) GetOrderWithFitmentDataByID(id int) (*handlrdto.OrderWithFitmentDataDTO, error) {
	order, err := s.repo.GetOrderForPDF(id)
	if err != nil {
		return nil, err
	}

	return handlrdto.MapOrderToFitmentDTO(order), nil
}

func (s *OrderService) CreateTransaction(input handlrdto.CreateOrderInputs) (*roder.Order, *roder.Invoice, error) {
	var count int
	for _, item := range input.Items {
		count += item.Quantity
	}

	order := roder.Order{
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		CompanyName:     input.CompanyName,
		Email:           input.Email,
		ShippingAddress: input.ShippingAddress,
		Status:          "pending",
		TotalItems:      count,
	}

	invoice := roder.Invoice{
		InvoiceNumber: fmt.Sprintf("INV-%d", time.Now().Unix()),
		Paid:          false,
	}

	var orderItems []roder.OrderItem

	for _, item := range input.Items {
		orderItems = append(
			orderItems,
			roder.OrderItem{
				ArticleItemID: item.ID,
				Quantity:      item.Quantity,
			},
		)
	}

	err := s.repo.CreateOrderTransaction(&order, &invoice, orderItems)

	if err != nil {
		return nil, nil, err
	}

	return &order, &invoice, nil
}
