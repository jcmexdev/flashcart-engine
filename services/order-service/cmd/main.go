package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	httpAdapter "github.com/jcmexdev/flashcart-engine/services/order/internal/adapters/http"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/adapters/messaging/rabbitmq"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/adapters/persistence/postgres"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/application"
)


func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	port := getEnv("PORT", "8080")

	// 1. PostgreSQL Database Configuration
	dbCfg := postgres.Config{
		Host:     getEnv("DB_HOST", "postgres"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "secretpassword"),
		DBName:   getEnv("DB_NAME", "orders"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Connecting to PostgreSQL at %s:%s (db: %s)...", dbCfg.Host, dbCfg.Port, dbCfg.DBName)
	pool, err := postgres.NewConnectionPool(ctx, dbCfg)
	if err != nil {
		log.Fatalf("Critical: unable to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("PostgreSQL connection pool initialized successfully.")

	// 2. RabbitMQ Configuration & Publisher
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	rabbitExchange := getEnv("RABBITMQ_EXCHANGE", "order.events")

	log.Printf("Connecting to RabbitMQ at %s...", rabbitURL)
	publisher, closeRabbit, err := rabbitmq.NewPublisher(rabbitmq.Config{
		URL:          rabbitURL,
		ExchangeName: rabbitExchange,
		ExchangeType: "topic",
	})
	if err != nil {
		log.Fatalf("Critical: unable to connect to RabbitMQ: %v", err)
	}
	defer func() {
		if err := closeRabbit(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}()
	log.Println("RabbitMQ publisher initialized successfully.")

	// 3. Composition Root (Hexagonal Wiring with Unit of Work)
	uowManager := postgres.NewUnitOfWorkManager(pool)
	createOrderUseCase := application.NewCreateOrderUseCase(uowManager)
	orderHandler := httpAdapter.NewOrderHandler(createOrderUseCase)

	// 4. Background Outbox Relayer Worker
	intervalMs, _ := strconv.Atoi(getEnv("RELAYER_INTERVAL_MS", "2000"))
	batchSize, _ := strconv.Atoi(getEnv("RELAYER_BATCH_SIZE", "50"))
	if intervalMs <= 0 {
		intervalMs = 2000
	}
	if batchSize <= 0 {
		batchSize = 50
	}

	relayer := application.NewOutboxRelayer(uowManager, publisher, application.OutboxRelayerConfig{
		PollInterval: time.Duration(intervalMs) * time.Millisecond,
		BatchSize:    batchSize,
	})

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	workerStopped := make(chan struct{})

	go func() {
		defer close(workerStopped)
		relayer.Start(workerCtx)
	}()

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      orderHandler.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Graceful Shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Order Service listening on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-stopChan
	log.Println("Shutdown signal received, shutting down gracefully...")

	// Stop background worker
	cancelWorker()
	<-workerStopped

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}

	log.Println("Order Service stopped.")
}

