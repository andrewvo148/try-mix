package domain

import (
	"time"

	"github.com/google/uuid"
)

// StockStatus represents the current stock level status of a product
type StockStatus string

const (
	StockStatusInStock      StockStatus = "IN_STOCK"
	StockStatusLowStock     StockStatus = "LOW_STOCK"
	StockStatusOutOfStock   StockStatus = "OUT_OF_STOCK"
	StockStatusDiscontinued StockStatus = "DISCONTINUED"
)

// Product represents a product in the inventory system

type Product struct {
	ID          uuid.UUID
	SKU         string
	Name        string
	Description string
	Category    string
	Price       float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// InventoryItem represents the stock level of a product
type InvetoryItem struct {
	ID                uuid.UUID
	ProductID         uuid.UUID
	Quantity          int32
	ReservedQuantity  int32
	AvailableQuantity int32
	ReorderPoint      int32
	ReorderQuantity   int32
	StockStatus       StockStatus
	LocationCode      string
	LastStockedAt     time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// InventoryTransaction represents a change in inventory
type InventoryTransaction struct {
	ID uuid.UUID
	ProductID uuid.UUID
	Quantity int32
	Type string // // "RESTOCK", "SALE", "RETURN", "ADJUSTMENT", "RESERVATION", "RELEASE"
	ReferenceID string // OrderID or other reference
	Note string
	PerformedBy string // UserID who performed the transaction
	TransactedAt time.Time
	CreatedAt time.Time
}