package domain

import (
	"encoding/json"
	"time"
)


// Command topics and types
const (
	CommandTopic = "commands"

	CreateOrderCommandType = "CreateOrder"
	SubmitPaymentCommandType = "SubmitPayment"
	ReserveProductsCommandType = "ReserveProducts"
	ShipProductsCommandType = "ShipProducts"
)
// Command represents a command in the system
type Command struct {
	Type      string
	Timestamp time.Time
	Data      json.RawMessage
}

// CreateOrderCommand represents a command to create a new order
type CreateOrderCommand struct {
	CustomerID string `json:"customer_id"`
	Products   []Product `json:"products"`
}

type SubmitPaymentCommand struct {
	OrderID string    `json:"order_id"`
	Amount  float64   `json:"amount"`
}

type ReserveProductsCommand struct {
	OrderID string `json:"order_id"`
	Products   []Product `json:"products"`
}
type ShipProductsCommand struct {
	OrderID string `json:"order_id"`
}

