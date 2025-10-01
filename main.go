package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
)

// TODO Implement API call
// TODO Validate input
// TODO Move .env file to .config/linear-cli (possibly, but not SIS)

func main() {
	var apiKey string

	// take user input
	huh.NewInput().
		Title("Enter your Credentials").
		Prompt("API Key:").
		Password(true).
		Value(&apiKey).
		Run()

	// save to .env file for authentication
	file, err := os.Create(".env")
	if err != nil {
		log.Fatalf("failed to create .env file: %v", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "API_KEY=%s\n", apiKey)
	if err != nil {
		log.Fatalf("failed to write to .env file: %v", err)
	}

	fmt.Println("Saved to .env")

	// Load from .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Retrieve variable
	loadedKey := os.Getenv("API_KEY")
	// fmt.Println("Loaded API key from .env:", loadedKey) // for debugging .env reading
}
