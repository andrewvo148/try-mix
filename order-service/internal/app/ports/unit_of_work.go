package ports

import (
	"context"
	"database/sql"
)

// UnitOfWork defines the interface for managing transactions
type UnitOfWork interface {
	// Execute the function within the transaction
	Execute(ctx context.Context, fn func(*sql.Tx) error) error
}

// RepositoryFactory creates repositories for use within a transaction
type RepositoryFactory interface {
	// GetRepository returns a repository for the requested type
	// The type parameter allows for compile-time type checking
	GetRepository(ctx context.Context, requestID string) (interface{}, error)
}

// EventPublisher publishes events to external systems
