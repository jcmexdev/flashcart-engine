package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
)

const (
	AggregateTypeOrder    = "ORDER"
	EventTypeOrderCreated = "ORDER_CREATED"
)

type OutboxEvent struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
	Processed     bool
	CreatedAt     time.Time
	ProcessedAt   *time.Time
}

// OrderCreatedPayload represents the serialized payload stored within the outbox event.
type OrderCreatedPayload struct {
	OrderID     string    `json:"order_id"`
	CustomerID  string    `json:"customer_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewOrderCreatedOutboxEvent serializes the order and creates an outbox event ready for persistence.
func NewOrderCreatedOutboxEvent(order *Order) (*OutboxEvent, error) {
	payload := OrderCreatedPayload{
		OrderID:     order.ID.String(),
		CustomerID:  order.CustomerID,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		CreatedAt:   order.CreatedAt,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, apperror.NewInternal("failed to serialize order created event payload", err)
	}

	return &OutboxEvent{
		ID:            uuid.New(),
		AggregateType: AggregateTypeOrder,
		AggregateID:   order.ID.String(),
		EventType:     EventTypeOrderCreated,
		Payload:       payloadBytes,
		Processed:     false,
		CreatedAt:     order.CreatedAt,
		ProcessedAt:   nil,
	}, nil
}
