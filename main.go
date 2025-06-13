package main

import (
	"fmt"
	"os"

	"github.com/Matthew-K310/linear-cli.git/cmd"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Println("API_KEY environment variable not set!")
		return
	}
	fmt.Println("API Key:", apiKey)
	cmd.Execute()
}
