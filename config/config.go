// config/config.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MODE            string
	DB_USER         string
	DB_PASSWORD     string
	DB_HOST         string
	DB_PORT         string
	DB_NAME         string
	X_RAPIDAPI_HOST string
	X_RAPIDAPI_KEY  string
}

func Load() Config {
	// ===============================
	// Environment
	// ===============================
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded, using environment variables")
	}

	return Config{
		MODE:            os.Getenv("MODE"),
		DB_USER:         os.Getenv("DB_USER"),
		DB_PASSWORD:     os.Getenv("DB_PASSWORD"),
		DB_HOST:         os.Getenv("DB_HOST"),
		DB_PORT:         os.Getenv("DB_PORT"),
		DB_NAME:         os.Getenv("DB_NAME"),
		X_RAPIDAPI_HOST: os.Getenv("X_RAPIDAPI_HOST"),
		X_RAPIDAPI_KEY:  os.Getenv("X_RAPIDAPI_KEY"),
	}
}
