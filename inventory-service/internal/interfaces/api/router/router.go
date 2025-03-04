package router

import (
	"order-service/internal/interfaces/api/handlers"
	"time"

	customMiddleware "order-service/internal/interfaces/api/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Setup configures and returns the API router
func Setup(orderHandler *handlers.OrderHandler) *chi.Mux {
	r := chi.NewRouter()

	// Apply global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(customMiddleware.ContentTypeJSON)
	r.Use(customMiddleware.Cors)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/orders", func(r chi.Router) {
			r.Post("/", orderHandler.Create) // Create a new order
		})
	})

	return r
}
