package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	order.SagaID = uuid.New()

	// Calculate total value for the event

	// Prepare order created event
	orderCreatedEvent := event.OrderCreatedEvent{
		EventID:    uuid.New(),
		SagaID:     order.SagaID,
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		Items:      order.Items,
		TotalPrice: order.CalculateTotalPrice(),
		CreatedAt:  time.Now(),
	}

	// Marshal event payload
	eventPayload, err := json.Marshal(orderCreatedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order created event: %w", err)
	}

	// Begin transaction
	txCtx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is either committed or rolled back
	defer func() {
		if err != nil {
			// Rollback on error
			if rbErr := uc.uow.Rollback(txCtx); rbErr != nil {
				// Log rollback error but don't override original error
				log.Printf("rollback error: %v", rbErr)
			}
		}
	}()

	// Create order
	if err = uc.uow.Orders(txCtx).Create(txCtx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create outbox message
	if err = uc.uow.OutboxMessages(txCtx).CreateMessage(
		txCtx,
		order.ID.String(),
		"order.created",
		eventPayload,
	); err != nil {
		return nil, fmt.Errorf("failed to create outbox message: %w", err)
	}

	// Commit transaction
	if err = uc.uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
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

// // GetOrder retrieves an order by its ID
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
