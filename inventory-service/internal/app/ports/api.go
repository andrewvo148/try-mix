package ports

import (
	"context"
	"inventory-service/internal/domain"
)

// InvetoryUseCase
type InventoryUseCase interface {
	CreateProduct(ctx context.Context, product domain.Product) (domain.Product, error)
}
