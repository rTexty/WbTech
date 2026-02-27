// Package main implements the HTTP server and main application entry point.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"wildberries-tech/internal/cache"
	"wildberries-tech/internal/config"
	"wildberries-tech/internal/handlers"
	"wildberries-tech/internal/kafka"
	"wildberries-tech/internal/repository"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	repo, err := repository.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
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

	consumer := kafka.NewConsumer(repo, c, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.DLQTopic)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Consumer stopped with error: %v", err)
			cancel() // Stop the server if consumer fails critically
		}
	}()

	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
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
	cancel() // Cancel context for consumer

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
