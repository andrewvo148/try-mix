package ports

import (
	"context"
	"order-service/internal/domain"
)

// UnitOfWork defines the interface for managing transactions
type UnitOfWork interface {
	// This method should be called before accessing any  repositories
	Begin(ctx context.Context) error

	// Commit commits the current transaction
	// Returns an error if commit fails or if no transaction is active
	Commit(ctx context.Context) error

	// Rollback rolls back the current transaction
	// Returns an error if rollback fails or if no transaction is active
	Rollback(ctx context.Context) error

	// Orders returns the order repository for the current transaction
	Orders() OrderRepository
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
