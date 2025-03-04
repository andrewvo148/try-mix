package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusFailed OrderStatus = "FAILED"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusDelivered OrderStatus = "DELIVERED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID         uuid.UUID
	CustomerID string
	Items      []OrderItem
	TotalPrice float64
	Status     OrderStatus
	SagaID     uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderItem struct {
	ID        uuid.UUID
	ProductID string
	Quantity  int32
	Price     float64
}

type OrderEvent struct {
	EventID     string      `json:"event_id"`
	EventType   string      `json:"event_type"`
	OrderNumber string      `json:"order_number"`
	CustomerID  string      `json:"customer_id"`
	Timestamp   time.Time   `json:"timestamp"`
	Data        interface{} `json:"data"`
}

// NewOrder creates a new order with the given details
func NewOrder(customerID string, items []OrderItem) *Order {
	orderID := uuid.New()
	now := time.Now()

	var totalPrice float64
	for _, item := range items {
		totalPrice += item.Price * float64(item.Quantity)
	}

	return &Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     OrderStatusPending,
		SagaID: uuid.New(),
		TotalPrice: totalPrice,
		Items:      items,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// ChangeStatus updates the status of an order
func (o *Order) ChangeStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}

// AddItem adds an item to the order and recalculates the total price
func (o *Order) AddItem(productID string, quantity int32, price float64) {
	item := OrderItem{
		ID:        uuid.New(),
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
	}

	o.Items = append(o.Items, item)
	o.TotalPrice += price * float64(quantity)
	o.UpdatedAt = time.Now()
}

// RemoveItem removes an item from the order by its ID
func (o *Order) RemoveItem(itemID uuid.UUID) bool {
	for i, item := range o.Items {
		if item.ID == itemID {
			o.TotalPrice -= item.Price * float64(item.Quantity)
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			o.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}
