// Package main implements the HTTP server and main application entry point.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"wildberries-tech/internal/cache"
	"wildberries-tech/internal/config"
	"wildberries-tech/internal/handlers"
	"wildberries-tech/internal/health"
	"wildberries-tech/internal/kafka"
	"wildberries-tech/internal/metrics"
	"wildberries-tech/internal/repository"
	"wildberries-tech/internal/tracing"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize tracing
	tracer, err := tracing.New(ctx, tracing.Config{
		ServiceName:    "order-service",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		JaegerEndpoint: cfg.Tracing.Endpoint,
		Enabled:        cfg.Tracing.Enabled,
	})
	if err != nil {
		log.Printf("Warning: failed to initialize tracing: %v", err)
	}
	defer func() {
		if tracer != nil {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := tracer.Shutdown(shutdownCtx); err != nil {
				log.Printf("Error shutting down tracer: %v", err)
			}
		}
	}()

	repo, err := repository.New(cfg.Database)
	if err != nil {
		log.Printf("Failed to initialize repository: %v", err)
		return
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Println("Error closing repository:", err)
		}
	}()

	c := cache.New(cfg.Cache.TTL, cfg.Cache.CleanupInterval)

	orders, err := repo.GetAllOrders()
	if err != nil {
		log.Printf("Warning: failed to load orders from DB: %v", err)
	} else {
		c.LoadFromDB(orders)
		log.Printf("Loaded %d orders to cache", len(orders))
	}

	handler := handlers.New(repo, c)

	m := metrics.NewPrometheus()

	// Initialize health checker
	sqlDB, err := repo.DB()
	if err != nil {
		log.Printf("Warning: failed to get sql.DB for health checks: %v", err)
	}
	healthChecker := health.NewChecker(sqlDB, cfg.Kafka.Brokers, m, 30*time.Second)
	go healthChecker.Start(ctx)

	consumer := kafka.NewConsumer(repo, c, m, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.DLQTopic)

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Consumer stopped with error: %v", err)
			cancel()
		}
	}()

	r := mux.NewRouter()
	r.Use(otelmux.Middleware("order-service"))
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		status := healthChecker.Status()
		w.Header().Set("Content-Type", "application/json")
		if healthChecker.IsHealthy() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		if err := json.NewEncoder(w).Encode(status); err != nil {
			log.Printf("Error encoding health status: %v", err)
		}
	}).Methods("GET")
	r.HandleFunc("/order/{order_uid}", handler.GetOrder).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	srv := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
