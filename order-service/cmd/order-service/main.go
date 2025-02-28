package main

import (
	"log"

	"order-service/internal/event"
	"order-service/internal/handler"
	"order-service/internal/repository"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Kafka handler
	kafkaHandler, err := event.NewKafkaHandler([]string{"localhost:29092", "localhost:29093", "localhost:29094"})
	if err != nil {
		log.Fatalf("Failed to initialize Kafka handler: %v", err)
	}
	defer kafkaHandler.Close()

	// Initialize repository
	orderRepo := repository.NewOrderRepository()

	// Initialize service
	orderService := service.NewOrderService(orderRepo, kafkaHandler)

	// Initialize handler
	orderHandler := handler.NewOrderHandler(orderService)

	// Setup router
	router := gin.Default()

	// Register routes
	router.POST("/orders", orderHandler.CreateOrder)
	router.PUT("/orders/:id/status", orderHandler.UpdateOrderStatus)

	// Start server
	if err := router.Run(":8085"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
