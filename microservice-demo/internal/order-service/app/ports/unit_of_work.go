package ports

import (
	"context"
	"database/sql"
)

// UnitOfWork defines the interface for managing transactions
type UnitOfWork interface {
	// Execute the function within the transaction
	Execute(ctx context.Context, fn func(*sql.Tx) error) error
	RegisterRepository(tx *sql.Tx, repository interface{}) error
	GetCurrentTransaction() *sql.Tx
}
