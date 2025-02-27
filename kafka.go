package main

import (
	"context"
	"currency_polling/config"
	"currency_polling/service"
	"encoding/json"
	"fmt"
	"log"
	_ "log"

	"github.com/segmentio/kafka-go"
)

func setupKafkaWriter(broker, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func runPollingJob(cfg *config.Config, writer *kafka.Writer) {
	ctx := context.Background() // Create a context for the polling job.
	data, err := service.PollAPI(ctx, cfg.APIURL, cfg.APIKey)
	if err != nil {
		log.Printf("Polling error: %v", err)
		return
	}

	if err := publishToKafka(ctx, writer, data); err != nil {
		log.Printf("Kafka publishing error: %v", err)
		return
	}

	log.Println("Data polled and published to Kafka")
}

func publishToKafka(ctx context.Context, w *kafka.Writer, data *service.OpenExchangeRatesResponse) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	err = w.WriteMessages(ctx, kafka.Message{
		Value: jsonData,
	})
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	return nil
}
