package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config - конфиг приложения.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	LogLevel    string

	PaymentBaseURL string
	PaymentTimeout time.Duration
	PaymentRetries int
}

// LoadConfig загружаем конфиг из переменных окружения. Ошибка если нет обязательного поля.
func LoadConfig() (Config, error) {
	cfg := Config{
		HTTPAddr:       getenvOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		LogLevel:       getenvOrDefault("LOG_LEVEL", "info"),
		PaymentBaseURL: os.Getenv("PAYMENT_BASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.PaymentBaseURL == "" {
		return Config{}, fmt.Errorf("PAYMENT_BASE_URL is required")
	}

	timeout, err := time.ParseDuration(getenvOrDefault("PAYMENT_TIMEOUT", "2s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse PAYMENT_TIMEOUT: %w", err)
	}
	cfg.PaymentTimeout = timeout

	retries, err := strconv.Atoi(getenvOrDefault("PAYMENT_RETRIES", "3"))
	if err != nil {
		return Config{}, fmt.Errorf("parse PAYMENT_RETRIES: %w", err)
	}
	cfg.PaymentRetries = retries

	return cfg, nil
}

// getenvOrDefault получает из переменных окружения переменную или возвращает дефолтное значение.
func getenvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}
