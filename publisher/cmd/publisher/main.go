package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yash/transaction-system/publisher/internal/publisher"
	"github.com/yash/transaction-system/shared/config"
	"github.com/yash/transaction-system/shared/db"
	"github.com/yash/transaction-system/shared/tracing"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize tracing
	shutdownTracer, err := tracing.InitTracer("publisher-service", cfg.JaegerEndpoint, logger)
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

	// Create publisher
	outboxPublisher := publisher.NewOutboxPublisher(
		database.DB,
		cfg.KafkaBrokers,
		cfg.KafkaTransactionsTopic,
		cfg.PublisherBatchSize,
		cfg.PublisherInterval,
		logger,
	)
	defer outboxPublisher.Close()

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		if err := http.ListenAndServe(":8082", nil); err != nil {
			logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	// Start publisher
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := outboxPublisher.Start(ctx); err != nil {
		logger.Fatal("Publisher failed", zap.Error(err))
	}

	logger.Info("Publisher stopped")
}

func initLogger() (*zap.Logger, error) {
	env := os.Getenv("ENV")
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}


