package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"microservice-demo/internal/order-service/app/ports"
	"microservice-demo/internal/order-service/domain"
	"microservice-demo/internal/order-service/events"

	"microservice-demo/internal/order-service/infrastructure/repository"
	"time"

	"github.com/google/uuid"
)

// OrderUsecase implements the order business logic
type OrderUseCase struct {
	eventPublisher ports.EventPublisher
	uow            ports.UnitOfWork
}

// AddOrderItem implements ports.OrderUseCase.
func (uc *OrderUseCase) AddOrderItem(ctx context.Context, orderID string, productID string, quantity int32, price float64) error {
	panic("unimplemented")
}

// CancelOrder implements ports.OrderUseCase.
func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) error {
	panic("unimplemented")
}

// GetOrder implements ports.OrderUseCase.
func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	panic("unimplemented")
}

// ListOrders implements ports.OrderUseCase.
func (uc *OrderUseCase) ListOrders(ctx context.Context, limit int, offset int) ([]*domain.Order, error) {
	panic("unimplemented")
}

// RemoveOrderItem implements ports.OrderUseCase.
func (uc *OrderUseCase) RemoveOrderItem(ctx context.Context, orderID string, itemID string) error {
	panic("unimplemented")
}

// UpdateOrderStatus implements ports.OrderUseCase.
func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	panic("unimplemented")
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(
	uow ports.UnitOfWork,
	orderRepo domain.OrderRepository,
	outboxRepo domain.OutboxRepository,
	eventPublisher ports.EventPublisher,
) *OrderUseCase {
	return &OrderUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
	}
}

// CreateOrder creates a new order with the given details
func (uc *OrderUseCase) CreateOrder(
	ctx context.Context,
	customerID string,
	items []domain.OrderItem,
) (*domain.Order, error) {
	// Validate input
	if customerID == "" {
		return nil, domain.ErrInvalidCustomerID
	}
	if len(items) == 0 {
		return nil, domain.ErrEmptyOrderItems
	}

	// Create order entity
	order := domain.NewOrder(customerID, items)
	order.Status = domain.OrderStatusPending
	order.SagaID = uuid.New()

	// Prepare order created event
	orderCreatedEvent := events.OrderCreatedEvent{
		EventID:    uuid.New(),
		SageID:     order.SagaID,
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		Items:      order.Items,
		TotalPrice: order.TotalPrice,
		CreatedAt:  time.Now(),
	}

	// Marshal event payload
	eventPayload, err := json.Marshal(orderCreatedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order created event: %w", err)
	}



	err = uc.uow.Execute(ctx, func(tx *sql.Tx) error {
		// Create order
		repository.NewOrderRepository(tx.)
		if err := uc.orderRepo.Create(ctx, order); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		if err = uc.outboxRepo.CreateMessage(
			ctx,
			order.ID,
			"order.created",
			eventPayload,
		); err != nil {
			return fmt.Errorf("failed to create outbox message: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	// After transaction is committed, publish event to broker
	// This is done outside the transaction for "at-least-once" delivery semantics
	// If publishing fails, the event is still in the outbox table and can be published later by an outbox processor
	err = uc.publishEvent(ctx, "order.created", orderCreatedEvent)
	if err != nil {
		// Log the error but don't fail the operation
		// The outbox pattern ensures events will be delivered eventually
		log.Printf("warning: failed to publish order.created event: %v", err)
	}

	return order, nil
}

// publishEvent publishes an event to the message broker
func (uc *OrderUseCase) publishEvent(ctx context.Context, topic string, event interface{}) error {
	// Publish the event to the message broker
	err := uc.eventPublisher.Publish(ctx, topic, event)
	if err != nil {
		return fmt.Errorf("failed to publish event to broker: %w", err)
	}

	return nil
}

// // publishOrderCreatedEvent publishes an event to notify other services
// func (uc *OrderUseCase) publishOrderCreatedEvent(ctx context.Context, order *domain.Order) error {
// 	event := event.OrderCreatedEvent{
// 		EventID:    uuid.New(),
// 		SageID:     order.SagaID,
// 		OrderID:    order.ID,
// 		CustomerID: order.CustomerID,
// 		Items:      order.Items,
// 		CreatedAt:  time.Now(),
// 	}

// 	return uc.eventPublisher.Publish(ctx, "order.created", event)
// }

// GetOrder retrieves an order by its ID
// func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
// 	if id == "" {
// 		return nil, domain.ErrInvalidOrderID
// 	}

// 	return uc.orderRepo.GetByID(ctx, id)
// }

// // UpdateOrderStatus updates the status of an order
// func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) error {
// 	if id == "" {
// 		return domain.ErrInvalidOrderID
// 	}

// 	order, err := uc.orderRepo.GetByID(ctx, id)
// 	if err != nil {
// 		return err
// 	}

// 	order.ChangeStatus(status)

// 	return uc.orderRepo.Update(ctx, order)
// }

// // AddOrderItem adds an item to an existing order
// func (uc *OrderUseCase) AddOrderItem(
// 	ctx context.Context,
// 	orderID string,
// 	productID string,
// 	quantity int32,
// 	price float64,
// ) error {
// 	if orderID == "" {
// 		return domain.ErrInvalidOrderID
// 	}

// 	if productID == "" {
// 		return domain.ErrInvalidProductID
// 	}

// 	if quantity <= 0 {
// 		return domain.ErrInvalidQuantity
// 	}

// 	if price <= 0 {
// 		return domain.ErrInvalidPrice
// 	}

// 	order, err := uc.orderRepo.GetByID(ctx, orderID)
// 	if err != nil {
// 		return err
// 	}

// 	order.AddItem(productID, quantity, price)

// 	return uc.orderRepo.Update(ctx, order)
// }

// // RemoveOrderItem removes an item from an order
// func (uc *OrderUseCase) RemoveOrderItem(ctx context.Context, orderID string, itemID string) error {
// 	if orderID == "" {
// 		return domain.ErrInvalidOrderID
// 	}

// 	if itemID == "" {
// 		return domain.ErrInvalidOrderID
// 	}

// 	order, err := uc.orderRepo.GetByID(ctx, orderID)
// 	if err != nil {
// 		return err
// 	}

// 	itemUUID, err := uuid.Parse(itemID)
// 	if err != nil {
// 		return domain.ErrInvalidOrderID
// 	}

// 	if !order.RemoveItem(itemUUID) {
// 		return domain.ErrOrderNotFound
// 	}

// 	if len(order.Items) == 0 {
// 		return domain.ErrEmptyOrderItems
// 	}

// 	return uc.orderRepo.Update(ctx, order)
// }

// // CancelOrder cancels an order
// func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) error {
// 	if id == "" {
// 		return domain.ErrInvalidOrderID
// 	}

// 	order, err := uc.orderRepo.GetByID(ctx, id)
// 	if err != nil {
// 		return err
// 	}

// 	order.ChangeStatus(domain.OrderStatusCancelled)

// 	return uc.orderRepo.Update(ctx, order)
// }

// func (uc *OrderUseCase) ListOrders(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
// 	if limit <= 0 {
// 		limit = 10
// 	}

// 	if offset < 0 {
// 		offset = 0
// 	}

// 	return uc.orderRepo.List(ctx, limit, offset)
// }
