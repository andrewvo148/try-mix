package domain

import "errors"

// Domain errors
var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidOrderID = errors.New("invalid order ID")
	ErrInvalidCustomerID = errors.New("invalid customer ID")
	ErrInvalidProductID = errors.New("invalid product ID")
	ErrInvalidQuantity = errors.New("invalid quantity")
	ErrInvalidPrice = errors.New("invalid price")
	ErrEmptyOrderItems = errors.New("order must have at least one item")
)
