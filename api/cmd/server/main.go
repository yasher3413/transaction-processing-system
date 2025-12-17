package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yash/transaction-system/api/internal/handler"
	"github.com/yash/transaction-system/api/internal/middleware"
	"github.com/yash/transaction-system/api/internal/service"
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
	shutdownTracer, err := tracing.InitTracer("api-service", cfg.JaegerEndpoint, logger)
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

	// Initialize services
	accountService := service.NewAccountService(database.DB, logger)
	transactionService := service.NewTransactionService(database.DB, logger)

	// Initialize handlers
	accountHandler := handler.NewAccountHandler(accountService, logger)
	transactionHandler := handler.NewTransactionHandler(transactionService, logger)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Metrics)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/v1", func(r chi.Router) {
		// Apply API key auth to all v1 routes
		r.Use(middleware.APIKeyAuth(cfg.APIKey))

		r.Route("/accounts", func(r chi.Router) {
			r.Post("/", accountHandler.CreateAccount)
			r.Get("/{id}", accountHandler.GetAccount)
		})

		r.Route("/transactions", func(r chi.Router) {
			r.Post("/", transactionHandler.CreateTransaction)
			r.Get("/", transactionHandler.ListTransactions)
			r.Get("/{id}", transactionHandler.GetTransaction)
		})
	})

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.APIPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("API server starting", zap.Int("port", cfg.APIPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func initLogger() (*zap.Logger, error) {
	env := os.Getenv("ENV")
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}


