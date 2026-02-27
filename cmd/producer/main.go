// Package main implements a simple Kafka producer for testing.
package main

import (
	"encoding/json"
	"log"
	"time"

	"wildberries-tech/internal/config"
	"wildberries-tech/internal/models"

	"github.com/IBM/sarama"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, saramaConfig)
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Println("Error closing producer:", err)
		}
	}()

	gofakeit.Seed(0)

	for {
		order := generateOrder()
		data, err := json.Marshal(order)
		if err != nil {
			log.Println("Error marshaling order:", err)
			continue
		}

		message := &sarama.ProducerMessage{
			Topic: cfg.Kafka.Topic,
			Value: sarama.StringEncoder(data),
		}

		partition, offset, err := producer.SendMessage(message)
		if err != nil {
			log.Println("Error sending message:", err)
		} else {
			log.Printf("Sent order %s to partition %d at offset %d\n", order.OrderUID, partition, offset)
		}

		time.Sleep(5 * time.Second)
	}
}

func generateOrder() models.Order {
	return models.Order{
		OrderUID:    gofakeit.UUID(),
		TrackNumber: gofakeit.Numerify("WBILM##########"),
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Address().Address,
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: models.Payment{
			Transaction:  gofakeit.UUID(),
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       gofakeit.Number(100, 10000),
			PaymentDt:    int(time.Now().Unix()),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      gofakeit.Number(100000, 999999),
				TrackNumber: gofakeit.Numerify("WBILM##########"),
				Price:       gofakeit.Number(100, 5000),
				Rid:         gofakeit.UUID(),
				Name:        gofakeit.ProductName(),
				Sale:        gofakeit.Number(0, 50),
				Size:        "0",
				TotalPrice:  317,
				NmID:        gofakeit.Number(1000000, 9999999),
				Brand:       gofakeit.Company(),
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}
