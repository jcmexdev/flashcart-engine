package ports

import "context"

// UnitOfWork represents a transactional unit of work.
// It provides access to repositories sharing the same ACID transaction context.
type UnitOfWork interface {
	OrderRepository() OrderRepository
	OutboxRepository() OutboxRepository
}

// UnitOfWorkManager coordinates the execution, commit, and rollback of a UnitOfWork.
type UnitOfWorkManager interface {
	// ExecuteInTx executes the provided function within an atomic transaction.
	// If fn returns an error or panics, the transaction is rolled back. If fn returns nil, it is committed.
	ExecuteInTx(ctx context.Context, fn func(uow UnitOfWork) error) error
}
