 package usecase

import (
	"context"
	"errors"
	"order-service/internal/app/ports"
	"order-service/internal/domain"
	"testing"

)

// MockUnitOfWork implements ports.UnitOfWork for testing
type MockUnitOfWork struct {
	mock.Mock
}

func (m *MockUnitOfWork) Begin(ctx context.Context) (context.Context, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(context.Context), args.Error(1)
}

func (m *MockUnitOfWork) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockUnitOfWork) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockUnitOfWork) Orders() ports.OrderRepository {
	args := m.Called()
	return args.Get(0).(ports.OrderRepository)
}

func (m *MockUnitOfWork) OutboxMessages() ports.OutboxRepository {
	args := m.Called()
	return args.Get(0).(ports.OutboxRepository)
}

type MockOutboxRepository struct {
	mock.Mock
}

func (m *MockOutboxRepository) CreateMessage(ctx context.Context, aggregateID, messageType string, payload interface{}) error {
	args := m.Called(ctx, aggregateID, messageType, payload)
	return args.Error(0)
}

func (m *MockOutboxRepository) GetPendingMessages(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	return nil, nil
}

func (m *MockOutboxRepository) MarkMessageAsProcessed(ctx context.Context, messageID string) error {
	return nil
}

func (m *MockOutboxRepository) MarkMessageAsFailed(ctx context.Context, messageID string, reason string) error {
	return nil
}

var _ ports.OrderRepository = (*MockOrderRepository)(nil)

// func (m *MockUnitOfWork) Begin(ctx context.Context) (context.Context, error) {
// 	args := m.Called(ctx)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(context.Context), args.Error(1)
// }

// func (m *MockUnitOfWork) Commit(ctx context.Context) error {
// 	args := m.Called(ctx)
// 	return args.Error(0)
// }

// func (m *MockUnitOfWork) Rollback(ctx context.Context) error {
// 	args := m.Called(ctx)
// 	return args.Error(0)
// }

// func (m *MockUnitOfWork) Orders(ctx context.Context) ports.UnitOfWork {
// 	args := m.Called()
// 	return args.Get(0).(ports.UnitOfWork)
// }

// func (m *MockUnitOfWork) OutboxMessages() ports.OutboxRepository {
// 	args := m.Called()
// 	return args.Get(0).(ports.OutboxRepository)
// }

// // MockOrderRepository implements for ports.OrderRepository for testing purposes
// type MockOrderRepository struct {
// 	mock.Mock
// }

var _ ports.EventPublisher = (*MockEventPubliser)(nil)

// func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
// 	args := m.Called(ctx, id)
// 	return args.Get(0).(*domain.Order), args.Error(1)
// }

// func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
// 	args := m.Called(ctx, order)
// 	return args.Error(0)
// }

func TestCreateOrder_Success(t *testing.T) {
	// Arrange
	mockUow := new(MockUnitOfWork)
	mockOrderRepo := new(MockOrderRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	mockEventPubliser := new(MockEventPubliser)

	// Setup mock transaction context
	ctx := context.Background()
	txCtx := context.WithValue(ctx, "tx", "transaction-id")

	customerID := "customer-123"
	items := []domain.OrderItem{
		{ProductID: "product-1", Quantity: 2, Price: 10.0},
		{ProductID: "product-2", Quantity: 2, Price: 15.0},
	}

	// Setup Uow expectations
	mockUow.On("Begin", ctx).Return(txCtx, nil)
	mockUow.On("Orders").Return(mockOrderRepo)
	mockUow.On("OutboxMessages").Return(mockOutboxRepo)
	mockUow.On("Commit", txCtx).Return(nil)

	// Setup OrderRepository expectations
	mockOrderRepo.On("Create", txCtx, mock.MatchedBy(func(order *domain.Order) bool {
		return order.CustomerID == customerID &&
			len(order.Items) == 2 &&
			order.Status == domain.OrderStatusPending &&
			order.SagaID != uuid.Nil
	})).Return(nil)

	mockOutboxRepo.On("CreateMessage", txCtx, mock.AnythingOfType("string"),
		"order.created", mock.AnythingOfType("[]uint8")).Return(nil)

	uc := NewOrderUseCase(mockUow, mockEventPubliser)

	// Act
	order, err := uc.CreateOrder(ctx, customerID, items)

// func (m *MockOutboxRepository) CreateMessage(ctx context.Context, aggregateID string, eventType string, payload []byte) error {
// 	args := m.Called(ctx, aggregateID, eventType, payload)
// 	return args.Error(0)
// }

	mockUow.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
}

func TestCreateOrder_InvalidInput(t *testing.T) {
	// Arrange
	mockUow := new(MockUnitOfWork)
	mockEventPubliser := new(MockEventPubliser)
	uc := NewOrderUseCase(mockUow, mockEventPubliser)
	ctx := context.Background()

	// Test cases
	testCases := []struct {
		name          string
		customerID    string
		items         []domain.OrderItem
		expectedError error
	}{
		{
			name:          "empty customer ID",
			customerID:    "",
			items:         []domain.OrderItem{{ProductID: "product-1", Quantity: 1, Price: 10.0}},
			expectedError: domain.ErrInvalidCustomerID,
		},
		{
			name:          "empty order items",
			customerID:    "customer-123",
			items:         []domain.OrderItem{},
			expectedError: domain.ErrEmptyOrderItems,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			order, err := uc.CreateOrder(ctx, tc.customerID, tc.items)

			// Assert
			assert.Error(t, err)
			assert.Equal(t, tc.expectedError, err)
			assert.Nil(t, order)
		})
	}
}

