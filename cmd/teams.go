package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
)

type TeamResponse struct {
	Data struct {
		Teams struct {
			Nodes []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"nodes"`
		} `json:"teams"`
	} `json:"data"`
}

//	var teamSelectCmd = &cobra.Command{
//		Use:   "team select",
//		Short: "Select team and save TEAM_ID to .env",
//		Run: func(cmd *cobra.Command, args []string) {
func main() {
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

	teamsQuery := `
	query Teams {
		teams {
			nodes {
				id
				name
			}
		}
	}
    `

	// request body as JSON
	jsonBody := fmt.Appendf(nil, `{"query": %q}`, teamsQuery)

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

	body, _ := io.ReadAll(resp.Body)

	var result TeamResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error unmarshaling:", err)
		return
	}

	// Collect team names + IDs
	var teamNames []string
	var teamIDs []string
	for _, team := range result.Data.Teams.Nodes {
		teamNames = append(teamNames, team.Name)
		teamIDs = append(teamIDs, team.ID)
	}

	// Variable to hold the user’s selection (team name)
	var selectedTeam string

	// Build the select menu
	if err := huh.NewSelect[string]().
		Options(huh.NewOptions(teamNames...)...).
		Value(&selectedTeam).
		Title("Team").
		Run(); err != nil {
		fmt.Println("Error in select:", err)
		return
	}

	// Find the ID that corresponds to the selected team
	var selectedTeamID string
	for i, name := range teamNames {
		if name == selectedTeam {
			selectedTeamID = teamIDs[i]
			break
		}
	}

	// Read existing .env into a map
	envMap, err := godotenv.Read(".env")
	if err != nil {
		// If the file doesn't exist yet, just start fresh
		envMap = make(map[string]string)
	}

	// Update/insert TEAM_ID
	envMap["TEAM_ID"] = selectedTeamID

	// Rewrite the .env file with all keys
	file, err := os.Create(".env")
	if err != nil {
		log.Fatalf("failed to create .env file: %v", err)
	}
	defer file.Close()

	for k, v := range envMap {
		_, err := fmt.Fprintf(file, "%s=%s\n", k, v)
		if err != nil {
			log.Fatalf("failed to write to .env file: %v", err)
		}
	}

	fmt.Println("Saved TEAM_ID to .env")

	// Load from .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	// },
}
