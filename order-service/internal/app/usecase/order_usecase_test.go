package usecase_test

import (
	"context"
	"database/sql"
	"order-service/internal/app/usecase"
	"order-service/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories and UnitOfWork
type mockOrderRepo struct {
	mock.Mock
}

func (m *mockOrderRepo) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

type mockOutboxRepo struct {
	mock.Mock
}

func (m *mockOutboxRepo) CreateMessage(ctx context.Context, aggregateID uuid.UUID, eventType string, payload interface{}) error {
	args := m.Called(ctx, aggregateID, eventType, payload)
	return args.Error(0)
}

type mockUnitOfWork struct {
	mock.Mock
	mockOrderRepo  *mockOrderRepo
	mockOutboxRepo *mockOutboxRepo
	shouldFail     bool
	failAt         string
}

func (m *mockUnitOfWork) Execute(ctx context.Context, fn func(tx *sql.Tx) error) error {
	args := m.Called(ctx)
	if m.shouldFail {
		return args.Error(0)
	}

	err := fn(nil) // Pass nil as tx since we're mocking
	return err
}

type mockEventPublisher struct {
	mock.Mock
}

func (m *mockEventPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
	args := m.Called(ctx, topic, event)
	return args.Error(0)
}

func TestCreateOrder(t *testing.T) {
	// Test cases
	testCases := []struct {
		name            string
		customerID      string
		items           []domain.OrderItem
		setupMocks      func(*mockUnitOfWork, *mockOrderRepo, *mockOutboxRepo, *mockEventPublisher)
		expectedError   bool
		expectedErrType error
	}{
		{
			name:       "Success - Order creation successful",
			customerID: "customer-123",
			items: []domain.OrderItem{
				{ID: uuid.New(), ProductID: "product-1", Quantity: 2, Price: 10.0},
				{ID: uuid.New(), ProductID: "product-2", Quantity: 1, Price: 20.0},
			},
			setupMocks: func(muow *mockUnitOfWork, mor *mockOrderRepo, moutbox *mockOutboxRepo, mep *mockEventPublisher) {
				muow.On("Execute", mock.Anything).Return(nil)
				mor.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)
				moutbox.On("CreateMessage", mock.Anything, uuid.New(), "order.created", mock.AnythingOfType("[]uint8")).Return(nil)
				mep.On("Publish", mock.Anything, "order.created", mock.AnythingOfType("*events.OrderCreatedEvent")).Return(nil)
			},
			expectedError: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockOrderRepo := new(mockOrderRepo)
			mockOutboxRepo := new(mockOutboxRepo)
			mockUoW := &mockUnitOfWork{
				mockOrderRepo:  mockOrderRepo,
				mockOutboxRepo: mockOutboxRepo,
			}
			mockPubliser := new(mockEventPublisher)

			// Apply test case setup

			tc.setupMocks(mockUoW, mockOrderRepo, mockOutboxRepo, mockPubliser)

			// Create the use case
			orderUseCase := usecase.NewOrderUseCase(mockUoW, mockPubliser)

			// Setup context
			ctx := context.Background()

			// Call method
			order, err := orderUseCase.CreateOrder(ctx, tc.customerID, tc.items)

			// Check expectations
			if tc.expectedError {
				assert.Error(t, err)
				if tc.expectedErrType != nil {
					assert.ErrorIs(t, err, tc.expectedErrType)
				}
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
			}
		})
	}
}
