package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/app/ports"
	"order-service/internal/domain"
	"order-service/internal/events"
	"time"
)

// OutboxProcessor handles processing and publishing of outbox messages
type OutboxProcessor struct {
	outboxRepo      ports.OutboxRepository
	eventPublisher  ports.EventPublisher
	batchSize       int
	processInterval time.Duration
	maxRetries      int
}

// NewOutboxProcessor creates a new outbox processor
func NewOutboxProcessor(
	outboxRepo ports.OutboxRepository,
	eventPublisher ports.EventPublisher,
	batchSize int,
	processInterval time.Duration,
	maxRetries int,
) *OutboxProcessor {
	return &OutboxProcessor{
		outboxRepo:      outboxRepo,
		eventPublisher:  eventPublisher,
		batchSize:       batchSize,
		processInterval: processInterval,
		maxRetries:      maxRetries,
	}
}

// Start begins the outbox processing loop
func (p *OutboxProcessor) Start(ctx context.Context) error {
	log.Println("Starting outbox processor...")

	ticker := time.NewTicker(p.processInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Outbox processor stopping due to context cancellation")
			return ctx.Err()
		case <-ticker.C:
			if err := p.processOutboxMessages(ctx); err != nil {
				log.Printf("Error processing outbox messages: %v", err)
				// Continue processing on next tick
			}
		}
	}
}

// processOutboxMessages fetches and processes pending outbox messages
func (p *OutboxProcessor) processOutboxMessages(ctx context.Context) error {
	// Get unprocessed messages
	messages, err := p.outboxRepo.GetPendingMessages(ctx, p.batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending outbox messages: %w", err)
	}

	if len(messages) == 0 {
		// No messages to process
		return nil
	}

	log.Printf("Processing %d outbox messages", len(messages))

	for _, msg := range messages {
		// Process each message in its own context to isolate failures
		err := p.processMessage(ctx, msg)
		if err != nil {
			log.Printf("Failed to process message %s: %v", msg.ID, err)
			// Update message with failed attempt
			updateErr := p.outboxRepo.IncrementAttempt(ctx, msg.ID)
			if updateErr != nil {
				log.Printf("Failed to update attempt count for message %s: %v", msg.ID, updateErr)
			}

			// Check if max retries reached
			if msg.AttemptCount+1 >= p.maxRetries {
				log.Printf("Message %s reached max retry count, marking as failed", msg.ID)
				deadLetterErr := p.outboxRepo.MarkMessageAsFailed(ctx, msg.ID, err.Error())
				if deadLetterErr != nil {
					log.Printf("Failed to mark message as failed: %v", deadLetterErr)
				}
			}

			// Continue with next message
			continue
		}

		// Mark as processed after successful publishing
		if err := p.outboxRepo.MarkMessageAsProcessed(ctx, msg.ID); err != nil {
			log.Printf("Failed to mark message %s as processed: %v", msg.ID, err)
			// Continue with next message
		}
	}

	return nil
}

// processMessage processes a single outbox message
func (p *OutboxProcessor) processMessage(ctx context.Context, msg domain.OutboxMessage) error {
	// Handle different event types
	switch msg.EventType {
	case "order.created":
		var event events.OrderCreatedEvent
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			return fmt.Errorf("failed to unmarshal order.created event: %w", err)
		}

		// Publish to the message broker
		if err := p.eventPublisher.Publish(ctx, msg.EventType, event); err != nil {
			return fmt.Errorf("failed to publish event to broker: %w", err)
		}

	// case "order.updated":
	// 	var event domain.OrderUpdatedEvent
	// 	if err := json.Unmarshal(msg.Payload, &event); err != nil {
	// 		return fmt.Errorf("failed to unmarshal order.updated event: %w", err)
	// 	}

	// 	// Publish to the message broker
	// 	if err := p.eventPublisher.Publish(ctx, msg.EventType, event); err != nil {
	// 		return fmt.Errorf("failed to publish event to broker: %w", err)
	// 	}

	// Add more event types as needed

	default:
		log.Printf("Unknown event type: %s", msg.EventType)
		// We'll still mark it as processed since we can't do anything with it
		return nil
	}

	return nil
}
