package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings" // Import strings package for validation

	"github.com/manifoldco/promptui" // Import the promptui package
	"github.com/spf13/cobra"

	"github.com/Matthew-K310/linear-cli/api"
	"github.com/Matthew-K310/linear-cli/config"
)

type promptContent struct {
	errorMsg string
	label    string
}

// Structs specific to the issues query response (kept here as listCmd uses them)
type IssueNode struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"state"`
	Assignee *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"assignee"`
	Team struct { // Include team details if listing issues from multiple teams
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	CreatedAt string `json:"createdAt"`
}

type IssuesConnection struct {
	Nodes []IssueNode `json:"nodes"`
}

type IssuesResponseData struct {
	Issues IssuesConnection `json:"issues"`
}

type IssueCreateResponseData struct {
	IssueCreate struct {
		Success bool `json:"success"`
		Issue   struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"issue"`
	} `json:"issueCreate"`
}

var issuesRootCmd = &cobra.Command{
	Use:   "issues",
	Short: "Manage Linear issues",
	Long:  `Provides commands to create, list, and modify Linear issues.`,
}

// listCmd represents the list subcommand for issues
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Linear issues",
	Long:  `Lists issues. Can be filtered by flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		teamID, _ := cmd.Flags().GetString("team")
		stateType, _ := cmd.Flags().GetString("state-type")
		limit, _ := cmd.Flags().GetInt("limit")

		// Updated query to use variables for filtering
		query := `
		query Issue($teamId: String, $stateType: String, $first: Int) {
			issues(filter: {teamId: {eq: $teamId}, state: {type: {eq: $stateType}}}, first: $first) {
				nodes {
					id
					title
					description
					state {
						id
						name
						type
					}
					assignee {
						id
						name
					}
					team {
						id
						name
					}
					createdAt
				}
			}
		}
		`

		// Define variables based on flags
		variables := map[string]interface{}{}
		if teamID != "" {
			variables["teamId"] = teamID
		}
		if stateType != "" {
			variables["stateType"] = stateType
		}
		if limit > 0 {
			variables["first"] = limit // Use 'first' for limiting results in GraphQL
		} else {
			// Set a default limit if none is provided, otherwise the API might return too many
			variables["first"] = 50 // Example default limit
		}

		apiKey := config.GetAPIKey()
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: API_KEY is not set.")
			os.Exit(1)
		}

		fmt.Println("Fetching issues...")

		// Pass the query and variables to the API request
		data, err := api.MakeGraphQLRequest(apiKey, query, variables)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error making GraphQL request: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("GraphQL request successful!")

		// Unmarshal the specific response data structure
		var issuesResponse IssuesResponseData
		if err := json.Unmarshal(data, &issuesResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling issues data: %v\n", err)
			os.Exit(1)
		}

		// Process and print the data
		fmt.Printf("\nFound %d issues:\n", len(issuesResponse.Issues.Nodes))
		if len(issuesResponse.Issues.Nodes) > 0 {
			fmt.Println("--------------------")
			for _, issue := range issuesResponse.Issues.Nodes {
				fmt.Printf("  Issue ID: %s\n", issue.ID)
				fmt.Printf("  Title: %s\n", issue.Title)
				fmt.Printf("  Team: %s (%s)\n", issue.Team.Name, issue.Team.ID)
				fmt.Printf("  State: %s (Type: %s)\n", issue.State.Name, issue.State.Type)
				if issue.Assignee != nil {
					fmt.Printf("  Assignee: %s (%s)\n", issue.Assignee.Name, issue.Assignee.ID)
				} else {
					fmt.Println("  Assignee: Unassigned")
				}
				fmt.Printf("  Created At: %s\n", issue.CreatedAt)
				fmt.Println("--------------------")
			}
		}
	},
}

