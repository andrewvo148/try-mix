package dto

import (
	"order-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// Request DTOs

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	CustomerID string `json:"customer_id"`
	Items []CreateOrderItemRequest `json:"item"`
}

// CreateOrderItemRequest represents an item in the order creation request
type CreateOrderItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity int `json:"quantity"`
	Price float64 `json:"price"`
}

// UpdateOrderStatusRequest represents the request to update an order's status
type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

// Response DTOs

// OrderResponse represents the response format for an order
type OrderResponse struct {
	ID uuid.UUID
	CustomerID string
	Status string
	TotalPrice float64
	Items []OrderItemResponse
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrderItemResonse represents an item in the order reponse
type OrderItemResponse struct {
	ID uuid.UUID
	ProductID string
	Quantity int
	Price float64
}


// Conversion functions

// OrderToReponse converts a domain order model to response DTO
func OrderToResponse(order *domain.Order) OrderResponse {
	itemResponses := make([]OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		itemResponses = append(itemResponses, OrderItemResponse{
			ID: item.ID,
			ProductID: item.ProductID,
			Price: item.Price,
			Quantity: int(item.Quantity),
		})
	}

	return OrderResponse{
		ID: order.ID,
		CustomerID: order.CustomerID,
		Status: string(order.Status),
		TotalPrice: order.TotalPrice,
		Items: itemResponses,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}

// OrderLi