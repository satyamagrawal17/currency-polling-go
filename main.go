package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"

	"currency_polling/config"
	"github.com/robfig/cron/v3"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	kafkaWriter := setupKafkaWriter(cfg.KafkaBroker, cfg.KafkaTopic)
	defer kafkaWriter.Close()

	c := cron.New()
	_, err = c.AddFunc(cfg.CronSchedule, func() {
		runPollingJob(cfg, kafkaWriter)
	})
	if err != nil {
		log.Fatalf("Cron error: %v", err)
	}

	c.Start()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	c.Stop()
	log.Println("Shutdown complete.")
}
