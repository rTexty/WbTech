package cache

import (
	"sync"
	"wildberries-tech/internal/models"
)

type Cache struct {
	data map[string]models.Order
	mu   sync.RWMutex
}

func New() *Cache {
	return &Cache{
		data: make(map[string]models.Order),
	}
}

func (c *Cache) Set(orderUID string, order models.Order) {
	c.mu.Lock()
	c.data[orderUID] = order
	c.mu.Unlock()
}

func (c *Cache) Get(orderUID string) (models.Order, bool) {
	c.mu.RLock()
	order, exists := c.data[orderUID]
	c.mu.RUnlock()
	return order, exists
}

func (c *Cache) LoadFromDB(orders []models.Order) {
	c.mu.Lock()
	for _, order := range orders {
		c.data[order.OrderUID] = order
	}
	c.mu.Unlock()
}