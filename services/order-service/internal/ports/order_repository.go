package ports

import (
	"context"

	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
)

// OrderRepository defines the persistence outbound port dedicated to orders.
type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
}
