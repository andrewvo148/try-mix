package ports

import (
	"context"
	"inventory-service/internal/domain"
)

// ProductRepository defines the interface for product persistence
type ProductRepository interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
}
