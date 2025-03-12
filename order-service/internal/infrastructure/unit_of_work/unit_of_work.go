package unitofwork

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"order-service/internal/app/ports"
)

type SQLUnitOfWork struct {
	db                 *sql.DB
	repositoryCreators map[string]func(*sql.Tx) interface{}
}

func NewSQLUnitOfWork(db *sql.DB) *SQLUnitOfWork {
	return &SQLUnitOfWork{
		db:                 db,
		repositoryCreators: make(map[string]func(*sql.Tx) interface{}),
	}
}

// Execute runs a function within a transaction context
func (uow *SQLUnitOfWork) Execute(ctx context.Context, fn func(factory ports.RepositoryFactory) error) error {
	// Create a new transaction
	tx, err := uow.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Ensure transaction is eventually rolled back or committed
	defer func() {
		if tx != nil {
			err = tx.Rollback()
			if err != nil {
				log.Printf("rollback error: %v", err)
			}
		}
	}()

	factory := &SQL

	// Execute the function within the transaction
	if err = fn(f); err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Prevent rollback in defer
	tx = nil
	return nil

}

type SQLRepositoryFactory struct {
	tx                 *sql.Tx
	repositoryCreators map[string]func(*sql.Tx) interface{}
}
