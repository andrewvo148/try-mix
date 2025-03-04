package handlers

import (
	"encoding/json"
	"net/http"
	"order-service/internal/app/ports"
	"order-service/internal/domain"
	"order-service/internal/interfaces/api/dto"

	"github.com/google/uuid"
)

// OrderHandler handles HTTP requests related to orders
type OrderHandler struct {
	orderUseCase ports.OrderUseCase
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderUseCase ports.OrderUseCase) *OrderHandler {
	return &OrderHandler{
		orderUseCase: orderUseCase,
	}
}

// Create handles the creation of a new order
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse("invalid request body"))
		return
	}

	// Validate request
	if req.CustomerID == "" {
		writeJSON(w, http.StatusBadGateway, errorResponse("invalid request body"))
		return
	}

	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadGateway, errorResponse("order must have at least one item"))
	}

	// Convert DTO to domain model
	items := make([]domain.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.OrderItem{
			ID:        uuid.New(),
			ProductID: item.ProductID,
			Quantity:  int32(item.Quantity),
			Price:     item.Price,
		})
	}

	// Create order
	order, err := h.orderUseCase.CreateOrder(r.Context(), req.CustomerID, items)
	if err != nil {
		handleError(w, err)
		return
	}

	// Convert domain model to response DTO
	resp := dto.OrderToResponse(order)
	writeJSON(w, http.StatusCreated, resp)

}

// Helper function

// writeJSON writes a JSON response to the given response writer
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// errorResponse creates a standardized error response
func errorResponse(message string) map[string]string {
	return map[string]string{"error": message}
}

// handleError handles domain-specific errors and returns appropriate HTTP responses
func handleError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrOrderNotFound:
		writeJSON(w, http.StatusNotFound, errorResponse("order not found"))
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse("internal server error"))
	}
}