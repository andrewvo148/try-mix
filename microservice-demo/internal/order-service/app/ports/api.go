package ports

import (
	"context"
	"microservice-demo/internal/order-service/domain"
)

type OrderUseCase interface {
	CreateOrder(ctx context.Context, customerID string, items []domain.OrderItem) (*domain.Order, error)
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status domain.OrderStatus) error
	AddOrderItem(ctx context.Context, orderID string, productID string, quantity int32, price float64) error
	RemoveOrderItem(ctx context.Context, orderID string, itemID string) error
	CancelOrder(ctx context.Context, id string) error
	ListOrders(ctx context.Context, limit, offset int) ([]*domain.Order, error)
}
