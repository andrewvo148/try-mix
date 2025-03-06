package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"order-service/internal/app/ports"
	"order-service/internal/domain"
	"order-service/internal/event"
	"time"

	"github.com/google/uuid"
)

// OrderUsecase implements the order business logic
type OrderUseCase struct {
	uow            ports.UnitOfWork
	eventPublisher ports.EventPublisher
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(
	uow ports.UnitOfWork,
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

	// Create order
	order := domain.NewOrder(customerID, items)
	order.SagaID = uuid.New()

	// Calculate total price for the event
	totalPrice := order.CalculateTotalPrice()

	// Prepare order created event for outbox
	orderCreatedEvent := event.OrderCreatedEvent{
		EventID:    uuid.New(),
		SageID:     order.SagaID,
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		Items:      order.Items,
		TotalPrice: totalPrice,
		CreatedAt:  time.Now(),
	}

	// Marshal event payload
	eventPayload, err := json.Marshal(orderCreatedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order created event: %w", err)
	}

	// Execute operations within a single unit of work
	err = uc.uow.Execute(ctx, func() error {
		// Get repositories from the unit of work
		orderRepo := uc.uow.Orders()
		outboxRepo := uc.uow.OutboxMessages()

		// Create order
		if err := orderRepo.Create(ctx, order); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// Create outbox message
		if err := outboxRepo.CreateMessage(
			ctx,
			order.ID.String(),
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

	return order, nil

	// Start the saga by publish OrderCreated event
	// err = uc.publishOrderCreatedEvent(ctx, order)
	// if err != nil {
	// 	// Failed to publish event - compensate by marking order as failed
	// 	order.Status = domain.OrderStatusFailed
	// 	uc.orderRepo.Update(ctx, order)
	// 	return nil, err
	// }

	
}

// publishOrderCreatedEvent publishes an event to notify other services
func (uc *OrderUseCase) publishOrderCreatedEvent(ctx context.Context, order *domain.Order) error {
	event := event.OrderCreatedEvent{
		EventID:    uuid.New(),
		SageID:     order.SagaID,
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		Items:      order.Items,
		CreatedAt:  time.Now(),
	}

	return uc.eventPublisher.Publish(ctx, "order.created", event)
}

// GetOrder retrieves an order by its ID
func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	if id == "" {
		return nil, domain.ErrInvalidOrderID
	}

	return uc.orderRepo.GetByID(ctx, id)
}

// UpdateOrderStatus updates the status of an order
func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	if id == "" {
		return domain.ErrInvalidOrderID
	}

	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	order.ChangeStatus(status)

	return uc.orderRepo.Update(ctx, order)
}

// AddOrderItem adds an item to an existing order
func (uc *OrderUseCase) AddOrderItem(
	ctx context.Context,
	orderID string,
	productID string,
	quantity int32,
	price float64,
) error {
	if orderID == "" {
		return domain.ErrInvalidOrderID
	}

	if productID == "" {
		return domain.ErrInvalidProductID
	}

	if quantity <= 0 {
		return domain.ErrInvalidQuantity
	}

	if price <= 0 {
		return domain.ErrInvalidPrice
	}

	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	order.AddItem(productID, quantity, price)

	return uc.orderRepo.Update(ctx, order)
}

// RemoveOrderItem removes an item from an order
func (uc *OrderUseCase) RemoveOrderItem(ctx context.Context, orderID string, itemID string) error {
	if orderID == "" {
		return domain.ErrInvalidOrderID
	}

	if itemID == "" {
		return domain.ErrInvalidOrderID
	}

	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	itemUUID, err := uuid.Parse(itemID)
	if err != nil {
		return domain.ErrInvalidOrderID
	}

	if !order.RemoveItem(itemUUID) {
		return domain.ErrOrderNotFound
	}

	if len(order.Items) == 0 {
		return domain.ErrEmptyOrderItems
	}

	return uc.orderRepo.Update(ctx, order)
}

// CancelOrder cancels an order
func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInvalidOrderID
	}

	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	order.ChangeStatus(domain.OrderStatusCancelled)

	return uc.orderRepo.Update(ctx, order)
}

func (uc *OrderUseCase) ListOrders(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	return uc.orderRepo.List(ctx, limit, offset)
}
