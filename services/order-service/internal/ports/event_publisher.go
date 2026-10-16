package ports

import (
	"context"

	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
)

// EventPublisher defines the outbound port to publish domain events to an external message broker.
type EventPublisher interface {
	Publish(ctx context.Context, event *domain.OutboxEvent) error
}
