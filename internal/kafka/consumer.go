package kafka

import (
	"encoding/json"
	"log"
	"wildberries-tech/internal/cache"
	"wildberries-tech/internal/models"
	"wildberries-tech/internal/repository"

	"github.com/IBM/sarama"
)

type Consumer struct {
	repo  *repository.Repository
	cache *cache.Cache
}

func NewConsumer(repo *repository.Repository, cache *cache.Cache) *Consumer {
	return &Consumer{
		repo:  repo,
		cache: cache,
	}
}

func (c *Consumer) Start() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal("Error creating consumer:", err)
	}

	partitionConsumer, err := consumer.ConsumePartition("orders", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal("Error creating partition consumer:", err)
	}

	log.Println("Kafka consumer started...")

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			c.processMessage(msg.Value)
		case err := <-partitionConsumer.Errors():
			log.Println("Consumer error:", err)
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

	err = c.repo.SaveOrder(order)
	if err != nil {
		log.Println("Error saving to DB:", err)
		return
	}

	c.cache.Set(order.OrderUID, order)
	
	log.Printf("Order %s processed successfully", order.OrderUID)
}