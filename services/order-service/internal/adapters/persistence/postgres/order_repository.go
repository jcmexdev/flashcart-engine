package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
)

type orderRepository struct {
	tx pgx.Tx
}

// NewOrderRepository instantiates the order repository bound to an active database transaction.
func NewOrderRepository(tx pgx.Tx) ports.OrderRepository {
	return &orderRepository{tx: tx}
}

// Save inserts an order into the orders table.
func (r *orderRepository) Save(ctx context.Context, order *domain.Order) error {
	const insertOrderQuery = `
		INSERT INTO orders (id, customer_id, total_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.tx.Exec(
		ctx,
		insertOrderQuery,
		order.ID,
		order.CustomerID,
		order.TotalAmount,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return apperror.NewInternal("failed to insert order", err)
	}

	return nil
}
