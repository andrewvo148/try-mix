package repository

import (
	"database/sql"
	"inventory-service/internal/infrastructure/sqlc"
)

// inventoryRepository implements the inventoryRepository interface using SQLC and PostgresSQL
type inventoryRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// // NewinventoryRepository creates a new inventory repository
// func NewinventoryRepository(db *sql.DB) ports.inventoryRepository {
// 	return &inventoryRepository{
// 		db:      db,
// 		queries: sqlc.New(db),
// 	}
// }
