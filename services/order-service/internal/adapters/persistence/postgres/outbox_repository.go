package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
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

// LockUnprocessed queries unprocessed outbox events with FOR UPDATE SKIP LOCKED up to the given limit.
func (r *outboxRepository) LockUnprocessed(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	const query = `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, processed, created_at, processed_at
		FROM outbox_events
		WHERE processed = FALSE
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.tx.Query(ctx, query, limit)
	if err != nil {
		return nil, apperror.NewInternal("failed to query unprocessed outbox events", err)
	}
	defer rows.Close()

	var events []*domain.OutboxEvent
	for rows.Next() {
		var ev domain.OutboxEvent
		if err := rows.Scan(
			&ev.ID,
			&ev.AggregateType,
			&ev.AggregateID,
			&ev.EventType,
			&ev.Payload,
			&ev.Processed,
			&ev.CreatedAt,
			&ev.ProcessedAt,
		); err != nil {
			return nil, apperror.NewInternal("failed to scan outbox event", err)
		}
		events = append(events, &ev)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.NewInternal("error iterating outbox events", err)
	}

	return events, nil
}

// MarkAsProcessed updates the status of the specified events to processed.
func (r *outboxRepository) MarkAsProcessed(ctx context.Context, ids []uuid.UUID, processedAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}

	const query = `
		UPDATE outbox_events
		SET processed = TRUE, processed_at = $1
		WHERE id = ANY($2)
	`

	_, err := r.tx.Exec(ctx, query, processedAt, ids)
	if err != nil {
		return apperror.NewInternal("failed to mark outbox events as processed", err)
	}

	return nil
}

