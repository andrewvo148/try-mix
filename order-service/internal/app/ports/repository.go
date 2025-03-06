package ports

import (
	"context"
	"order-service/internal/domain"
)

// UnitOfWork defines the interfaces for managing transactions
type UnitOfWork interface {
	// Execute runs the given function within a tracsaction
	Execute(ctx context.Context, fn func() error) error

	// Orders returns the order repository for the current transaction
	Orders() OrderRepository

	// OutboxMessages returns the outbox repository for the current transaction
	OutboxMessages() OutboxRepository

}
// OrderRepository defines the interface for order data access
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*domain.Order, error)
}

type OutboxRepository interface {
	CreateMessage(ctx context.Context, aggregateID, messageType string, payload interface{}) error
	GetPendingMessages(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	MarkMessageAsProcessed(ctx context.Context, messageID string) error
	MarkMessageAsFailed(ctx context.Context, messageID string, reason string) error
}


