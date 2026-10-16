package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
)

// unitOfWork implements ports.UnitOfWork by wrapping an active pgx transaction.
type unitOfWork struct {
	tx pgx.Tx
}

func (u *unitOfWork) OrderRepository() ports.OrderRepository {
	return NewOrderRepository(u.tx)
}

func (u *unitOfWork) OutboxRepository() ports.OutboxRepository {
	return NewOutboxRepository(u.tx)
}

// unitOfWorkManager coordinates PostgreSQL transactions implementing ports.UnitOfWorkManager.
type unitOfWorkManager struct {
	pool *pgxpool.Pool
}

// NewUnitOfWorkManager creates a new Unit of Work manager.
func NewUnitOfWorkManager(pool *pgxpool.Pool) ports.UnitOfWorkManager {
	return &unitOfWorkManager{
		pool: pool,
	}
}

// ExecuteInTx executes the given function within an atomic pgx transaction.
func (m *unitOfWorkManager) ExecuteInTx(ctx context.Context, fn func(uow ports.UnitOfWork) error) (err error) {
	tx, beginErr := m.pool.Begin(ctx)
	if beginErr != nil {
		return apperror.NewInternal("failed to begin transaction", beginErr)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-panic after ensuring rollback
		} else if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	uow := &unitOfWork{tx: tx}
	if err = fn(uow); err != nil {
		return err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return apperror.NewInternal("failed to commit transaction", commitErr)
	}

	return nil
}
