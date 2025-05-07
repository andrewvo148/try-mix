package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"microservice-demo/internal/order-service/app/ports"
	"microservice-demo/internal/order-service/infrastructure/config"

	"time"

	"github.com/IBM/sarama"
)

// KafkaConfig holds configuration for Kafka connection
// type KafkaConfig struct {
// 	Brokers           []string
// 	ClientID          string
// 	ProducerRetry     int
// 	RequiredAcks      sarama.RequiredAcks
// 	FlushFrequency    time.Duration
// 	FlushMessage      int
// 	ConnectionTimeout time.Duration
// }

// DefaultConfig provides sensible default configuration
// func DefaultConfig() *KafkaConfig {
// 	return &KafkaConfig{
// 		Brokers:           []string{"localhost:9092"},
// 		ClientID:          "order-service",
// 		ProducerRetry:     3,
// 		RequiredAcks:      sarama.WaitForLocal,
// 		FlushFrequency:    500 * time.Millisecond,
// 		FlushMessage:      1000,
// 		ConnectionTimeout: 15 * time.Second,
// 	}
// }

// EventPublisher implements the ports.EventPublisher interface using Sarama
type EventPublisher struct {
	producer sarama.AsyncProducer
	config   config.KafkaConfig
}

// NewEventPublisher creates a new EventPublisher with configuration
func NewEventPublisher(cfg config.KafkaConfig) (*EventPublisher, error) {
	// Create Sarama configuration
	saramaConfig := sarama.NewConfig()

	// Apply configuration from our unified config
	saramaConfig.ClientID = "order-service" // You may want to add this to your config
	saramaConfig.Producer.Retry.Max = cfg.Producer.RetryMax

	// Map our RequiredAcks string to Sarama's RequiredAcks type
	switch cfg.Producer.RequiredAcks {
	case "none":
		saramaConfig.Producer.RequiredAcks = sarama.NoResponse
	case "leader":
		saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal
	case "all":
		saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	default:
		saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal
	}

	// Set additional producer configurations
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Retry.Backoff = cfg.Producer.RetryBackoff
	saramaConfig.Producer.Timeout = cfg.Producer.MessageTimeout

	// Default flush settings (you might want to add these to your config)
	saramaConfig.Producer.Flush.Frequency = 500 * time.Millisecond
	saramaConfig.Producer.Flush.Messages = 1000

	// Connection timeout
	saramaConfig.Net.DialTimeout = cfg.ConnectionTimeout

	// Apply security settings if enabled
	if cfg.Security.Enabled {
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.Mechanism = sarama.SASLMechanism(cfg.Security.SaslMechanism)
		saramaConfig.Net.SASL.User = cfg.Security.SaslUsername
		saramaConfig.Net.SASL.Password = cfg.Security.SaslPassword

		switch cfg.Security.Protocol {
		case "ssl":
			saramaConfig.Net.TLS.Enable = true
		case "sasl_ssl":
			saramaConfig.Net.TLS.Enable = true
			saramaConfig.Net.SASL.Enable = true
		case "sasl_plaintext":
			saramaConfig.Net.SASL.Enable = true
		}
	}

	// Create a new async producer
	producer, err := sarama.NewAsyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Start a goroutine to handle success and error responses
	go func() {
		for {
			select {
			case success := <-producer.Successes():
				// You might want to log this or emit metrics
				_ = success
			case err := <-producer.Errors():
				// Handle errors (ideally with proper error handling/logging)
				fmt.Printf("Failed to publish message: %v\n", err)
			}
		}
	}()

	// Return the EventPublisher instance with the configured producer
	return &EventPublisher{
		producer: producer,
		config:   cfg,
	}, nil

}

// Publish publishes an event to the specified Kafka topic
func (p *EventPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
	// Marshal the event payload into JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create a new Kafka message
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
	}

	// Add any headers from context if needed
	// This could include tracing IDs or other metadata
	traceID := ctx.Value("traceID")
	if traceID != nil {
		if tracIDstr, ok := traceID.(string); ok {
			msg.Headers = append(msg.Headers, sarama.RecordHeader{
				Key:   []byte("trace-id"),
				Value: []byte(tracIDstr),
			})
		}
	}

	// Send the message
	select {
	case p.producer.Input() <- msg:
		// Message accepted
		return nil
	case <-ctx.Done():
		// Context cancelled
		return ctx.Err()
	}
}

// Close closes the Kafka producer
func (p *EventPublisher) Close() error {
	return p.producer.Close()
}

var _ ports.EventPublisher = (*EventPublisher)(nil)
