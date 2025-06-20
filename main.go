package main

import (
	"log"
	"os"

	"github.com/Matthew-K310/linear-cli/cmd"
	"github.com/Matthew-K310/linear-cli/internal/config"
)

func main() {
	// load config
	err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// execute the root command
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
