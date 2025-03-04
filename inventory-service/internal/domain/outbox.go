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
	ID          uuid.UUID
	EventType   string
	Payload     interface{}
	Status      OutboxStatus
	CreatedAt   time.Time
	ProcessedAt time.Time
	FailReason  string `json:"failReason,omitempty"`
	RetryCount  int    `json:"retryCount"`
	MaxRetries  int    `json:"maxRetries"`
}
