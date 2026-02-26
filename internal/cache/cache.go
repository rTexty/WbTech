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

type Cache struct {
	store *gocache.Cache
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		store: gocache.New(defaultExpiration, cleanupInterval),
	}
}

func (c *Cache) Set(orderUID string, order models.Order) {
	c.store.Set(orderUID, order, gocache.DefaultExpiration)
}

func (c *Cache) Get(orderUID string) (models.Order, bool) {
	if val, found := c.store.Get(orderUID); found {
		if order, ok := val.(models.Order); ok {
			return order, true
		}
	}
	return models.Order{}, false
}

func (c *Cache) LoadFromDB(orders []models.Order) {
	for _, order := range orders {
		c.Set(order.OrderUID, order)
	}
}
