package domain

import (
	"context"

	"github.com/google/uuid"
)

// OrderRepository defines the interface for order data access
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Order, error)
}

// OutboxRepository defines the interface for outbox operations
type OutboxRepository interface {
	CreateMessage(ctx context.Context, aggregateID uuid.UUID, eventType string, payload interface{}) error
	GetPendingMessages(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkMessageAsProcessed(ctx context.Context, messageID uuid.UUID) error
	MarkMessageAsFailed(ctx context.Context, messageID uuid.UUID, reason string) error
	IncrementAttempt(ctx context.Context, messageID uuid.UUID) error
}
