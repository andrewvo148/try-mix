package main

import (
	"encoding/json"
	"time"
)

const (
	OrderCreatedEventType     = "OrderCreated"
	PaymentApprovedEventType  = "PaymentApproved"
	PaymentDeclinedEventType  = "PaymentDeclined"
	ProductsReservedEventType = "ProductsReserved"
	ProductsShippedEventType  = "ProductsShipped"
)

// Kafka topics
const (
	OrdersTopic    = "orders"
	PaymentsTopic  = "payments"
	InventoryTopic = "inventory"
	ShippingTopic  = "shipping"
)

// Event represents a domain event
type Event struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// Order represents an order entity
type Order struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Status     string    `json:"status"`
	Products   []Product `json:"products"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Product represents a product entity in an order
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type producer struct {
}
type OrderService struct {
	orders    map[string]Order
	producter producer
}
