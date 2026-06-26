package config

import (
	"fmt"
	"os"
)

// Config - конфиг приложения.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	LogLevel    string
}

// LoadConfig загружаем конфиг из переменных окружения. Ошибка если нет обязательного поля.
func LoadConfig() (Config, error) {
	cfg := Config{
		HTTPAddr:    getenvOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		LogLevel:    getenvOrDefault("LOG_LEVEL", "info"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getenvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}
