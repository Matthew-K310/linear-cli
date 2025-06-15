package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var apiKey string

func GetAPIKey() string {
	return apiKey
}

func Load() error {
	err := godotenv.Load()
	if err != nil {
		log.Printf(
			"Warning: Error loading .env file: %v.\n", err,
		)
	}

	apiKey = os.Getenv("API_KEY")

	if apiKey == "" {
		return fmt.Errorf("API_KEY not set in .env or in env vars")
	}

	log.Println("APIKey loaded successfully from environment.")
	return nil
}
