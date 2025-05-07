package usecase_test

import (
	"context"
	"database/sql"
	"errors"
	"microservice-demo/internal/order-service/app/usecase"
	"microservice-demo/internal/order-service/domain"

	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations of the dependencies
type MockUnitOfWork struct {
	mock.Mock
}

func (m *MockUnitOfWork) Execute(ctx context.Context, fn func(*sql.Tx) error) error {
	args := m.Called(ctx, mock.AnythingOfType("func(*sql.Tx) error"))
	// Execute the function with nil to simulate the transaction
	if args.Get(0) == nil {
		_ = fn(nil)
	}
	return args.Error(0)
}

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) WithTx(tx *sql.Tx) ports.OrderRepository {
	args := m.Called(tx)
	return args.Get(0).(ports.OrderRepository)
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

type MockOutboxRepository struct {
	mock.Mock
}

func (m *MockOutboxRepository) WithTx(tx *sql.Tx) ports.OutboxRepository {
	args := m.Called(tx)
	return args.Get(0).(ports.OutboxRepository)
}

func (m *MockOutboxRepository) CreateMessage(ctx context.Context, aggregateID uuid.UUID, eventType string, payload interface{}) error {
	args := m.Called(ctx, aggregateID, eventType, payload)
	return args.Error(0)
}

func (m *MockOutboxRepository) GetPendingMessages(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]domain.OutboxMessage), args.Error(1)
}

func (m *MockOutboxRepository) MarkMessageAsProcessed(ctx context.Context, messageID uuid.UUID) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockOutboxRepository) MarkMessageAsFailed(ctx context.Context, messageID uuid.UUID, reason string) error {
	args := m.Called(ctx, messageID, reason)
	return args.Error(0)
}

func (m *MockOutboxRepository) IncrementAttempt(ctx context.Context, messageID uuid.UUID) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
	args := m.Called(ctx, topic, event)
	return args.Error(0)
}

func TestCreateOrder_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Create mocks
	mockUow := new(MockUnitOfWork)
	mockOrderRepo := new(MockOrderRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	mockEventPublisher := new(MockEventPublisher)

	// Setup mock expectations
	mockUow.On("Execute", ctx, mock.AnythingOfType("func(*sql.Tx) error")).Return(nil)
	mockOrderRepo.On("WithTx", mock.Anything).Return(mockOrderRepo)
	mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	mockOutboxRepo.On("WithTx", mock.Anything).Return(mockOutboxRepo)
	mockOutboxRepo.On("CreateMessage", ctx, mock.AnythingOfType("uuid.UUID"), "order.created", mock.AnythingOfType("[]uint8")).Return(nil)
	mockEventPublisher.On("Publish", ctx, "order.created", mock.AnythingOfType("event.OrderCreatedEvent")).Return(nil)

	// Create the use case
	orderUseCase := usecase.NewOrderUseCase(mockUow, mockOrderRepo, mockOutboxRepo, mockEventPublisher)

	// Test data
	customerID := "customer123"
	items := []domain.OrderItem{
		{
			ProductID: "product1",
			Quantity:  2,
			Price:     100.00,
		},
		{
			ProductID: "product2",
			Quantity:  1,
			Price:     50.00,
		},
	}

	// Act
	order, err := orderUseCase.CreateOrder(ctx, customerID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, customerID, order.CustomerID)
	assert.Equal(t, domain.OrderStatusPending, order.Status)
	assert.Len(t, order.Items, 2)
	assert.Equal(t, float64(250.00), order.TotalPrice) // 2*100 + 1*50
	assert.NotEmpty(t, order.ID)
	assert.NotEmpty(t, order.SagaID)

	// Verify the mocks were called as expected
	mockUow.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestCreateOrder_ValidationFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Create mocks
	mockUow := new(MockUnitOfWork)
	mockOrderRepo := new(MockOrderRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	mockEventPublisher := new(MockEventPublisher)

	// Create the use case
	orderUseCase := usecase.NewOrderUseCase(mockUow, mockOrderRepo, mockOutboxRepo, mockEventPublisher)

	// Test cases for validation failures
	testCases := []struct {
		name        string
		customerID  string
		items       []domain.OrderItem
		expectedErr error
	}{
		{
			name:        "Empty CustomerID",
			customerID:  "",
			items:       []domain.OrderItem{{ProductID: "product1", Quantity: 1, Price: 100.00}},
			expectedErr: domain.ErrInvalidCustomerID,
		},
		{
			name:        "Empty Items",
			customerID:  "customer123",
			items:       []domain.OrderItem{},
			expectedErr: domain.ErrEmptyOrderItems,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			order, err := orderUseCase.CreateOrder(ctx, tc.customerID, tc.items)

			// Assert
			assert.Error(t, err)
			assert.Equal(t, tc.expectedErr, err)
			assert.Nil(t, order)

			// No mocks should be called in validation failure cases
			mockUow.AssertNotCalled(t, "Execute")
			mockOrderRepo.AssertNotCalled(t, "Create")
			mockOutboxRepo.AssertNotCalled(t, "CreateMessage")
			mockEventPublisher.AssertNotCalled(t, "Publish")
		})
	}
}

