package handlers

import (
	"encoding/json"
	"net/http"
	"wildberries-tech/internal/cache"
	"wildberries-tech/internal/repository"

	"github.com/gorilla/mux"
)

type Handler struct {
	repo  *repository.Repository
	cache *cache.Cache
}

func New(repo *repository.Repository, cache *cache.Cache) *Handler {
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
		json.NewEncoder(w).Encode(order)
		return
	}

	orderPtr, err := h.repo.GetOrder(orderUID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	h.cache.Set(orderUID, *orderPtr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(*orderPtr)
}