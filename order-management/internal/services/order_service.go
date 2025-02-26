package services

import (
	"fmt"
	"log"
	"order-management/internal/domain"
	"time"

	"github.com/google/uuid"
	"modernc.org/libc/uuid/uuid"
)

type OrderService struct {
	producer *domain.Producer
	orders   map[string]*domain.Order
}

// NewOrderService creates a new OrderService
func NewOrderService() *OrderService {
	return &OrderService{
		orders: make(map[string]*domain.Order),
	}
}

// HandleCreateOrder handles the CreateOrder command
func (s *OrderService) HandleCreateOrder(cmd *domain.CreateOrderCommand) error {
	orderID := uuid.New().string()
	order := &domain.Order{
		ID:         orderID,
		CustomerID: cmd.CustomerID,
		Products:   cmd.Products,
		Status:     "Created",
		CreatedAt:  time.Now(),
	}

	// Store order in memory
	s.orders[orderID] = order

	log.Printf("Order created: %s", orderID)

	// Publish order created event
	if s.producer != nil {
		err := s.producer.PublishEvent("order.created", domain.OrderCreatedEventType, order)
		if err != nil {
			return fmt.Errorf("failed to publish order created event: %w", err)
		}
	}

	return nil
}
