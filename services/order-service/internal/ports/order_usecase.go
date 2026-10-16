package ports

import (
	"context"

	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
)

// CreateOrderInput holds the input parameters for the CreateOrder use case.
type CreateOrderInput struct {
	CustomerID  string  `json:"customer_id"`
	TotalAmount float64 `json:"total_amount"`
}

// CreateOrderOutput holds the output result from the CreateOrder use case.
type CreateOrderOutput struct {
	Order *domain.Order
}

// CreateOrderUseCase defines the inbound port for creating orders.
type CreateOrderUseCase interface {
	Execute(ctx context.Context, input CreateOrderInput) (*CreateOrderOutput, error)
}
