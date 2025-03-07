package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"order-service/internal/app/ports"
	"order-service/internal/domain"
	"order-service/internal/infrastructure/sqlc"
	"time"

	"github.com/google/uuid"
)

// OutboxRepository implements the ports.OutboxRepository interface
type OutboxRepository struct {
	queries *sqlc.Queries
}

func NewOutboxRepository(tx *sql.Tx) ports.OutboxRepository {
	return &OutboxRepository{
		queries: sqlc.New(tx),
	}
}

// CreateMessage implements ports.OutboxRepository.
func (o *OutboxRepository) CreateMessage(ctx context.Context, aggregateID string, messageType string, payload interface{}) error {

	// Marshal the payload into JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	params := sqlc.CreateOutboxMessageParams{
		ID:          uuid.New(),
		AggregateID: aggregateID,
		EventType:   messageType,
		Payload:     jsonData, // Assuming this is already a json.RawMessage or similar
		CreatedAt:   time.Now(),
		Status:      "pending", // Default status as defined in your schema
	}

	return o.queries.CreateOutboxMessage(ctx, params)
}

// GetPendingMessages implements ports.OutboxRepository.
func (o *OutboxRepository) GetPendingMessages(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	messages, err := o.queries.GetPendingOutboxMessages(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	result := make([]domain.OutboxMessage, len(messages))
	for i, msg := range messages {
		result[i] = domain.OutboxMessage{
			ID:          msg.ID,
			AggregateID: msg.AggregateID,
			Type:        msg.EventType,
			Payload:     msg.Payload,
			CreatedAt:   msg.CreatedAt,
			ProcessedAt: msg.ProcessedAt.Time,
			Status:      domain.OutboxStatus(msg.Status),
		}
	}

	return result, nil
}

// MarkMessageAsFailed implements ports.OutboxRepository.
func (o *OutboxRepository) MarkMessageAsFailed(ctx context.Context, messageID string, reason string) error {
	panic("unimplemented")
}

// MarkMessageAsProcessed implements ports.OutboxRepository.
func (o *OutboxRepository) MarkMessageAsProcessed(ctx context.Context, messageID string) error {
	return o.queries.MarkOutboxMessageProcessed(ctx, sqlc.MarkOutboxMessageProcessedParams{
		ID:          uuid.MustParse(messageID),
		ProcessedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})
}
