package event

import (
	"order-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

type OrderCreatedEvent struct {
	EventID    uuid.UUID          `json:"event_id"`
	SagaID     uuid.UUID          `json:"saga_id"`
	OrderID    uuid.UUID          `json:"order_id"`
	CustomerID string             `json:"customer_id"`
	Items      []domain.OrderItem `json:"items"`
	TotalPrice float64            `json:"total_price"`
	CreatedAt  time.Time          `json:"created_at"`
}
