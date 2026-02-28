// Package kafka provides functionality for consuming order messages from Kafka.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"

	"wildberries-tech/internal/cache"
	"wildberries-tech/internal/metrics"
	"wildberries-tech/internal/models"
	"wildberries-tech/internal/repository"
)

// Consumer consumes orders from Kafka and saves them to the repository.
type Consumer struct {
	repo        repository.OrderRepository
	cache       cache.OrderCache
	metrics     metrics.Metrics
	brokers     []string
	topic       string
	dlqTopic    string
	dlqProducer sarama.SyncProducer
}

// NewConsumer creates a new Consumer instance.
func NewConsumer(repo repository.OrderRepository, cache cache.OrderCache, m metrics.Metrics,
	brokers []string, topic, dlqTopic string) *Consumer {
	return &Consumer{
		repo:     repo,
		cache:    cache,
		metrics:  m,
		brokers:  brokers,
		topic:    topic,
		dlqTopic: dlqTopic,
	}
}

// Start begins consuming messages from Kafka.
func (c *Consumer) Start(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Initialize DLQ Producer
	dlqConfig := sarama.NewConfig()
	dlqConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(c.brokers, dlqConfig)
	if err != nil {
		return fmt.Errorf("error creating DLQ producer: %w", err)
	}
	c.dlqProducer = producer
	defer func() {
		if err := c.dlqProducer.Close(); err != nil {
			log.Println("Error closing DLQ producer:", err)
		}
	}()

	consumer, err := sarama.NewConsumer(c.brokers, config)
	if err != nil {
		return fmt.Errorf("error creating consumer: %w", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Println("Error closing consumer:", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(c.topic, 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("error creating partition consumer: %w", err)
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Println("Error closing partition consumer:", err)
		}
	}()

	log.Println("Kafka consumer started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping consumer...")
			return nil
		case msg := <-partitionConsumer.Messages():
			c.processMessage(msg.Value)
		case err := <-partitionConsumer.Errors():
			log.Printf("Consumer error: %v", err)
		}
	}
}

func (c *Consumer) processMessage(data []byte) {
	var order models.Order

	err := json.Unmarshal(data, &order)
	if err != nil {
		log.Println("Error unmarshaling message:", err)
		c.handleError(data, err)
		return
	}

	if err := order.Validate(); err != nil {
		log.Printf("Validation failed for order %s: %v", order.OrderUID, err)
		c.handleError(data, err)
		return
	}

	err = c.repo.SaveOrder(order)
	if err != nil {
		log.Println("Error saving to DB:", err)
		c.handleError(data, err)
		return
	}

	c.cache.Set(order.OrderUID, order)
	c.metrics.IncMessagesTotal("success")

	log.Printf("Order %s processed successfully", order.OrderUID)
}

func (c *Consumer) handleError(data []byte, err error) {
	c.metrics.IncMessagesTotal("error")

	// Send to DLQ
	msg := &sarama.ProducerMessage{
		Topic: c.dlqTopic,
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{Key: []byte("error"), Value: []byte(err.Error())},
		},
	}

	partition, offset, err := c.dlqProducer.SendMessage(msg)
	if err != nil {
		log.Printf("FAILED to send message to DLQ: %v", err)
	} else {
		log.Printf("Message sent to DLQ topic %s (partition: %d, offset: %d)", c.dlqTopic, partition, offset)
	}
}