func TestCreateOrder_TransactionError(t *testing.T) {
	// Arrange
	mockUow := new(MockUnitOfWork)
	mockEventPubliser := new(MockEventPubliser)
	ctx := context.Background()

	customerID := "customer-123"
	items := []domain.OrderItem{
		{ProductID: "product-1", Quantity: 2, Price: 10.0},
	}

	// Setup Uow expectations - transaction begin fails
	mockUow.On("Begin", ctx).Return(nil, errors.New("transaction error"))

	uc := NewOrderUseCase(mockUow, mockEventPubliser)

	// Act
	order, err := uc.CreateOrder(ctx, customerID, items)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "failed to begin transaction")

	mockUow.AssertExpectations(t)
}

func TestCreateOrder_OrderRepositoryError(t *testing.T) {
	// Arrange
	mockUow := new(MockUnitOfWork)
	mockOrderRepo := new(MockOrderRepository)
	mockEventPubliser := new(MockEventPubliser)

	// Setup mock transaction context
	ctx := context.Background()
	txCtx := context.WithValue(ctx, "tx", "transaction-id")

	customerID := "customer-123"
	items := []domain.OrderItem{
		{ProductID: "product-1", Quantity: 2, Price: 10.0},
	}

	// Setup Uow expectations - transaction begin succeeds
	mockUow.On("Begin", ctx).Return(txCtx, nil)
	mockUow.On("Orders").Return(mockOrderRepo)
	mockUow.On("Rollback", txCtx).Return(nil)

	// Setup OrderRepository expectations - create fails
	mockOrderRepo.On("Create", txCtx, mock.AnythingOfType("*domain.Order")).Return(errors.New("database error"))

	uc := NewOrderUseCase(mockUow, mockEventPubliser)

	// Act
	order, err := uc.CreateOrder(ctx, customerID, items)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "failed to create order")

	mockUow.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
}

func TestCreateOrder_OutboxRepositoryError(t *testing.T) {
	// Arrange
	mockUow := new(MockUnitOfWork)
	mockOrderRepo := new(MockOrderRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	mockEventPubliser := new(MockEventPubliser)

	// Setup mock transaction context
	ctx := context.Background()
	txCtx := context.WithValue(ctx, "tx", "transaction-id")

	customerID := "customer-123"
	items := []domain.OrderItem{
		{ProductID: "product-1", Quantity: 2, Price: 10.0},
	}

	// Setup UoW expectations
	mockUow.On("Begin", ctx).Return(txCtx, nil)
	mockUow.On("Orders").Return(mockOrderRepo)
	mockUow.On("OutboxMessages").Return(mockOutboxRepo)
	mockUow.On("Rollback", txCtx).Return(nil)

	// Setup repository expectations
	mockOrderRepo.On("Create", txCtx, mock.AnythingOfType("*domain.Order")).Return(nil)
	mockOutboxRepo.On("CreateMessage", txCtx, mock.AnythingOfType("string"), "order.created",
		mock.AnythingOfType("[]uint8")).Return(errors.New("outbox error"))

	uc := NewOrderUseCase(mockUow, mockEventPubliser)

	// Act
	order, err := uc.CreateOrder(ctx, customerID, items)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "failed to create outbox message")
	mockUow.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)

// 	// Setup mock transaction context
// 	ctx := context.Background()
// 	txCtx := context.WithValue(ctx, "tx", "transaction-id")

// 	customerID := "customer-123"
// 	items := []domain.OrderItem{
// 		{ProductID: "product-1", Quantity: 2, Price: 10.0},
// 		{ProductID: "product-2", Quantity: 1, Price: 15.0},
// 	}

// 	// Setup UoW expectations
// 	mockUow.On("Begin", ctx).Return(txCtx, nil)
// 	mockUow.On("Orders").Return(mockOrderRepo)
// 	mockUow.On("OutboxMessages").Return(mockOutboxRepo)
// 	mockUow.On("Commit", txCtx).Return(nil)

