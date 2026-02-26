package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"

	"wildberries-tech/internal/cache"
	"wildberries-tech/internal/models"
	"wildberries-tech/internal/repository"
)

type Consumer struct {
	repo  repository.OrderRepository
	cache cache.OrderCache
}

func NewConsumer(repo repository.OrderRepository, cache cache.OrderCache) *Consumer {
	return &Consumer{
		repo:  repo,
		cache: cache,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal("Error creating consumer:", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Println("Error closing consumer:", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition("orders", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal("Error creating partition consumer:", err)
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Println("Error closing partition consumer:", err)
		}
	}()

	log.Println("Kafka consumer started...")

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			c.processMessage(msg.Value)
		case err := <-partitionConsumer.Errors():
			log.Println("Consumer error:", err)
		case <-ctx.Done():
			log.Println("Stopping consumer...")
			return
		}
	}
}

func (c *Consumer) processMessage(data []byte) {
	var order models.Order

	err := json.Unmarshal(data, &order)
	if err != nil {
		log.Println("Error unmarshaling message:", err)
		return
	}

	if err := order.Validate(); err != nil {
		log.Printf("Validation failed for order %s: %v", order.OrderUID, err)
		return
	}

	err = c.repo.SaveOrder(order)
	if err != nil {
		log.Println("Error saving to DB:", err)
		return
	}

	c.cache.Set(order.OrderUID, order)

	log.Printf("Order %s processed successfully", order.OrderUID)
}
