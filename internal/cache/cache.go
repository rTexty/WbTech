// Package cache implements in-memory caching for orders.
package cache

import (
	"time"
	"wildberries-tech/internal/models"

	gocache "github.com/patrickmn/go-cache"
)

// OrderCache defines the interface for caching orders
type OrderCache interface {
	Set(orderUID string, order models.Order)
	Get(orderUID string) (models.Order, bool)
	LoadFromDB(orders []models.Order)
}

// Cache provides methods for storing and retrieving orders from memory.
type Cache struct {
	store *gocache.Cache
}

// New initializes and returns a new Cache instance.
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		store: gocache.New(defaultExpiration, cleanupInterval),
	}
}

// Set adds an order to the cache.
func (c *Cache) Set(orderUID string, order models.Order) {
	c.store.Set(orderUID, order, gocache.DefaultExpiration)
}

// Get retrieves an order from the cache.
func (c *Cache) Get(orderUID string) (models.Order, bool) {
	if val, found := c.store.Get(orderUID); found {
		if order, ok := val.(models.Order); ok {
			return order, true
		}
	}
	return models.Order{}, false
}

// LoadFromDB populates the cache with a list of orders.
func (c *Cache) LoadFromDB(orders []models.Order) {
	for _, order := range orders {
		c.Set(order.OrderUID, order)
	}
}
