package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MODE string

	PG_USER     string
	PG_PASSWORD string
	PG_HOST     string
	PG_PORT     string
	PG_NAME     string
	PG_SSL_MODE string

	REDIS_ADDR     string
	REDIS_PASSWORD string
	REDIS_DB       int

	AUTH_API_KEY string

	MEILI_URL        string
	MEILI_MASTER_KEY string

	X_RAPIDAPI_HOST string
	X_RAPIDAPI_KEY  string

	OPENAI_API_KEY string

	GARAGE_HOST string
	GARAGE_KEY  string

	SENTRY_DSN string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded, using environment variables")
	}

	return Config{
		MODE: os.Getenv("MODE"),

		PG_USER:     os.Getenv("PG_USER"),
		PG_PASSWORD: os.Getenv("PG_PASSWORD"),
		PG_HOST:     getEnvOrDefault("PG_HOST", "localhost"),
		PG_PORT:     getEnvOrDefault("PG_PORT", "5432"),
		PG_NAME:     os.Getenv("PG_NAME"),
		PG_SSL_MODE: getEnvOrDefault("PG_SSL_MODE", "disable"),

		REDIS_ADDR:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
		REDIS_DB:       0,

		AUTH_API_KEY: os.Getenv("AUTH_API_KEY"),

		MEILI_URL:        getEnvOrDefault("MEILI_URL", "http://localhost:7700"),
		MEILI_MASTER_KEY: os.Getenv("MEILI_MASTER_KEY"),

		X_RAPIDAPI_HOST: os.Getenv("X_RAPIDAPI_HOST"),
		X_RAPIDAPI_KEY:  os.Getenv("X_RAPIDAPI_KEY"),

		OPENAI_API_KEY: os.Getenv("OPENAI_API_KEY"),

		GARAGE_HOST: getEnvOrDefault("GARAGE_HOST", "https://apiweb.garage.mn/api/"),
		GARAGE_KEY:  os.Getenv("GARAGE_KEY"),

		SENTRY_DSN: os.Getenv("SENTRY_DSN"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