func TestCreateOrder_RepositoryFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Create mocks
	mockUow := new(MockUnitOfWork)
	mockOrderRepo := new(MockOrderRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	mockEventPublisher := new(MockEventPublisher)

	// Define expected error
	repoErr := errors.New("database error")

	// Setup mock expectations - simulate repository failure
	mockUow.On("Execute", ctx, mock.AnythingOfType("func(*sql.Tx) error")).Return(repoErr)
	mockOrderRepo.On("WithTx", mock.Anything).Return(mockOrderRepo)
	mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(repoErr)

	// Create the use case
	orderUseCase := usecase.NewOrderUseCase(mockUow, mockOrderRepo, mockOutboxRepo, mockEventPublisher)

	// Test data
	customerID := "customer123"
	items := []domain.OrderItem{
		{
			ProductID:   "product1",
			Quantity:    1,
			Price:       100.00,
			Description: "Product 1",
		},
	}

	// Act
	order, err := orderUseCase.CreateOrder(ctx, customerID, items)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)

	// Event publisher should not be called if transaction fails
	mockEventPublisher.AssertNotCalled(t, "Publish")
}

func TestCreateOrder_OutboxFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Create mocks
	mockUow := new(MockUnitOfWork)
	mockOrderRepo := new(MockOrderRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	mockEventPublisher := new(MockEventPublisher)

	// Define expected error
	outboxErr := errors.New("outbox error")

	// Setup mock expectations
	mockOrderRepo.On("WithTx", mock.Anything).Return(mockOrderRepo)
	mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	mockOutboxRepo.On("WithTx", mock.Anything).Return(mockOutboxRepo)
	mockOutboxRepo.On("CreateMessage", ctx, mock.AnythingOfType("uuid.UUID"), "order.created", mock.AnythingOfType("[]uint8")).Return(outboxErr)

	// UoW should return error because outbox creation failed
	mockUow.On("Execute", ctx, mock.AnythingOfType("func(*sql.Tx) error")).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*sql.Tx) error)
		_ = fn(nil) // This will execute the function and trigger the outbox error
	}).Return(outboxErr)

	// Create the use case
	orderUseCase := usecase.NewOrderUseCase(mockUow, mockOrderRepo, mockOutboxRepo, mockEventPublisher)

	// Test data
	customerID := "customer123"
	items := []domain.OrderItem{
		{
			ProductID:   "product1",
			Quantity:    1,
			Price:       100.00,
			Description: "Product 1",
		},
	}

	// Act
	order, err := orderUseCase.CreateOrder(ctx, customerID, items)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)

	// Event publisher should not be called if transaction fails
	mockEventPublisher.AssertNotCalled(t, "Publish")
}

func TestCreateOrder_EventPublishingFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Create mocks
	mockUow := new(MockUnitOfWork)
	mockOrderRepo := new(MockOrderRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	mockEventPublisher := new(MockEventPublisher)

	// Define expected error
	publishErr := errors.New("publish error")

	// Setup mock expectations - transaction succeeds but publishing fails
	mockUow.On("Execute", ctx, mock.AnythingOfType("func(*sql.Tx) error")).Return(nil)
	mockOrderRepo.On("WithTx", mock.Anything).Return(mockOrderRepo)
	mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	mockOutboxRepo.On("WithTx", mock.Anything).Return(mockOutboxRepo)
	mockOutboxRepo.On("CreateMessage", ctx, mock.AnythingOfType("uuid.UUID"), "order.created", mock.AnythingOfType("[]uint8")).Return(nil)
	mockEventPublisher.On("Publish", ctx, "order.created", mock.AnythingOfType("event.OrderCreatedEvent")).Return(publishErr)

	// Create the use case
	orderUseCase := usecase.NewOrderUseCase(mockUow, mockOrderRepo, mockOutboxRepo, mockEventPublisher)

	// Test data
	customerID := "customer123"
	items := []domain.OrderItem{
		{
			ProductID:   "product1",
			Quantity:    1,
			Price:       100.00,
			Description: "Product 1",
		},
	}

	// Act
	order, err := orderUseCase.CreateOrder(ctx, customerID, items)

	// Assert
	// Operation should still succeed even if publishing fails because we have the outbox as a backup
	assert.NoError(t, err)
	assert.NotNil(t, order)

	// Verify the mocks were called as expected
	mockUow.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestCreateOrder_EventMarshalFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Create mocks but use a real json package to simulate marshal error
	mockUow := new(MockUnitOfWork)
	mockOrderRepo := new(MockOrderRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	mockEventPublisher := new(MockEventPublisher)

	// Create a custom domain.OrderItem with a channel that can't be marshaled to JSON
	type badItem struct {
		domain.OrderItem
		BadField chan int // This will cause json.Marshal to fail
	}

	// Create the use case - we need to modify this test to work with the current implementation
	// since we can't directly inject a json.Marshal function to simulate failure
	orderUseCase := usecase.NewOrderUseCase(mockUow, mockOrderRepo, mockOutboxRepo, mockEventPublisher)

	// Test data with an item that can't be marshaled
	customerID := "customer123"
	items := []domain.OrderItem{
		{
			ProductID:   "product1",
			Quantity:    1,
			Price:       100.00,
			Description: "Product 1",
		},
	}

	// This test is a bit tricky because we can't easily mock json.Marshal failure
	// We'll skip the actual test but document how it might be approached

	t.Skip("This test requires the ability to mock or inject a failing json.Marshal function")

	/*
		Note: To properly test JSON marshal failure, we would need to:
		1. Modify the OrderUseCase to allow injecting a marshal function for testing
		2. Or use a mock that replaces the actual json package
		3. Or create a custom domain.OrderItem that has a field like a channel that can't be marshaled

		Since these approaches require changing the code under test or using advanced mocking techniques,
		we're documenting the test case but not implementing it directly.
	*/
}
