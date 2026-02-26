package cache

import (
	"testing"
	"time"
	"wildberries-tech/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	c := New(5*time.Minute, 10*time.Minute)

	order := models.Order{OrderUID: "test-uid"}
	c.Set("test-uid", order)

	val, found := c.Get("test-uid")
	assert.True(t, found)
	assert.Equal(t, order, val)

	val, found = c.Get("non-existent")
	assert.False(t, found)
	assert.Equal(t, models.Order{}, val)
}

func TestCacheTTL(t *testing.T) {
	c := New(100*time.Millisecond, 200*time.Millisecond)

	order := models.Order{OrderUID: "test-uid"}
	c.Set("test-uid", order)

	time.Sleep(200 * time.Millisecond)

	_, found := c.Get("test-uid")
	assert.False(t, found, "Item should have expired")
}
