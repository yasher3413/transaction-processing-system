package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yash/transaction-system/shared/config"
	"github.com/yash/transaction-system/shared/db"
	"github.com/yash/transaction-system/shared/tracing"
	"github.com/yash/transaction-system/worker/internal/consumer"
	"github.com/yash/transaction-system/worker/internal/processor"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize tracing
	shutdownTracer, err := tracing.InitTracer("worker-service", cfg.JaegerEndpoint, logger)
	if err != nil {
		logger.Warn("Failed to initialize tracing", zap.Error(err))
	}
	defer shutdownTracer()

	// Connect to database
	database, err := db.NewDB(cfg.GetPostgresDSN(), logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// Create processor
	transactionProcessor := processor.NewTransactionProcessor(database.DB, logger)

	// Create consumer
	kafkaConsumer := consumer.NewKafkaConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTransactionsTopic,
		cfg.WorkerConsumerGroup,
		cfg.KafkaDLQTopic,
		5,             // max retries
		2*time.Second, // retry backoff
		transactionProcessor,
		logger,
	)
	defer kafkaConsumer.Close()

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})
		if err := http.ListenAndServe(":8081", nil); err != nil {
			logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	// Start consumer
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := kafkaConsumer.Start(ctx); err != nil {
		logger.Fatal("Consumer failed", zap.Error(err))
	}

	logger.Info("Worker stopped")
}

func initLogger() (*zap.Logger, error) {
	env := os.Getenv("ENV")
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
