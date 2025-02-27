package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	APIURL       string
	KafkaBroker  string
	KafkaTopic   string
	APIKey       string
	CronSchedule string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfg := &Config{
		APIURL:       os.Getenv("API_URL"),
		KafkaBroker:  os.Getenv("KAFKA_BROKER"),
		KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
		APIKey:       os.Getenv("API_KEY"),
		CronSchedule: os.Getenv("CRON_SCHEDULE"),
	}

	if cfg.APIURL == "" || cfg.KafkaBroker == "" || cfg.KafkaTopic == "" || cfg.APIKey == "" || cfg.CronSchedule == "" {
		return nil, fmt.Errorf("environment variables not set")
	}

	return cfg, nil
}
