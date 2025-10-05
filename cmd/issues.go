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

type IssueResponse struct {
	Data struct {
		Team struct {
			ID       string `json:"id"`
			Projects struct {
				ID     string `json:"id"`
				Issues struct {
					Nodes []struct {
						ID          string `json:"id"`
						Title       string `json:"title"`
						Description string `json:"description"`
						State       struct {
							Name string `json:"name"`
							Type string `json:"type"`
						} `json:"state"`
					} `json:"nodes"`
				} `json:"issues"`
			} `json:"projects"`
		} `json:"teams"`
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

	loadedProject := os.Getenv("PROJECT_ID")
	if loadedProject == "" {
		panic("PROJECT_ID not set")
	}

	url := "https://api.linear.app/graphql"

	issuesQuery := `
	query ProjectIssues($projectid: String!) {
		team(id: $id){
			project(id: $projectid) {
				issues {
					nodes {
						id
						title
						state {
							name
						}
					}
				}
			}
		}
   }`

	issueBody, err := json.Marshal(map[string]any{
		"query": issuesQuery,
		"variables": map[string]string{
			"id":        loadedTeam,    // this comes from .env TEAM_ID
			"projectid": loadedProject, // this comes from .env TEAM_ID
		},
	})
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(issueBody))
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

	issueBody, _ = io.ReadAll(resp.Body)

	fmt.Println("Raw response:", string(issueBody))

	var result IssueResponse
	if err := json.Unmarshal(issueBody, &result); err != nil {
		fmt.Println("Error unmarshaling:", err)
		return
	}

	// Collect project names + IDs
	var issueNames []string
	var issueIDs []string
	for _, issue := range result.Data.Team.Projects.Issues.Nodes {
		issueNames = append(issueNames, issue.Title)
		issueIDs = append(issueIDs, issue.ID)
	}

	// Variable to hold the user’s selection (project name)
	var selectedIssue string

	// Build the select menu
	if err := huh.NewSelect[string]().
		Options(huh.NewOptions(issueNames...)...).
		Value(&selectedIssue).
		Title("Issues").
		Run(); err != nil {
		fmt.Println("Error in select:", err)
		return
	}

	// Find the ID that corresponds to the selected project
	var selectedIssueID string
	for i, name := range issueNames {
		if name == selectedIssue {
			selectedIssueID = issueIDs[i]
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
	envMap["ISSUE_ID"] = selectedIssueID

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

	fmt.Println("Saved ISSUE_ID to .env")

	// Load from .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	// },
}
