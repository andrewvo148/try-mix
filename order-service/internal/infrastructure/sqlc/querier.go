// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	// db/queries.sql
	CreateOrder(ctx context.Context, arg CreateOrderParams) error
	CreateOrderItem(ctx context.Context, arg CreateOrderItemParams) error
	DeleteOrder(ctx context.Context, id uuid.UUID) error
	DeleteOrderItem(ctx context.Context, arg DeleteOrderItemParams) error
	DeleteOrderItems(ctx context.Context, orderID uuid.UUID) error
	GetOrder(ctx context.Context, id uuid.UUID) (Order, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error)
	ListOrders(ctx context.Context, arg ListOrdersParams) ([]Order, error)
	UpdateOrder(ctx context.Context, arg UpdateOrderParams) error
}

var _ Querier = (*Queries)(nil)
