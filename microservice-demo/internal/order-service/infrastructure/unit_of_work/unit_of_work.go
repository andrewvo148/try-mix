package unitofwork

import (
	"context"
	"database/sql"
	"fmt"
	"log"

)

type UnitOfWork struct {
	tx           *sql.Tx
	db           *sql.DB
	repositories map[string]interface{}
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{
		db: db,
	}
}

// Execute runs a function within a transaction context
func (uow *UnitOfWork) Execute(ctx context.Context, fn func(*sql.Tx) error) error {
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

	// Execute the function within the transaction
	if err = fn(tx); err != nil {
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

// Register a repository within the unit of work
func (uow *UnitOfWork) RegisterRepository(name string, repository interface{}) {
    uow.repositories[name] = repository
}

func (uow *UnitOfWork) GetRepository(name string) interface{} {
	if uow.tx != nil {
		return uow.repositories[name]
	}
}
