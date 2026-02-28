// Package health provides health checking for application dependencies.
package health

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"wildberries-tech/internal/metrics"
)

// Status represents the health status of a resource.
type Status struct {
	Database bool `json:"database"`
	Kafka    bool `json:"kafka"`
}

// Checker periodically checks the health of dependencies and updates metrics.
type Checker struct {
	db           *sql.DB
	kafkaBrokers []string
	metrics      metrics.Metrics
	interval     time.Duration

	mu     sync.RWMutex
	status Status
}

// NewChecker creates a new health checker.
func NewChecker(db *sql.DB, kafkaBrokers []string, m metrics.Metrics, interval time.Duration) *Checker {
	return &Checker{
		db:           db,
		kafkaBrokers: kafkaBrokers,
		metrics:      m,
		interval:     interval,
		status:       Status{Database: true, Kafka: true},
	}
}

// Start begins periodic health checking. Should be called in a goroutine.
func (c *Checker) Start(ctx context.Context) {
	c.check()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Health checker stopping...")
			return
		case <-ticker.C:
			c.check()
		}
	}
}

func (c *Checker) check() {
	dbUp := c.pingDB()
	kafkaUp := c.pingKafka()

	c.mu.Lock()
	c.status.Database = dbUp
	c.status.Kafka = kafkaUp
	c.mu.Unlock()

	if dbUp {
		c.metrics.SetResourceUp("database", 1)
	} else {
		c.metrics.SetResourceUp("database", 0)
	}

	if kafkaUp {
		c.metrics.SetResourceUp("kafka", 1)
	} else {
		c.metrics.SetResourceUp("kafka", 0)
	}
}

func (c *Checker) pingDB() bool {
	if c.db == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := c.db.PingContext(ctx); err != nil {
		log.Printf("Database ping failed: %v", err)
		return false
	}
	return true
}

func (c *Checker) pingKafka() bool {
	if len(c.kafkaBrokers) == 0 {
		return false
	}

	config := sarama.NewConfig()
	config.Net.DialTimeout = 2 * time.Second

	client, err := sarama.NewClient(c.kafkaBrokers, config)
	if err != nil {
		log.Printf("Kafka ping failed: %v", err)
		return false
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing Kafka client: %v", err)
		}
	}()

	// Requesting brokers proves connectivity
	_ = client.Brokers()
	return true
}

// Status returns the current health status.
func (c *Checker) Status() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// IsHealthy returns true if all dependencies are up.
func (c *Checker) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status.Database && c.status.Kafka
}
