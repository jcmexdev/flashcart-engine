package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID          uuid.UUID
	CustomerID  string
	TotalAmount float64
	Status      OrderStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewOrder creates and validates a new Order domain entity.
func NewOrder(customerID string, totalAmount float64) (*Order, error) {
	details := make(map[string]string)

	if customerID == "" {
		details["customer_id"] = "customer_id is required"
	}

	if totalAmount <= 0 {
		details["total_amount"] = "total_amount must be greater than zero"
	}

	if len(details) > 0 {
		return nil, apperror.NewInvalidInput("invalid order parameters", details)
	}

	now := time.Now().UTC()
	return &Order{
		ID:          uuid.New(),
		CustomerID:  customerID,
		TotalAmount: totalAmount,
		Status:      OrderStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
