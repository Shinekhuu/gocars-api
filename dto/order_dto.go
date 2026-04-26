package dto

// ---------- Inputs ----------

type OrderItemInput struct {
	ID       uint `json:"id" binding:"required"`
	Quantity int  `json:"quantity" binding:"required,min=1"`
}

type CreateOrderInputs struct {
	FirstName       string  `json:"first_name" binding:"required"`
	LastName        string  `json:"last_name" binding:"required"`
	CompanyName     *string `json:"company_name"`
	Email           string  `json:"email" binding:"required,email"`
	Phone           string  `json:"phone" binding:"required"`
	ShippingAddress string  `json:"shipping_address" binding:"required"`

	Items []OrderItemInput `json:"items" binding:"required"`
}

//
// ---------- Shared DTOs ----------
//

type OrderBaseDTO struct {
	ID              uint   `json:"ID"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	CompanyName     string `json:"companyName"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	ShippingAddress string `json:"shippingAddress"`
	Status          string `json:"status"`
	TotalItems      int    `json:"totalItems"`
	InvoiceNumber   string `json:"invoiceNumber"`
	CreatedAt       string `json:"createdAt"`
}

type OrderItemDTO struct {
	ID                 uint   `json:"ID"`
	ArticleItemID      uint   `json:"articleItemId"`
	ArticleProductName string `json:"articleProductName"`
	Quantity           int    `json:"quantity"`
	SupplierName       string `json:"supplierName"`
	S3Image            string `json:"s3image"`
}

type FitmentDTO struct {
	OEMS              []OemDTO              `json:"oems"`
	AllSpecifications []AllSpecificationDTO `json:"allSpecifications"`
}

type OrderItemWithFitmentDataDTO struct {
	OrderItemDTO
	FitmentDTO
}

//
// ---------- Response DTOs ----------
//

// GET /order-success/:id
type OrderItemSummaryDTO struct {
	OrderItemDTO

	AllSpecifications []AllSpecificationDTO `json:"allSpecifications"`
}

type OrderSummaryDTO struct {
	OrderBaseDTO

	Items []OrderItemSummaryDTO `json:"items"`
}

// GET /order-pdf/:id
type OrderWithFitmentDataDTO struct {
	OrderBaseDTO
	Items []OrderItemWithFitmentDataDTO `json:"items"`
}
