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
				Nodes []struct {
					ID     string `json:"id"`
					Name   string `json:"name"`
					Issues struct {
						Nodes []struct {
							ID          string `json:"id"`
							Title       string `json:"title"`
							Description string `json:"description"`
							State       struct {
								Name string `json:"name"`
							} `json:"state"`
						} `json:"nodes"`
					} `json:"issues"`
				} `json:"nodes"`
			} `json:"projects"`
		} `json:"team"`
	} `json:"data"`
}

//	var projectSelectCmd = &cobra.Command{
//		Use:   "issue select",
//		Short: "Select issue and save ISSUE_ID to .env",
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
query ProjectIssues($teamId: String!, $projectId: ID!) {
  team(id: $teamId) {
	projects(filter: { id: { eq: $projectId } }) {
      nodes {
        id
        name
        issues {
          nodes {
            id
            title
            description
            state {
              name
            }
          }
        }
      }
    }
  }
}`

	issueBody, err := json.Marshal(map[string]any{
		"query": issuesQuery,
		"variables": map[string]string{
			"teamId":    loadedTeam,
			"projectId": loadedProject,
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

	// Collect all issues across all projects
	var issueNames []string
	var issueIDs []string
	for _, project := range result.Data.Team.Projects.Nodes {
		if project.ID != loadedProject {
			continue
		}
		for _, issue := range project.Issues.Nodes {
			issueNames = append(issueNames, issue.Title)
			issueIDs = append(issueIDs, issue.ID)
		}
	}

	// Variable to hold the user’s selection
	var selectedIssue string

	if err := huh.NewSelect[string]().
		Options(huh.NewOptions(issueNames...)...).
		Value(&selectedIssue).
		Title("Issues").
		Run(); err != nil {
		fmt.Println("Error in select:", err)
		return
	}

	// Find the ID that corresponds to the selected issue
	var selectedIssueID string
	for i, name := range issueNames {
		if name == selectedIssue {
			selectedIssueID = issueIDs[i]
			break
		}
	}

	// Read/modify .env
	envMap, err := godotenv.Read(".env")
	if err != nil {
		envMap = make(map[string]string)
	}

	envMap["ISSUE_ID"] = selectedIssueID

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

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	// },
}
