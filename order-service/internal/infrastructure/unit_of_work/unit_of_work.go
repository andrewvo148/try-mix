package unitofwork

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"order-service/internal/app/ports"
	"order-service/internal/infrastructure/repository"
)

type UnitOfWork struct {
	db         *sql.DB
	tx         *sql.Tx
	txFactory  TransactionFactory
	orderRepo  ports.OrderRepository
	outboxRepo ports.OutboxRepository
}

type TransactionFactory interface {
	CreateTransaction(ctx context.Context, db *sql.DB) (*sql.Tx, error)
}

func (uow *UnitOfWork) Begin(ctx context.Context) error {
	if uow.tx != nil {
		return errors.New("transaction already in progress")
	}

	tx, err := uow.txFactory.CreateTransaction(ctx, uow.db)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	uow.tx = tx
	uow.orderRepo = repository.NewOrderRepository(tx)
	
	return nil
}

func (uow *UnitOfWork) Commit(ctx context.Context) error {
	if uow.tx == nil {
		return errors.New("no active transaction")
	}
	return uow.tx.Commit()

}

func (uow *UnitOfWork) Rollback(ctx context.Context) error {
	if uow.tx == nil {
		return errors.New("no active transaction")
	}

	return uow.tx.Rollback()
}

func (uow *UnitOfWork) orders() ports.OrderRepository {
	return repository.NewOrderRepository(uow.tx)
}
