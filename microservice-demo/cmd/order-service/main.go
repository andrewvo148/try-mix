package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"microservice-demo/internal/order-service/app/usecase"
	"microservice-demo/internal/order-service/infrastructure/config"
	"microservice-demo/internal/order-service/infrastructure/messaging/kafka"
	"microservice-demo/internal/order-service/infrastructure/repository"
	"microservice-demo/internal/order-service/infrastructure/worker"
	"microservice-demo/internal/order-service/interfaces/api/handlers"
	"microservice-demo/internal/order-service/interfaces/api/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	dbConn, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer dbConn.Close()

	// Check database connection
	if err := dbConn.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// run migrations
	// Run database migrations
	if err := runMigrations(dbConn, cfg.Database.MigrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize kafka producer
	producer, err := kafka.NewEventPublisher(cfg.Kafka)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	// Initialize dependencies
	workOfUnit := unitofwork.NewUnitOfWork(dbConn)
	orderRepo := repository.NewOrderRepository(dbConn)
	outboxRepo := repository.NewOutboxRepository(dbConn)

	orderUseCase := usecase.NewOrderUseCase(workOfUnit, orderRepo, outboxRepo, producer)
	orderHandler := handlers.NewOrderHandler(orderUseCase)

	// Setup router
	r := router.Setup(orderHandler)

	// Configure server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Create the outbox worker
	worker := worker.NewOutboxProcessor(
		outboxRepo,
		producer,

		100,
		5*time.Second,
		3,
	)

	// Start the worker with context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Start(ctx)

	// Wait for interrup signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

// runMigrations runs the database migrations from the specified path
func runMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}