// createCmd represents the create subcommand for issues
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Linear issue interactively",
	Long:  `Interactively prompts for details to create a new Linear issue.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Use promptui to get input for each field

		// Prompt for Title (Required)
		titlePrompt := promptui.Prompt{
			Label: "Issue Title",
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("title cannot be empty")
				}
				return nil
			},
		}
		title, err := titlePrompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
			if err == promptui.ErrInterrupt { // Handle user interruption (Ctrl+C)
				os.Exit(0)
			}
			os.Exit(1)
		}

		// Prompt for Description (Optional)
		descriptionPrompt := promptui.Prompt{
			Label: "Issue Description (Optional)",
			// No validation needed as description is optional
		}
		description, err := descriptionPrompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			os.Exit(1)
		}
		// Use empty string if user just pressed enter for description
		if strings.TrimSpace(description) == "" {
			description = ""
		}

		// Prompt for Team ID (Required)
		teamIDPrompt := promptui.Prompt{
			Label: "Team ID",
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("team ID cannot be empty")
				}
				// You could add more complex validation here, e.g., regex for format
				return nil
			},
		}
		teamID, err := teamIDPrompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			os.Exit(1)
		}

		// 2. Define the GraphQL Mutation Query
		// Use variables ($title, $description, $teamId) to pass dynamic data
		// REMOVED THE COMMENT LINE STARTING WITH //
		mutation := `
		mutation CreateIssue($title: String!, $description: String, $teamId: String!) {
			issueCreate(
				input: {
					title: $title
					description: $description
					teamId: $teamId
				}
			) {
				success
				issue {
					id
					title
					# You could request more fields of the created issue here if needed
				}
				# errors { message } // Consider adding error handling details
			}
		}
		`
		// NOTE: I changed the comment to # just to show how GraphQL comments look,
		// but removing the line entirely is also perfectly fine and perhaps cleaner.
		// If your api.MakeGraphQLRequest client is sensitive to *any* non-query text,
		// removing the line is the safest bet. Let's remove it to be safe:

		mutation = `
		mutation CreateIssue($title: String!, $description: String, $teamId: String!) {
			issueCreate(
				input: {
					title: $title
					description: $description
					teamId: $teamId
				}
			) {
				success
				issue {
					id
					title
				}
				# errors { message }
			}
		}
		`

		// 3. Define Variables Map from prompt values
		variables := map[string]interface{}{
			"title":  title,
			"teamId": teamID,
		}
		// Add description only if it's not empty after the prompt
		if description != "" {
			variables["description"] = description
		}

		apiKey := config.GetAPIKey()
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: API_KEY is not set.")
			os.Exit(1)
		}

		fmt.Println("Creating issue...")

		// 4. Call API to execute the mutation
		data, err := api.MakeGraphQLRequest(apiKey, mutation, variables)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error making GraphQL request: %v\n", err)
			// Depending on your API client, you might need to unmarshal 'data'
			// even on error to get specific GraphQL errors.
			os.Exit(1)
		}

		fmt.Println("GraphQL mutation executed.")

		// 5. Unmarshal and Process the Response
		var createResponse IssueCreateResponseData
		if err := json.Unmarshal(data, &createResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling create issue data: %v\n", err)
			os.Exit(1)
		}

		if createResponse.IssueCreate.Success {
			fmt.Println("Issue created successfully!")
			fmt.Printf("  ID: %s\n", createResponse.IssueCreate.Issue.ID)
			fmt.Printf("  Title: %s\n", createResponse.IssueCreate.Issue.Title)
		} else {
			fmt.Fprintln(os.Stderr, "Error creating issue:")
			// TODO: Add logic to print specific errors from the API response
			// if the API returns an 'errors' field in the response payload
			fmt.Fprintf(os.Stderr, "  API reported success: false\n") // Fallback if no error details are unmarshaled
			os.Exit(1)
		}
	},
}

// modifyCmd represents the modify subcommand for issues
var modifyCmd = &cobra.Command{
	Use:   "modify [issue-id]",
	Short: "Modify an existing Linear issue",
	Long:  `Modifies an existing Linear issue identified by its ID.`,
	Args:  cobra.ExactArgs(1), // Requires exactly one argument: the issue ID
	Run: func(cmd *cobra.Command, args []string) {
		issueID := args[0]
		// TODO: Implement logic for modifying an issue
		// This will involve defining flags for fields to modify (title, state, assignee, etc.)
		// and making a GraphQL mutation request using the issueID.
		fmt.Printf("modify issue called for ID: %s\n", issueID)
		fmt.Println("TODO: Implement issue modification logic")
	},
}

func init() {
	// Add the issuesRootCmd to the main application root command
	// Assuming 'rootCmd' is defined elsewhere as your application's entry point
	rootCmd.AddCommand(issuesRootCmd)

	// Add the subcommands to the issuesRootCmd
	issuesRootCmd.AddCommand(listCmd)
	issuesRootCmd.AddCommand(createCmd) // createCmd now uses promptui
	issuesRootCmd.AddCommand(modifyCmd)

	// Define flags specifically for the 'list' command
	listCmd.Flags().StringP("team", "t", "", "Filter issues by Team ID")
	listCmd.Flags().
		StringP("state-type", "s", "", "Filter issues by State Type (e.g., 'started', 'completed')")
	listCmd.Flags().IntP("limit", "l", 0, "Limit the number of results")

	// Remove flags for 'createCmd' as it now uses promptui for input
	// createCmd.Flags().StringP("title", "T", "", "Title of the new issue")
	// createCmd.Flags().StringP("description", "d", "", "Description of the new issue (optional)")
	// createCmd.Flags().StringP("team-id", "", "", "Team ID for the new issue")
	// createCmd.MarkFlagRequired("title") // Also remove required marks
	// createCmd.MarkFlagRequired("team-id")

	// TODO: Add flags for 'modifyCmd' (e.g., --title, --state-id, --assignee-id, etc.)
	// modifyCmd.Flags().StringP("title", "T", "", "New title for the issue")
	// modifyCmd.Flags().StringP("state-id", "S", "", "New state ID for the issue")
}
