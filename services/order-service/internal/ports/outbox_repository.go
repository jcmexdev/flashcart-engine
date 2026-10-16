package ports

import (
	"context"

	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
)

// OutboxRepository defines the persistence outbound port dedicated to transactional outbox events.
type OutboxRepository interface {
	Save(ctx context.Context, event *domain.OutboxEvent) error
}
