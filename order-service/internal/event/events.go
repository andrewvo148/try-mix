package event

import (
	"order-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

type OrderCreatedEvent struct {
	EventID    uuid.UUID
	SageID     uuid.UUID
	OrderID    uuid.UUID
	CustomerID string
	Items      []domain.OrderItem
	CreatedAt  time.Time
}
