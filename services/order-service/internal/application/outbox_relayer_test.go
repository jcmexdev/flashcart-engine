package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"


	"github.com/jcmexdev/flashcart-engine/services/order/internal/application"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/mocks"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
)

func TestOutboxRelayer_ProcessBatch_Success(t *testing.T) {
	mockUowMgr := mocks.NewMockUnitOfWorkManager(t)
	mockUow := mocks.NewMockUnitOfWork(t)
	mockOutboxRepo := mocks.NewMockOutboxRepository(t)
	mockPublisher := mocks.NewMockEventPublisher(t)

	eventID1 := uuid.New()
	eventID2 := uuid.New()
	events := []*domain.OutboxEvent{
		{
			ID:            eventID1,
			AggregateType: domain.AggregateTypeOrder,
			AggregateID:   uuid.New().String(),
			EventType:     domain.EventTypeOrderCreated,
			Payload:       []byte(`{"order_id":"1"}`),
			Processed:     false,
			CreatedAt:     time.Now().UTC(),
		},
		{
			ID:            eventID2,
			AggregateType: domain.AggregateTypeOrder,
			AggregateID:   uuid.New().String(),
			EventType:     domain.EventTypeOrderCreated,
			Payload:       []byte(`{"order_id":"2"}`),
			Processed:     false,
			CreatedAt:     time.Now().UTC(),
		},
	}

	mockUowMgr.EXPECT().
		ExecuteInTx(mock.Anything, mock.AnythingOfType("func(ports.UnitOfWork) error")).
		RunAndReturn(func(ctx context.Context, fn func(ports.UnitOfWork) error) error {
			return fn(mockUow)
		}).
		Once()

	mockUow.EXPECT().OutboxRepository().Return(mockOutboxRepo).Twice()

	mockOutboxRepo.EXPECT().
		LockUnprocessed(mock.Anything, 10).
		Return(events, nil).
		Once()

	// Both events must be published
	mockPublisher.EXPECT().Publish(mock.Anything, events[0]).Return(nil).Once()
	mockPublisher.EXPECT().Publish(mock.Anything, events[1]).Return(nil).Once()

	// Both events must be marked as processed
	mockOutboxRepo.EXPECT().
		MarkAsProcessed(mock.Anything, []uuid.UUID{eventID1, eventID2}, mock.AnythingOfType("time.Time")).
		Return(nil).
		Once()

	relayer := application.NewOutboxRelayer(mockUowMgr, mockPublisher, application.OutboxRelayerConfig{
		PollInterval: 100 * time.Millisecond,
		BatchSize:    10,
	})

	relayer.ProcessBatch(context.Background())
}

func TestOutboxRelayer_ProcessBatch_NoEvents(t *testing.T) {
	mockUowMgr := mocks.NewMockUnitOfWorkManager(t)
	mockUow := mocks.NewMockUnitOfWork(t)
	mockOutboxRepo := mocks.NewMockOutboxRepository(t)
	mockPublisher := mocks.NewMockEventPublisher(t)

	mockUowMgr.EXPECT().
		ExecuteInTx(mock.Anything, mock.AnythingOfType("func(ports.UnitOfWork) error")).
		RunAndReturn(func(ctx context.Context, fn func(ports.UnitOfWork) error) error {
			return fn(mockUow)
		}).
		Once()

	mockUow.EXPECT().OutboxRepository().Return(mockOutboxRepo).Once()

	mockOutboxRepo.EXPECT().
		LockUnprocessed(mock.Anything, 50).
		Return([]*domain.OutboxEvent{}, nil).
		Once()

	relayer := application.NewOutboxRelayer(mockUowMgr, mockPublisher, application.OutboxRelayerConfig{})
	relayer.ProcessBatch(context.Background())
}

func TestOutboxRelayer_ProcessBatch_PublishFails_RollsBack(t *testing.T) {
	mockUowMgr := mocks.NewMockUnitOfWorkManager(t)
	mockUow := mocks.NewMockUnitOfWork(t)
	mockOutboxRepo := mocks.NewMockOutboxRepository(t)
	mockPublisher := mocks.NewMockEventPublisher(t)

	event := &domain.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: domain.AggregateTypeOrder,
		AggregateID:   "123",
		EventType:     domain.EventTypeOrderCreated,
		Payload:       []byte(`{}`),
	}

	publishErr := errors.New("network timeout to rabbitmq")

	mockUowMgr.EXPECT().
		ExecuteInTx(mock.Anything, mock.AnythingOfType("func(ports.UnitOfWork) error")).
		RunAndReturn(func(ctx context.Context, fn func(ports.UnitOfWork) error) error {
			return fn(mockUow)
		}).
		Once()

	mockUow.EXPECT().OutboxRepository().Return(mockOutboxRepo).Once()
	mockOutboxRepo.EXPECT().LockUnprocessed(mock.Anything, 50).Return([]*domain.OutboxEvent{event}, nil).Once()

	mockPublisher.EXPECT().Publish(mock.Anything, event).Return(publishErr).Once()

	// MarkAsProcessed should NOT be called because publishing failed
	relayer := application.NewOutboxRelayer(mockUowMgr, mockPublisher, application.OutboxRelayerConfig{})
	relayer.ProcessBatch(context.Background())
}

func TestOutboxRelayer_Start_GracefulShutdown(t *testing.T) {
	mockUowMgr := mocks.NewMockUnitOfWorkManager(t)
	mockUow := mocks.NewMockUnitOfWork(t)
	mockOutboxRepo := mocks.NewMockOutboxRepository(t)
	mockPublisher := mocks.NewMockEventPublisher(t)

	mockUowMgr.EXPECT().
		ExecuteInTx(mock.Anything, mock.AnythingOfType("func(ports.UnitOfWork) error")).
		RunAndReturn(func(ctx context.Context, fn func(ports.UnitOfWork) error) error {
			return fn(mockUow)
		}).
		Maybe()

	mockUow.EXPECT().OutboxRepository().Return(mockOutboxRepo).Maybe()
	mockOutboxRepo.EXPECT().LockUnprocessed(mock.Anything, mock.Anything).Return([]*domain.OutboxEvent{}, nil).Maybe()

	relayer := application.NewOutboxRelayer(mockUowMgr, mockPublisher, application.OutboxRelayerConfig{
		PollInterval: 10 * time.Millisecond,
		BatchSize:    10,
	})

	ctx, cancel := context.WithCancel(context.Background())

	stopped := make(chan struct{})
	go func() {
		relayer.Start(ctx)
		close(stopped)
	}()

	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case <-stopped:
		// Succeeded gracefully
	case <-time.After(1 * time.Second):
		t.Fatal("Relayer did not shutdown gracefully within timeout")
	}
}
