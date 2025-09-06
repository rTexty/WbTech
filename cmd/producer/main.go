package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/IBM/sarama"
)

func main() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal("Error creating producer:", err)
	}
	defer producer.Close()

	data, err := os.ReadFile("model.json")
	if err != nil {
		log.Fatal("Error reading model.json:", err)
	}

	var testOrder interface{}
	if err := json.Unmarshal(data, &testOrder); err != nil {
		log.Fatal("Invalid JSON in model.json:", err)
	}
	message := &sarama.ProducerMessage{
		Topic: "orders",
		Value: sarama.StringEncoder(data),
	}

	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		log.Fatal("Error sending message:", err)
	}

	log.Printf("Message sent successfully! Partition: %d, Offset: %d", partition, offset)
	log.Printf("Sent order with UID: %s", getOrderUID(data))
}

func getOrderUID(data []byte) string {
	var order map[string]interface{}
	json.Unmarshal(data, &order)
	if uid, ok := order["order_uid"].(string); ok {
		return uid
	}
	return "unknown"
}