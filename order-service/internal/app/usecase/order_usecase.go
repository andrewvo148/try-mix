package usecase

import (
	"context"
	"order-service/internal/app/ports"
	"order-service/internal/domain"
	"order-service/internal/event"
	"time"

	"github.com/google/uuid"
)

// OrderUsecase implements the order business logic
type OrderUseCase struct {
	orderRepo      ports.OrderRepository
	eventPublisher ports.Publisher
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(
	orderRepo ports.OrderRepository,
	eventPublisher ports.Publisher,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:      orderRepo,
		eventPublisher: eventPublisher,
	}
}

// CreateOrder creates a new order with the given details
func (uc *OrderUseCase) CreateOrder(ctx context.Context, customerID string, items []domain.OrderItem) (*domain.Order, error) {
	if customerID == "" {
		return nil, domain.ErrInvalidCustomerID
	}

	if len(items) == 0 {
		return nil, domain.ErrEmptyOrderItems
	}

	order := domain.NewOrder(customerID, items)

	err := uc.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	// Start the saga by publish OrderCreated event
	err = uc.publishOrderCreatedEvent(ctx, order)
	if err != nil {
		// Failed to publish event - compensate by marking order as failed
		order.Status = domain.OrderStatusFailed
		uc.orderRepo.Update(ctx, order)
		return nil, err
	}

	return order, nil
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
