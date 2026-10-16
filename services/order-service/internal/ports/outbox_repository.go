package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
)


// OutboxRepository defines the persistence outbound port dedicated to transactional outbox events.
type OutboxRepository interface {
	Save(ctx context.Context, event *domain.OutboxEvent) error
	LockUnprocessed(ctx context.Context, limit int) ([]*domain.OutboxEvent, error)
	MarkAsProcessed(ctx context.Context, ids []uuid.UUID, processedAt time.Time) error
}

