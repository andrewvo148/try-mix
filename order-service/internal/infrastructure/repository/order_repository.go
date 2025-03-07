package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"order-service/internal/app/ports"
	"order-service/internal/domain"
	"order-service/internal/infrastructure/sqlc"
	"strconv"

	"github.com/google/uuid"
)

// OrderRepository implements the OrderRepository interface using SQLC and PostgresSQL
type OrderRepository struct {
	queries *sqlc.Queries
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(tx *sql.Tx) ports.OrderRepository {
	return &OrderRepository{
		queries: sqlc.New(tx),
	}
}

// Create persists a new order to the database
func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {

	// Insert order
	err := r.queries.CreateOrder(ctx, sqlc.CreateOrderParams{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     string(order.Status),
		TotalPrice: fmt.Sprintf("%.2f", order.TotalPrice),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	})

	if err != nil {
		return err
	}

	// Insert order items
	for _, item := range order.Items {
		err = r.queries.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			ID:        item.ID,
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     fmt.Sprintf("%.2f", item.Price),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// GetByID retrieves an order by its ID
func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {

	// Validate UUID
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidOrderID
	}

	orderRow, err := r.queries.GetOrder(ctx, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}

		return nil, err
	}

	// Get order items
	items, err := r.queries.GetOrderItems(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// Map to domain model
	totalPriceFloat, err := strconv.ParseFloat(orderRow.TotalPrice, 64)
	if err != nil {
		return nil, err
	}

	order := &domain.Order{
		ID:         orderRow.ID,
		CustomerID: orderRow.CustomerID,
		Status:     domain.OrderStatus(orderRow.Status),
		TotalPrice: totalPriceFloat,
		CreatedAt:  orderRow.CreatedAt,
		UpdatedAt:  orderRow.UpdatedAt,
		Items:      make([]domain.OrderItem, 0, len(items)),
	}

	for _, item := range items {
		totalPriceFloat, err = strconv.ParseFloat(orderRow.TotalPrice, 64)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, domain.OrderItem{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     totalPriceFloat,
		})
	}

	return order, nil
}

// Update updates an existing order
func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	// Start a transaction

	// Update order
	err := r.queries.UpdateOrder(ctx, sqlc.UpdateOrderParams{
		Status:     string(order.Status),
		TotalPrice: fmt.Sprintf("%.2f", order.TotalPrice),
		UpdatedAt:  order.UpdatedAt,
		ID:         order.ID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	// Delete existing items
	err = r.queries.DeleteOrderItems(ctx, order.ID)
	if err != nil {
		return err
	}

	// Insert updated items
	for _, item := range order.Items {
		err = r.queries.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			ID:        item.ID,
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     fmt.Sprintf("%.2f", item.Price),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes an order
func (r *OrderRepository) Delete(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrInvalidOrderID
	}

	err = r.queries.DeleteOrder(ctx, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}

		return err
	}

	return nil
}

// List retrieves a paginated list of orders
func (r *OrderRepository) List(ctx context.Context, limit, offfset int) ([]*domain.Order, error) {
	// Get orders
	orderRows, err := r.queries.ListOrders(ctx, sqlc.ListOrdersParams{
		Limit:  int32(limit),
		Offset: int32(offfset),
	})

	if err != nil {
		return nil, err
	}

	// Map to domain model
	orders := make([]*domain.Order, 0, len(orderRows))
	for _, row := range orderRows {
		totalPrice, err := strconv.ParseFloat(row.TotalPrice, 64)
		if err != nil {
			return nil, err
		}

		order := &domain.Order{
			ID:         row.ID,
			CustomerID: row.CustomerID,
			Status:     domain.OrderStatus(row.Status),
			TotalPrice: totalPrice,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
			Items:      []domain.OrderItem{},
		}

		// Get items of this order
		items, err := r.queries.GetOrderItems(ctx, row.ID)
		if err != nil {
			continue
		}

		for _, item := range items {
			price, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				return nil, err
			}
			order.Items = append(order.Items, domain.OrderItem{
				ID:        item.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     price,
			})
		}
		orders = append(orders, order)
	}

	return orders, nil
}
