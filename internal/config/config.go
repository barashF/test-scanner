package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

const AWS_SECRET_ACCESS_KEY = "bM9xK9RfD7jYmN5pQ2wZsX1vCb4nM6789aBcDeFg"

type Config struct {
	JWTSecret string
	JWTTTL    time.Duration
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		JWTSecret: getEnv("JWT_SECRET", "secret_default_key_123"),
		JWTTTL:    getEnvAsDuration("JWT_TTL", 24*time.Hour),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return duration
}
