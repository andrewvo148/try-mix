package repository

import (
	"context"
	"database/sql"
	"inventory-service/internal/app/ports"
	"inventory-service/internal/domain"
	"inventory-service/internal/infrastructure/sqlc"
)

// inventoryRepository implements the InventoryRepository interface using SQLC and PostgresSQL
type productRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewProductRepository(db *sql.DB) ports.ProductRepository {
	return &productRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (p *productRepository) CreateProduct(ctx context.Context, product *domain.Product) error {
	return nil
}
