package handlers

import (
	"encoding/json"
	"net/http"
	"wildberries-tech/internal/cache"
	"wildberries-tech/internal/repository"

	"github.com/gorilla/mux"
)

type Handler struct {
	repo  repository.OrderRepository
	cache cache.OrderCache
}

func New(repo repository.OrderRepository, cache cache.OrderCache) *Handler {
	return &Handler{
		repo:  repo,
		cache: cache,
	}
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	order, exists := h.cache.Get(orderUID)
	if exists {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
		return
	}

	orderPtr, err := h.repo.GetOrder(orderUID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	h.cache.Set(orderUID, *orderPtr)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(*orderPtr); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
