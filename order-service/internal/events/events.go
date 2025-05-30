package events

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
	TotalPrice float64
	CreatedAt  time.Time
}
