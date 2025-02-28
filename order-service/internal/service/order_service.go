package service

import (
	"time"

	"github.com/google/uuid"
	"order-service/internal/event"
	"order-service/internal/model"
	"order-service/internal/repository"
)

type OrderService struct {
	repo         *repository.OrderRepository
	eventHandler *event.KafkaHandler
}

func NewOrderService(repo *repository.OrderRepository, eventHandler *event.KafkaHandler) *OrderService {
	return &OrderService{
		repo:         repo,
		eventHandler: eventHandler,
	}
}

func (s *OrderService) CreateOrder(customerID string, items []model.OrderItem) (*model.Order, error) {
	// Calculate total amount
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.TotalPrice
	}

	order := &model.Order{
		ID:          uuid.New(),
		CustomerID:  customerID,
		Items:       items,
		TotalAmount: totalAmount,
		Status:      model.OrderStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Save(order); err != nil {
		return nil, err
	}

	// Publish order created event
	if err := s.eventHandler.PublishOrderEvent(order, "ORDER_CREATED"); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) UpdateOrderStatus(orderID string, status model.OrderStatus) (*model.Order, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	if err := s.repo.Update(order); err != nil {
		return nil, err
	}

	// Publish order status updated event
	if err := s.eventHandler.PublishOrderEvent(order, "ORDER_STATUS_UPDATED"); err != nil {
		return nil, err
	}

	return order, nil
} 