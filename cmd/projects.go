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

type ProjectResponse struct {
	Data struct {
		Team struct {
			ID       string `json:"id"`
			Projects struct {
				Nodes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"projects"`
		} `json:"team"`
	} `json:"data"`
}

//	var projectSelectCmd = &cobra.Command{
//		Use:   "project select",
//		Short: "Select project and save PROJECT_ID to .env",
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

	loadedTeam := os.Getenv("TEAM_ID")
	if loadedTeam == "" {
		panic("TEAM_ID not set")
	}

	url := "https://api.linear.app/graphql"

	projectsQuery := `
	query TeamProjects($id: String!) {
		team(id: $id) {
			projects {
				nodes {
					id
					name
				}
			}
		}
	}`

	projectBody, err := json.Marshal(map[string]any{
		"query": projectsQuery,
		"variables": map[string]string{
			"id": loadedTeam, // this comes from .env TEAM_ID
		},
	})
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(projectBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", loadedKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	projectBody, _ = io.ReadAll(resp.Body)

	// fmt.Println("Raw response:", string(projectBody))

	var result ProjectResponse
	if err := json.Unmarshal(projectBody, &result); err != nil {
		fmt.Println("Error unmarshaling:", err)
		return
	}

	// Collect project names + IDs
	var projectNames []string
	var projectIDs []string
	for _, project := range result.Data.Team.Projects.Nodes {
		projectNames = append(projectNames, project.Name)
		projectIDs = append(projectIDs, project.ID)
	}

	// Variable to hold the user’s selection (project name)
	var selectedProject string

	// Build the select menu
	if err := huh.NewSelect[string]().
		Options(huh.NewOptions(projectNames...)...).
		Value(&selectedProject).
		Title("Project").
		Run(); err != nil {
		fmt.Println("Error in select:", err)
		return
	}

	// Find the ID that corresponds to the selected project
	var selectedProjectID string
	for i, name := range projectNames {
		if name == selectedProject {
			selectedProjectID = projectIDs[i]
			break
		}
	}

	// Read existing .env into a map
	envMap, err := godotenv.Read(".env")
	if err != nil {
		// If the file doesn't exist yet, just start fresh
		envMap = make(map[string]string)
	}

	// Update/insert PROJECT_ID
	envMap["PROJECT_ID"] = selectedProjectID

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

	fmt.Println("Saved PROJECT_ID to .env")

	// Load from .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	// },
}
