package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
)

type outboxRepository struct {
	tx pgx.Tx
}

// NewOutboxRepository instantiates the outbox repository bound to an active database transaction.
func NewOutboxRepository(tx pgx.Tx) ports.OutboxRepository {
	return &outboxRepository{tx: tx}
}

// Save inserts an event into the outbox_events table.
func (r *outboxRepository) Save(ctx context.Context, event *domain.OutboxEvent) error {
	const insertOutboxQuery = `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, processed, created_at, processed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.tx.Exec(
		ctx,
		insertOutboxQuery,
		event.ID,
		event.AggregateType,
		event.AggregateID,
		event.EventType,
		event.Payload,
		event.Processed,
		event.CreatedAt,
		event.ProcessedAt,
	)
	if err != nil {
		return apperror.NewInternal("failed to insert outbox event", err)
	}

	return nil
}
