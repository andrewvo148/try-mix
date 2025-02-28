package repository

import (
	"errors"
	"sync"

	"order-service/internal/model"
)

// In a real application, this would be using a database
type OrderRepository struct {
	orders map[string]*model.Order
	mutex  sync.RWMutex
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[string]*model.Order),
	}
}

func (r *OrderRepository) Save(order *model.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.orders[order.ID.String()] = order
	return nil
}

func (r *OrderRepository) FindByID(id string) (*model.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if order, exists := r.orders[id]; exists {
		return order, nil
	}
	return nil, errors.New("order not found")
}

func (r *OrderRepository) Update(order *model.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.orders[order.ID.String()]; !exists {
		return errors.New("order not found")
	}

	r.orders[order.ID.String()] = order
	return nil
} 