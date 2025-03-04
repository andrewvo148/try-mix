package usecase

import (
	"context"
	"order-service/internal/app/ports"
	"order-service/internal/domain"
	"order-service/internal/event"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ ports.OrderRepository = (*MockOrderRepository)(nil)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	return nil, nil
}

// Update implements ports.OrderRepository.
func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	return nil
}

func (m *MockOrderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	return nil, nil
}

func (m *MockOrderRepository) Delete(ctx context.Context, id string) error {
	return nil
}

var _ ports.Publisher = (*MockEventPubliser)(nil)

type MockEventPubliser struct {
	mock.Mock
}

// Publish implements ports.Publisher.
func (m *MockEventPubliser) Publish(ctx context.Context, topic string, event interface{}) error {
	args := m.Called(ctx, topic, event)
	return args.Error(0)
}

func TestCreateOrder_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockOrderRepository)
	mockPubliser := new(MockEventPubliser)

	uc := NewOrderUseCase(mockRepo, mockPubliser)

	ctx := context.Background()
	customerID := "customer-123"
	items := []domain.OrderItem{
		{ProductID: "product-1", Quantity: 2, Price: 10.0},
		{ProductID: "product-2", Quantity: 1, Price: 15.0},
	}

	// Expect repository and publisher calls
	mockRepo.On("Create", ctx, mock.MatchedBy(func(order *domain.Order) bool {
		return order.CustomerID == customerID &&
			len(order.Items) == 2 &&
			order.Status == domain.OrderStatusPending &&
			order.SagaID != uuid.Nil
	})).Return(nil)

	mockPubliser.On("Publish", ctx, "order.created", mock.MatchedBy(func(event event.OrderCreatedEvent) bool {
		return event.CustomerID == customerID &&
			len(event.Items) == 2 &&
			event.SageID != uuid.Nil
	})).Return(nil)
	// Act
	order, err := uc.CreateOrder(ctx, customerID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, customerID, order.CustomerID)
	assert.Equal(t, domain.OrderStatusPending, order.Status)
	assert.NotEmpty(t, order.SagaID)
	assert.Len(t, order.Items, 2)

	mockRepo.AssertExpectations(t)
	mockPubliser.AssertExpectations(t)

}
