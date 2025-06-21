package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath" // Import the filepath package

	"github.com/joho/godotenv"
)

var apiKey string // Consider making this private if only used within config package

func GetAPIKey() string {
	return apiKey
}

func Load() error {
	// --- MODIFIED SECTION ---

	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Log a warning, but don't necessarily return an error here.
		// The app might still work if API_KEY is in the environment already.
		log.Printf("Warning: Could not get user home directory while loading .env: %v", err)
		// Proceed to check environment variables later even if .env isn't loaded
	} else {
		// Construct the full path to the desired .env file
		configDir := filepath.Join(homeDir, ".config", "linear_cli")
		envFilePath := filepath.Join(configDir, ".env")

		// Attempt to load the .env file from the specific path
		// godotenv.Load() only errors on parsing issues or permission problems if file exists.
		// It does *not* error if the file simply doesn't exist.
		err = godotenv.Load(envFilePath)

		if err != nil {
			// Log a warning if there was an actual error loading the file (e.g., permission denied, parsing error)
			log.Printf("Warning: Error loading .env file from %s: %v", envFilePath, err)
		} else {
			// Optional: Log success if the file was found and loaded
			// log.Printf("Successfully attempted to load environment variables from %s", envFilePath)
		}
	}

	// --- END OF MODIFIED SECTION ---

	// The rest of your function to read from environment variables is correct
	apiKey = os.Getenv("API_KEY")

	if apiKey == "" {
		return fmt.Errorf("API_KEY not set in .env or in env vars")
	}

	// log.Println("APIKey loaded successfully from environment.")
	return nil
}
