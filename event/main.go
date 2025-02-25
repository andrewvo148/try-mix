package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// EventType defines the event type of events in our system
type EventType string

const (
	UserCreated EventType = "user.created"
	UserUpdated EventType = "user.updated"
	UserDeleted EventType = "user.deleted"
	OrderPlaced EventType = "order.placed"
)

// Event is the base event structure
type Event struct {
	ID          string      `json:"id"`
	Type        EventType   `json:"type"`
	Source      string      `json:"source"`
	Timestamp   time.Time   `json:"timestamp"`
	Data        interface{} `json:"data"`
	SpecVersion string      `json:"spec_version"`
}

// UserCreatedEvent represents an event of a user being created
type UserCreatedEvent struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UserUpdatedEvent represents an event of a user being updated
type UserUpdatedEvent struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// UserDeletedEvent represents an event of a user being deleted
type UserDeletedEvent struct {
	UserID string `json:"user_id"`
}

// OrderPlacedEvent represents an event of an order being placed
type OrderPlacedEvent struct {
	OrderID     string   `json:"order_id"`
	UserID      string   `json:"user_id"`
	Products    []string `json:"products"`
	TotalAmount float64  `json:"total_amount"`
}

// EventProducer handles event production
type EventProducer struct {
	producer sarama.SyncProducer
	topic    string
	source   string
}

// NewEventProducer creates a new event producer
func NewEventProducer(brokers []string, topic string, source string) (*EventProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {

		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &EventProducer{producer: producer, topic: topic, source: source}, nil
}

// Emit publishes an event to Kafka
func (p *EventProducer) Emit(ctx context.Context, eventType EventType, data interface{}) error {
	event := Event{
		ID:          uuid.New().String(),
		Type:        eventType,
		Source:      p.source,
		Timestamp:   time.Now(),
		Data:        data,
		SpecVersion: "1.0.0",
	}

	eventJson, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.ID),
		Value: sarama.ByteEncoder(eventJson),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	log.Printf("Event %s sent to partition %d at offset %d", event.ID, partition, offset)

	return nil
}

// Close closes the producer
func (p *EventProducer) Close() error {
	return p.producer.Close()
}

// EventConsumer handles event consumption
type EventConsumer struct {
	consumer  sarama.ConsumerGroup
	topics    []string
	handler   map[EventType]EventHandler
	cancelled bool
	mu        sync.Mutex
}

// EventHandler is the interface for handling events
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
}

// EventHandlerFunc is a function type that implements the EventHandler interface
type EventHandlerFunc func(ctx context.Context, event Event) error

// Handle implements the EventHandler interface
func (f EventHandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(brokers []string, groupID string, topics []string) (*EventConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &EventConsumer{
		consumer: consumer,
		topics:   topics,
		handler:  make(map[EventType]EventHandler),
	}, nil
}

// RegisterEventHandler registers a handler for a specific event type
func (c *EventConsumer) RegisterHandler(eventType EventType, handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handler[eventType] = handler
}

// Start beging consuming events from Kafka
func (c *EventConsumer) Start(ctx context.Context) error {
	handler := &consumerGroupHandler{
		consumer: c,
		ready:    make(chan bool),
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			// check if we need to stop
			c.mu.Lock()
			if c.cancelled {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()

			// consume events
			if err := c.consumer.Consume(ctx, c.topics, handler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Printf("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-handler.ready
	log.Println("Consumer up and running")

	select {
	case <-ctx.Done():
		log.Println("Context cancelled, stopping consumer")
		c.Stop()
	}

	wg.Wait()
	return nil
}

// Stop stops the consumer
func (c *EventConsumer) Stop() {
	c.mu.Lock()
	c.cancelled = true
	c.mu.Unlock()
	c.consumer.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer *EventConsumer
	ready    chan bool
}

// Setup is run at the beginning of a new session
func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

// Cleanup is run at the end of a session, either successfully or unsuccessfully
func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim handles the consumption of messages
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		event := Event{}
		err := json.Unmarshal(msg.Value, &event)
		if err != nil {
			log.Printf("Error unmarshalling event: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		log.Printf("Received event: %s, Type: %s", event.ID, event.Type)

		if handler, ok := h.consumer.handler[event.Type]; ok {
			if err = handler.Handle(session.Context(), event); err != nil {
				log.Printf("Error handling event: %v", err)
			}
		} else {
			log.Printf("No handler registered for event type: %s", event.Type)
		}

		// mark the message as processed
		session.MarkMessage(msg, "")
	}

	return nil
}

func main() {
	// Configuration
	brokers := []string{"localhost:29092"}
	topic := "bussiness-events"
	consumerGroup := "user-service"

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Create a new event producer
	producer, err := NewEventProducer(brokers, topic, "user-service")
	if err != nil {
		log.Fatalf("Failed to create event producer: %v", err)
	}
	defer producer.Close()

	go func() {
		time.Sleep(2 * time.Second)

		userData := UserCreatedEvent{
			UserID:    uuid.New().String(),
			Email:     "john.doe@test.com",
			FirstName: "John",
			LastName:  "Doe",
		}

		if err := producer.Emit(ctx, UserCreated, userData); err != nil {
			log.Printf("Error to emit event: %v", err)
		}

		time.Sleep(2 * time.Second)
		orderData := OrderPlacedEvent{
			OrderID:     uuid.New().String(),
			UserID:      userData.UserID,
			Products:    []string{"product-1", "product-2"},
			TotalAmount: 99.99,
		}
		if err := producer.Emit(ctx, OrderPlaced, orderData); err != nil {
			log.Printf("Error emitting event: %v", err)
		}
	}()

	// Create consumer

	// Create consumer
	consumer, err := NewEventConsumer(brokers, consumerGroup, []string{topic})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	// Register handlers
	consumer.RegisterHandler(UserCreated, EventHandlerFunc(func(ctx context.Context, event Event) error {
		var userData UserCreatedEvent
		if data, err := json.Marshal(event.Data); err == nil {
			if err := json.Unmarshal(data, &userData); err != nil {
				return fmt.Errorf("error unmarshaling user data: %w", err)
			}
			log.Printf("User created: %s (%s %s)", userData.UserID, userData.FirstName, userData.LastName)
			return nil
		}
		return fmt.Errorf("error processing user data")
	}))

	consumer.RegisterHandler(OrderPlaced, EventHandlerFunc(func(ctx context.Context, event Event) error {
		var orderData OrderPlacedEvent
		if data, err := json.Marshal(event.Data); err == nil {
			if err := json.Unmarshal(data, &orderData); err != nil {
				return fmt.Errorf("error unmarshaling order data: %w", err)
			}
			log.Printf("Order placed: %s for user %s with total amount %.2f",
				orderData.OrderID, orderData.UserID, orderData.TotalAmount)
			return nil
		}
		return fmt.Errorf("error processing order data")
	}))

	// Start consumer in a goroutine
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatalf("Failed to start consumer: %v", err)
		}
	}()

	// Wait for termination signal
	<-signals
	log.Println("Shutting down...")
	cancel()
	consumer.Stop()

}
