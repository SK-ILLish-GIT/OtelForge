package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	RabbitMQURL       string
	JWTSecret         string
	EncryptionKey     string
	APIAddr           string
	CORSOrigin        string
	WorkerConcurrency int
}

func Load() (Config, error) {
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		return Config{}, fmt.Errorf("DATABASE_URL required")
	}
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		return Config{}, fmt.Errorf("ENCRYPTION_KEY required")
	}
	wc, _ := strconv.Atoi(envOr("WORKER_CONCURRENCY", "2"))
	if wc < 1 {
		wc = 1
	}
	return Config{
		DatabaseURL:       db,
		RabbitMQURL:       envOr("RABBITMQ_URL", "amqp://otel:otel@localhost:5672/"),
		JWTSecret:         envOr("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		EncryptionKey:     key,
		APIAddr:           envOr("API_ADDR", ":8080"),
		CORSOrigin:        envOr("CORS_ORIGIN", "http://localhost:5173"),
		WorkerConcurrency: wc,
	}, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
