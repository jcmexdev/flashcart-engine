package application

import (
	"context"

	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
)

type createOrderUseCase struct {
	uowManager ports.UnitOfWorkManager
}

// NewCreateOrderUseCase constructs a new instance of the CreateOrder use case injecting the UnitOfWorkManager.
func NewCreateOrderUseCase(uowManager ports.UnitOfWorkManager) ports.CreateOrderUseCase {
	return &createOrderUseCase{
		uowManager: uowManager,
	}
}

// Execute coordinates order creation and outbox event publishing via the transactional Unit of Work.
func (uc *createOrderUseCase) Execute(ctx context.Context, input ports.CreateOrderInput) (*ports.CreateOrderOutput, error) {
	// 1. Create and validate the Order domain entity
	order, err := domain.NewOrder(input.CustomerID, input.TotalAmount)
	if err != nil {
		return nil, err
	}

	// 2. Create corresponding Outbox event
	event, err := domain.NewOrderCreatedOutboxEvent(order)
	if err != nil {
		return nil, err
	}

	// 3. Persist atomically within transactional Unit of Work
	err = uc.uowManager.ExecuteInTx(ctx, func(uow ports.UnitOfWork) error {
		if err := uow.OrderRepository().Save(ctx, order); err != nil {
			return err
		}

		if err := uow.OutboxRepository().Save(ctx, event); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ports.CreateOrderOutput{
		Order: order,
	}, nil
}
