package http

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
)

// OrderResponse represents the HTTP JSON response returned upon successful order creation.
type OrderResponse struct {
	ID          uuid.UUID          `json:"id"`
	CustomerID  string             `json:"customer_id"`
	TotalAmount float64            `json:"total_amount"`
	Status      domain.OrderStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}
