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

	"wildberries-tech/internal/cache"
	"wildberries-tech/internal/config"
	"wildberries-tech/internal/handlers"
	"wildberries-tech/internal/kafka"
	"wildberries-tech/internal/repository"
)

func main() {
	cfg := config.LoadConfig()

	repo, err := repository.New(cfg.Database.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		panic(err)
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

	consumer := kafka.NewConsumer(repo, c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.Start(ctx)

	r := mux.NewRouter()
	r.HandleFunc("/order/{order_uid}", handler.GetOrder).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	log.Println("Static file server configured: ./web/ directory")

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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
