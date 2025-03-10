package domain

import (
	"time"

	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusProcessed OutboxStatus = "PROCESSED"
	OutboxStatusFailed    OutboxStatus = "FAILED"
)

// OutputMessage represents a message in the outbox queue
type OutboxMessage struct {
	ID           uuid.UUID
	AggregateID  uuid.UUID
	EventType    string
	Payload      []byte
	Status       OutboxStatus
	AttemptCount int
	CreatedAt    time.Time
	ProcessedAt  time.Time
}
