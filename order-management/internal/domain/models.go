package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// Event represents a generic event in the system
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
	CreatedAt  time.Time `json:"created_at"`
}

// Product represents a product entity in an order
type Product struct {
	ID       string  `json:"id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// Payment represents a payment entity
type Payment struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Producer struct {
	producer EventProducer
}

func (p *Producer) PublishEvent(topic string, eventType string, data interface{}) error {
	// Marsha the data to JSON
	eventData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Create and populate the event

	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      eventData,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish the event
	if err := p.producer.Publish("events", []byte(eventType), eventBytes); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
