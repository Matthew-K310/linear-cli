package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/huh"
)

// TODO Implement API call
// TODO Validate input
// TODO Move .env file to .config/linear-cli (possibly, but not SIS)

func main() {
	var apiKey string

	// take user input
	// TODO Hide input field
	huh.NewInput().
		Title("Enter your Linear API key.").
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
}
