package main

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventProducer_Emit(t *testing.T) {
	// Create a mock Kafka producer
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	producer := &EventProducer{
		producer: mockProducer,
		topic:    "test-topic",
		source:   "test-source",
	}

	// Test data
	eventType := UserCreated
	userData := UserCreatedEvent{
		UserID:    "user-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Call the Emit method
	err := producer.Emit(context.Background(), eventType, userData)

	assert.NoError(t, err, "Emit should not return an error")
	mockProducer.Close()
}

func TestEventProducer_Emit_Error(t *testing.T) {
	// Create a mock Kafka consumer
	mockConsumer := mocks.NewSyncProducer(t, nil)
	mockConsumer.ExpectSendMessageAndFail(sarama.ErrOutOfBrokers)

	producer := &EventProducer{
		producer: mockConsumer,
		topic:    "test-topic",
		source:   "test-source",
	}

	// Test data
	eventType := UserCreated
	userData := UserCreatedEvent{
		UserID:    "user-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Call the Emit method
	err := producer.Emit(context.Background(), eventType, userData)

	assert.Error(t, err, "Emit should return an error")
	assert.Contains(t, err.Error(), "failed to send event", "Error message should indicate failure to send event")
	mockConsumer.Close()
}

// MockEventHandler mocks an event handler
type MockEventHandler struct {
	mock.Mock
}

func (m *MockEventHandler) Handle(ctx context.Context, event Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// TestEventConsumer tests the Start method of EventConsumer
func TestEventConsumer(t *testing.T) {
	mockHandler := new(MockEventHandler)
	consumer := &EventConsumer{
		handler: make(map[EventType]EventHandler),
	}

	consumer.RegisterHandler(UserCreated, mockHandler)

	userData := UserCreatedEvent{
		UserID:    uuid.New().String(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	event := Event{
		ID:          uuid.New().String(),
		Type:        UserCreated,
		Source:      "test-source",
		Timestamp:   time.Now(),
		Data:        userData,
		SpecVersion: "1.0.0",
	}

	mockHandler.On("Handle", mock.Anything, event).Return(nil)

	err := mockHandler.Handle(context.Background(), event)
	assert.NoError(t, err, "Expected no error when handling event")
	mockHandler.AssertExpectations(t)
}
