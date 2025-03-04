package usecase

import (
	"context"
	"inventory-service/internal/app/ports"
	"inventory-service/internal/domain"
)

// OrderUsecase implements the order business logic
type inventoryUseCase struct {
	productRepo ports.ProductRepository
}

// NewOrderUseCase creates a new order use case
func NewInventoryUseCase(productRepo ports.ProductRepository) ports.InventoryUseCase {
	return &inventoryUseCase{
		productRepo: productRepo,
	}
}

func (uc *inventoryUseCase) CreateProduct(ctx context.Context, product domain.Product) (*domain.Product, error) {
	// Set default values for new product
	// Check if product with same SKU already exists
	_, err := s.productRepo.GetProductBySKU(ctx, product.SKU)
	if err == nil {
		return nil, 
	}
}
