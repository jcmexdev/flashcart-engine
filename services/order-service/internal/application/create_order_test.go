package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/application"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/mocks"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
)

func TestCreateOrderUseCase_Success(t *testing.T) {
	mockUowMgr := mocks.NewMockUnitOfWorkManager(t)
	mockUow := mocks.NewMockUnitOfWork(t)
	mockOrderRepo := mocks.NewMockOrderRepository(t)
	mockOutboxRepo := mocks.NewMockOutboxRepository(t)

	var capturedOrder *domain.Order
	var capturedEvent *domain.OutboxEvent

	mockUowMgr.EXPECT().
		ExecuteInTx(mock.Anything, mock.AnythingOfType("func(ports.UnitOfWork) error")).
		RunAndReturn(func(ctx context.Context, fn func(ports.UnitOfWork) error) error {
			return fn(mockUow)
		}).
		Once()

	mockUow.EXPECT().OrderRepository().Return(mockOrderRepo).Once()
	mockUow.EXPECT().OutboxRepository().Return(mockOutboxRepo).Once()

	mockOrderRepo.EXPECT().
		Save(mock.Anything, mock.MatchedBy(func(order *domain.Order) bool {
			capturedOrder = order
			return order != nil && order.CustomerID == "cust_123" && order.TotalAmount == 99.99
		})).
		Return(nil).
		Once()

	mockOutboxRepo.EXPECT().
		Save(mock.Anything, mock.MatchedBy(func(event *domain.OutboxEvent) bool {
			capturedEvent = event
			return event != nil && event.EventType == domain.EventTypeOrderCreated
		})).
		Return(nil).
		Once()

	uc := application.NewCreateOrderUseCase(mockUowMgr)

	out, err := uc.Execute(context.Background(), ports.CreateOrderInput{
		CustomerID:  "cust_123",
		TotalAmount: 99.99,
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, out.Order)
	assert.Equal(t, "cust_123", out.Order.CustomerID)
	assert.Equal(t, 99.99, out.Order.TotalAmount)

	require.NotNil(t, capturedOrder)
	require.NotNil(t, capturedEvent)
	assert.Equal(t, capturedOrder.ID.String(), capturedEvent.AggregateID)
	assert.Equal(t, domain.EventTypeOrderCreated, capturedEvent.EventType)
}

func TestCreateOrderUseCase_InvalidInput(t *testing.T) {
	mockUowMgr := mocks.NewMockUnitOfWorkManager(t)
	uc := application.NewCreateOrderUseCase(mockUowMgr)

	out, err := uc.Execute(context.Background(), ports.CreateOrderInput{
		CustomerID:  "",
		TotalAmount: -10,
	})

	require.Error(t, err)
	assert.Nil(t, out)

	appErr, ok := apperror.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, apperror.CodeInvalidInput, appErr.Code)
}

func TestCreateOrderUseCase_UoWError(t *testing.T) {
	mockUowMgr := mocks.NewMockUnitOfWorkManager(t)
	dbErr := errors.New("db error")

	mockUowMgr.EXPECT().
		ExecuteInTx(mock.Anything, mock.AnythingOfType("func(ports.UnitOfWork) error")).
		RunAndReturn(func(ctx context.Context, fn func(ports.UnitOfWork) error) error {
			return apperror.NewInternal("transaction failed", dbErr)
		}).
		Once()

	uc := application.NewCreateOrderUseCase(mockUowMgr)

	out, err := uc.Execute(context.Background(), ports.CreateOrderInput{
		CustomerID:  "cust_123",
		TotalAmount: 100,
	})

	require.Error(t, err)
	assert.Nil(t, out)

	appErr, ok := apperror.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, apperror.CodeInternal, appErr.Code)
}