// 	// Setup repository expectations
// 	mockOrderRepo.On("Create", txCtx, mock.MatchedBy(func(order *domain.Order) bool {
// 		return order.CustomerID == customerID &&
// 			len(order.Items) == 2 &&
// 			order.Status == domain.OrderStatusPending &&
// 			order.SagaID != uuid.Nil
// 	})).Return(nil)

// 	mockOutboxRepo.On("CreateMessage", txCtx, mock.AnythingOfType("string"), "order.created", mock.AnythingOfType("[]uint8")).Return(nil)

// 	uc := NewOrderUseCase(mockUow, mockEventPublisher)

// 	// Act
// 	order, err := uc.CreateOrder(ctx, customerID, items)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.NotNil(t, order)
// 	assert.Equal(t, customerID, order.CustomerID)
// 	assert.Equal(t, domain.OrderStatusPending, order.Status)
// 	assert.NotEmpty(t, order.SagaID)
// 	assert.Len(t, order.Items, 2)

// 	mockUow.AssertExpectations(t)
// 	mockOrderRepo.AssertExpectations(t)
// 	mockOutboxRepo.AssertExpectations(t)
// }

// func TestCreateOrder_InvalidInput(t *testing.T) {
// 	// Arrange
// 	mockUow := new(MockUnitOfWork)
// 	mockEventPublisher := new(MockEventPubliser)
// 	uc := NewOrderUseCase(mockUow, mockEventPublisher)
// 	ctx := context.Background()

// 	// Test cases
// 	testCases := []struct {
// 		name        string
// 		customerID  string
// 		items       []domain.OrderItem
// 		expectedErr error
// 	}{
// 		{
// 			name:        "empty customer ID",
// 			customerID:  "",
// 			items:       []domain.OrderItem{{ProductID: "product-1", Quantity: 1, Price: 10.0}},
// 			expectedErr: domain.ErrInvalidCustomerID,
// 		},
// 		{
// 			name:        "empty order items",
// 			customerID:  "customer-123",
// 			items:       []domain.OrderItem{},
// 			expectedErr: domain.ErrEmptyOrderItems,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Act
// 			order, err := uc.CreateOrder(ctx, tc.customerID, tc.items)

// 			// Assert
// 			assert.Error(t, err)
// 			assert.Equal(t, tc.expectedErr, err)
// 			assert.Nil(t, order)
// 		})
// 	}

// 	// No methods should be called on the mocks for invalid input
// 	mockUow.AssertNotCalled(t, "Begin")
// 	mockUow.AssertNotCalled(t, "Commit")
// 	mockUow.AssertNotCalled(t, "Rollback")
// }

// func TestCreateOrder_TransactionError(t *testing.T) {
// 	// Arrange
// 	mockUow := new(MockUnitOfWork)
// 	mockEventPublisher := new(MockEventPubliser)
// 	ctx := context.Background()

// 	customerID := "customer-123"
// 	items := []domain.OrderItem{
// 		{ProductID: "product-1", Quantity: 2, Price: 10.0},
// 	}

// 	// Setup UoW expectations - transaction begin fails
// 	mockUow.On("Begin", ctx).Return(nil, errors.New("transaction error"))

// 	uc := NewOrderUseCase(mockUow, mockEventPublisher)

// 	// Act
// 	order, err := uc.CreateOrder(ctx, customerID, items)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Nil(t, order)
// 	assert.Contains(t, err.Error(), "failed to begin transaction")

// 	mockUow.AssertExpectations(t)
// }

// func TestCreateOrder_RepositoryError(t *testing.T) {
// 	// Arrange
// 	mockUow := new(MockUnitOfWork)
// 	mockOrderRepo := new(MockOrderRepository)
// 	mockEventPublisher := new(MockEventPubliser)

// 	// Setup mock transaction context
// 	ctx := context.Background()
// 	txCtx := context.WithValue(ctx, "tx", "transaction-id")

// 	customerID := "customer-123"
// 	items := []domain.OrderItem{
// 		{ProductID: "product-1", Quantity: 2, Price: 10.0},
// 	}

// 	// Setup UoW expectations
// 	mockUow.On("Begin", ctx).Return(txCtx, nil)
// 	mockUow.On("Orders").Return(mockOrderRepo)
// 	mockUow.On("Rollback", txCtx).Return(nil)

// 	// Setup repository expectations - Create fails
// 	mockOrderRepo.On("Create", txCtx, mock.AnythingOfType("*domain.Order")).Return(errors.New("database error"))

// 	uc := NewOrderUseCase(mockUow, mockEventPublisher)

// 	// Act
// 	order, err := uc.CreateOrder(ctx, customerID, items)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Nil(t, order)
// 	assert.Contains(t, err.Error(), "failed to create order")

// 	mockUow.AssertExpectations(t)
// 	mockOrderRepo.AssertExpectations(t)
// }
