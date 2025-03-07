package ports

import (
	"context"
	"order-service/internal/domain"
)

// UnitOfWork defines the interface for managing transactions
type UnitOfWork interface {
    // Begin starts a new transaction and returns a transaction context
    // This method should be called before accessing any repositories
    Begin(ctx context.Context) (context.Context, error)
    
    // Commit commits the current transaction
    // Returns an error if commit fails or if no transaction is active
    Commit(ctx context.Context) error
    
    // Rollback aborts the current transaction
    // Returns an error if rollback fails or if no transaction is active
    Rollback(ctx context.Context) error
    
    // Orders returns the order repository for the current transaction
    // The repository will use the active transaction from the context
    Orders(ctx context.Context) OrderRepository
    
    // OutboxMessages returns the outbox repository for the current transaction
    // The repository will use the active transaction from the context
    OutboxMessages(ctx context.Context) OutboxRepository
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


