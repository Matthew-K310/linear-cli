package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
)

// func AuthPrompt() {
func main() {
	var apiKey string

	// take user input
	// TODO Hide input field
	huh.NewInput().
		Title("Enter your Credentials").
		Prompt("API Key:").
		// Password(true).
		EchoMode(huh.EchoModePassword).
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

	fmt.Println("Saved API_KEY to .env")

	// Load from .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Retrieve variable (for debugging)
	loadedKey := os.Getenv("API_KEY")
	if loadedKey == "" {
		panic("API_KEY not set")
	}
	// fmt.Println("Loaded API key from .env:", loadedKey)

	url := "https://api.linear.app/graphql"

	query := `{
		viewer {
			id
			name
		}
	}`

	// request body as JSON
	jsonBody := fmt.Appendf(nil, `{"query": %q}`, query)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		panic(err)
	}

	// headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", loadedKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// body, _ := io.ReadAll(resp.Body)
	// fmt.Println("Response:", string(body))
}
