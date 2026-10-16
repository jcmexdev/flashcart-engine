package application

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
)

type OutboxRelayerConfig struct {
	PollInterval time.Duration
	BatchSize    int
}

// OutboxRelayer is the background worker that periodically polls unprocessed outbox events
// and publishes them to the message broker.
type OutboxRelayer struct {
	uowManager ports.UnitOfWorkManager
	publisher  ports.EventPublisher
	cfg        OutboxRelayerConfig
}

// NewOutboxRelayer creates a new OutboxRelayer instance.
func NewOutboxRelayer(
	uowManager ports.UnitOfWorkManager,
	publisher ports.EventPublisher,
	cfg OutboxRelayerConfig,
) *OutboxRelayer {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	return &OutboxRelayer{
		uowManager: uowManager,
		publisher:  publisher,
		cfg:        cfg,
	}
}

// Start begins the polling loop in a blocking manner until the context is cancelled.
func (r *OutboxRelayer) Start(ctx context.Context) {
	log.Printf("[Outbox Relayer] Started polling every %v (batch size: %d)...", r.cfg.PollInterval, r.cfg.BatchSize)
	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()

	// Run immediate first cycle
	r.ProcessBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("[Outbox Relayer] Polling loop stopped gracefully.")
			return
		case <-ticker.C:
			r.ProcessBatch(ctx)
		}
	}
}

// ProcessBatch executes one cycle: queries locked events, publishes to broker, and marks as processed.
func (r *OutboxRelayer) ProcessBatch(ctx context.Context) {
	err := r.uowManager.ExecuteInTx(ctx, func(uow ports.UnitOfWork) error {
		events, err := uow.OutboxRepository().LockUnprocessed(ctx, r.cfg.BatchSize)
		if err != nil {
			return err
		}

		if len(events) == 0 {
			return nil
		}

		log.Printf("[Outbox Relayer] Found %d unprocessed event(s) to publish", len(events))

		var processedIDs []uuid.UUID
		for _, ev := range events {
			// Check context cancellation between items
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err := r.publisher.Publish(ctx, ev); err != nil {
				log.Printf("[Outbox Relayer] Failed to publish event %s: %v. Rolling back transaction.", ev.ID, err)
				return err
			}
			processedIDs = append(processedIDs, ev.ID)
		}

		if len(processedIDs) > 0 {
			if err := uow.OutboxRepository().MarkAsProcessed(ctx, processedIDs, time.Now().UTC()); err != nil {
				log.Printf("[Outbox Relayer] Failed to mark events as processed: %v", err)
				return err
			}
			log.Printf("[Outbox Relayer] Successfully published and marked %d event(s)", len(processedIDs))
		}

		return nil
	})

	if err != nil && ctx.Err() == nil {
		log.Printf("[Outbox Relayer] Error during batch processing: %v", err)
	}
}
