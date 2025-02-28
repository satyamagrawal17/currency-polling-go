package main

import (
	"context"
	"currency_polling/config"
	"currency_polling/service"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/joho/godotenv"
	"log"
	"time"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

//const (
//	apiURL = "https://openexchangerates.org/api/latest.json"
//	apiKey = "your_api_key_here" // Replace with actual API key
//	topic  = "currency_rates"
//	broker = "localhost:9092"
//)

func main() {
	cfg, err := config.LoadConfig()
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBroker,
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	ctx := context.Background()

	for {
		// Fetch currency exchange rates
		data, err := service.PollAPI(ctx, cfg.APIURL, cfg.APIKey)
		if err != nil {
			log.Printf("Error polling API: %v\n", err)
			time.Sleep(1 * time.Hour) // Retry after 1 hour
			continue
		}

		// Convert data to JSON
		message, err := json.Marshal(data)
		if err != nil {
			log.Printf("Failed to marshal data: %v\n", err)
			continue
		}

		// Send message to Kafka
		deliveryChan := make(chan kafka.Event)
		err = producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &cfg.KafkaTopic, Partition: kafka.PartitionAny},
			Value:          message,
		}, deliveryChan)

		if err != nil {
			log.Printf("Failed to produce message: %v\n", err)
		} else {
			e := <-deliveryChan
			m := e.(*kafka.Message)
			if m.TopicPartition.Error != nil {
				log.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
			} else {
				log.Printf("Message delivered to %v\n", m.TopicPartition)
			}
		}

		close(deliveryChan)

		// Poll API every 1 hour
		time.Sleep(1 * time.Hour)
	}
}
