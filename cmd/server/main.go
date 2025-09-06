package main

import (
	"log"
	"net/http"

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
	defer repo.Close()

	cache := cache.New()

	orders, err := repo.GetAllOrders()
	if err != nil {
		log.Printf("Warning: failed to load orders from DB: %v", err)
	} else {
		cache.LoadFromDB(orders)
		log.Printf("Loaded %d orders to cache", len(orders))
	}

	handler := handlers.New(repo, cache)

	consumer := kafka.NewConsumer(repo, cache)
	go consumer.Start()

	r := mux.NewRouter()
	r.HandleFunc("/order/{order_uid}", handler.GetOrder).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	log.Println("Server starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
