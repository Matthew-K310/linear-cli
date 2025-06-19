package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/Matthew-K310/linear-cli/api"
	"github.com/Matthew-K310/linear-cli/config"
)

type promptContent struct {
	errorMsg string
	label    string
}

// Define the structure of an issues
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
	Team struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
}

type IssuesConnection struct {
	Nodes []IssueNode `json:"nodes"`
}

type IssuesResponseData struct {
	Issues IssuesConnection `json:"issues"`
}

// Define the structure of the issues response
type IssueCreateResponseData struct {
	IssueCreate struct {
		Success bool `json:"success"`
		Issue   struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Assignee    struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"assignee"`
			State struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"state"`
		} `json:"issue"`
	} `json:"issueCreate"`
}

// Define the elements
type TeamNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TeamsConnection struct {
	Nodes []TeamNode `json:"nodes"`
}

type TeamsResponseData struct {
	Teams TeamsConnection `json:"teams"`
}

// Represents a single user/member node
type UserNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Represents the connection of user/member nodes
type UserConnection struct {
	Nodes []UserNode `json:"nodes"`
}

// Corrected TeamMembersResponseData struct using 'members' field
type TeamMembersResponseData struct {
	Team struct {
		ID      string         `json:"id"`
		Name    string         `json:"name"`
		Members UserConnection `json:"members"` // <-- Use 'members' to get assignable users
	} `json:"team"`
}

type StateNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type StateConnection struct {
	Nodes []StateNode `json:"nodes"`
}

type TeamStatesResponseData struct {
	Team struct {
		ID     string          `json:"id"`
		Name   string          `json:"name"`
		States StateConnection `json:"states"`
	} `json:"team"`
}

// root "issues" command
var issuesRootCmd = &cobra.Command{
	Use:   "issues",
	Short: "Manage Linear issues",
	Long:  `Provides commands to create, list, and modify Linear issues.`,
}

// "issues list" command for listing issues
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Linear issues",
	Long:  `Lists issues. Can be filtered by flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		teamNameFromFlag, _ := cmd.Flags().GetString("team")
		stateType, _ := cmd.Flags().GetString("state-type")
		limit, _ := cmd.Flags().GetInt("limit")

		apiKey := config.GetAPIKey()
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: API_KEY is not set.")
			os.Exit(1)
		}

		teamID := ""

		if teamNameFromFlag != "" {
			teamQuery := `
			query GetTeamIdByName($teamName: String!) {
				teams(filter: {name: {eq: $teamName}}) {
					nodes {
						id
						name
					}
				}
			}
			`
			teamVars := map[string]interface{}{
				"teamName": teamNameFromFlag,
			}

			fmt.Printf("Looking up Team ID for name '%s'...\n", teamNameFromFlag)
			teamData, err := api.MakeGraphQLRequest(apiKey, teamQuery, teamVars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error looking up team '%s': %v\n", teamNameFromFlag, err)
				os.Exit(1)
			}

			var teamsResponse TeamsResponseData
			if err := json.Unmarshal(teamData, &teamsResponse); err != nil {
				fmt.Fprintf(os.Stderr, "Error unmarshalling team data: %v\n", err)
				os.Exit(1)
			}

			if len(teamsResponse.Teams.Nodes) == 0 {
				fmt.Fprintf(os.Stderr, "Error: Team '%s' not found.\n", teamNameFromFlag)
				os.Exit(1)
			} else if len(teamsResponse.Teams.Nodes) > 1 {
				fmt.Fprintf(os.Stderr, "Error: Multiple teams found with name '%s'. Please use the Team ID.\n", teamNameFromFlag)
				os.Exit(1)
			} else {
				teamID = teamsResponse.Teams.Nodes[0].ID
				fmt.Printf("Found Team ID: %s\n", teamID)
			}
		}

		query := `
		query Issue($teamId: ID, $stateType: String, $first: Int) {
			issues(filter: {team: {id: {eq: $teamId}}, state: {type: {eq: $stateType}}}, first: $first) {
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
						name
						id
					}
					team {
						id
						name
					}
				}
			}
		}
		`

		variables := map[string]interface{}{}
		if teamID != "" {
			variables["teamId"] = teamID
		}
		if stateType != "" {
			variables["stateType"] = stateType
		}
		if limit > 0 {
			variables["first"] = limit
		} else {
			variables["first"] = 50
		}

		fmt.Println("Fetching issues...")

		data, err := api.MakeGraphQLRequest(apiKey, query, variables)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error making GraphQL request: %v\n", err)
			os.Exit(1)
		}

		var issuesResponse IssuesResponseData
		if err := json.Unmarshal(data, &issuesResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling issues data: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nFound %d issues:\n", len(issuesResponse.Issues.Nodes))
		if len(issuesResponse.Issues.Nodes) > 0 {
			fmt.Println("--------------------")
			for _, issue := range issuesResponse.Issues.Nodes {
				fmt.Printf("  Issue ID: %s\n", issue.ID)
				fmt.Printf("  Title: %s\n", issue.Title)
				fmt.Printf("  Description: %s\n", issue.Description)
				fmt.Printf("  Team: %s (%s)\n", issue.Team.Name, issue.Team.ID)
				fmt.Printf("  State: %s (Type: %s)\n", issue.State.Name, issue.State.Type)
				if issue.Assignee != nil {
					fmt.Printf("  Assignee: %s (%s)\n", issue.Assignee.Name, issue.Assignee.ID)
				} else {
					fmt.Println("  Assignee: Unassigned")
				}
				fmt.Println("--------------------")
			}
		}
	},
}

// "issues create" command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Linear issue interactively",
	Long:  `Interactively prompts for details to create a new Linear issue.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := config.GetAPIKey()
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: API_KEY is not set.")
			os.Exit(1)
		}

		// Prompt for issue title
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
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			os.Exit(1)
		}

		// Prompt for issue description (optional)
		descriptionPrompt := promptui.Prompt{
			Label: "Issue Description (Optional)",
		}
		description, err := descriptionPrompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed %v\n", err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			os.Exit(1)
		}
		if strings.TrimSpace(description) == "" {
			description = ""
		}

		// Query teams to select from
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

		fmt.Println("Fetching teams...")
		teamsData, err := api.MakeGraphQLRequest(apiKey, teamsQuery, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching teams: %v\n", err)
			os.Exit(1)
		}

		var teamsResponse TeamsResponseData
		if err := json.Unmarshal(teamsData, &teamsResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling teams data: %v\n", err)
			os.Exit(1)
		}

		if len(teamsResponse.Teams.Nodes) == 0 {
			fmt.Fprintln(os.Stderr, "No teams found. Cannot create issue.")
			os.Exit(1)
		}

		teamNames := []string{}
		teamMap := make(map[string]string)
		for _, team := range teamsResponse.Teams.Nodes {
			teamNames = append(teamNames, team.Name)
			teamMap[team.Name] = team.ID
		}

		// Prompt to select team
		teamSelectPrompt := promptui.Select{
			Label: "Select Team",
			Items: teamNames,
			Searcher: func(input string, index int) bool {
				item := teamNames[index]
				return strings.Contains(strings.ToLower(item), strings.ToLower(input))
			},
		}

		_, selectedTeamName, err := teamSelectPrompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Team selection failed %v\n", err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			os.Exit(1)
		}

		selectedTeamID := teamMap[selectedTeamName]
		fmt.Printf("Selected Team: %s (ID: %s)\n", selectedTeamName, selectedTeamID)

		// Query members (assignees) for the selected team
		assigneesQuery := `
query TeamMembers($teamId: String!) {
  team(id: $teamId) {
    members {
      nodes {
        id
        name
      }
    }
  }
}
`

		assigneesVariables := map[string]interface{}{
			"teamId": selectedTeamID,
		}

		assigneesData, err := api.MakeGraphQLRequest(apiKey, assigneesQuery, assigneesVariables)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Error fetching assignees for team %s: %v\n",
				selectedTeamID,
				err,
			)
			os.Exit(1)
		}

		var teamMembersResponse TeamMembersResponseData
		if err := json.Unmarshal(assigneesData, &teamMembersResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling assignees data: %v\n", err)
			os.Exit(1)
		}

		assigneeNames := []string{"Unassigned"}
		assigneeMap := make(map[string]string)
		assigneeMap["Unassigned"] = ""

		if len(teamMembersResponse.Team.Members.Nodes) > 0 {
			for _, member := range teamMembersResponse.Team.Members.Nodes {
				assigneeNames = append(assigneeNames, member.Name)
				assigneeMap[member.Name] = member.ID
			}
		} else {
			fmt.Fprintln(os.Stderr, "No members found for the selected team, only 'Unassigned' option available.")
		}

		// Prompt to select assignee
		assigneeSelectPrompt := promptui.Select{
			Label: "Select Assignee",
			Items: assigneeNames,
			Searcher: func(input string, index int) bool {
				item := assigneeNames[index]
				return strings.Contains(strings.ToLower(item), strings.ToLower(input))
			},
		}

		_, selectedAssigneeNameFromPrompt, err := assigneeSelectPrompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Assignee selection failed %v\n", err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			os.Exit(1)
		}

		selectedAssigneeID := assigneeMap[selectedAssigneeNameFromPrompt]
		fmt.Printf(
			"Selected Assignee: %s (ID: %s)\n",
			selectedAssigneeNameFromPrompt,
			selectedAssigneeID,
		)

		// Query states (statuses) for the selected team
		statesQuery := `
    query TeamStates($teamId: String!) {
      team(id: $teamId) {
        id
        name
        states {
          nodes {
            id
            name
            type
          }
        }
      }
    }
    `

		statesVariables := map[string]interface{}{
			"teamId": selectedTeamID,
		}

		fmt.Println("Fetching possible statuses for the selected team...")
		statesData, err := api.MakeGraphQLRequest(apiKey, statesQuery, statesVariables)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching states for team %s: %v\n", selectedTeamID, err)
			os.Exit(1)
		}

		var teamStatesResponse TeamStatesResponseData
		if err := json.Unmarshal(statesData, &teamStatesResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling states data: %v\n", err)
			os.Exit(1)
		}

		if len(teamStatesResponse.Team.States.Nodes) == 0 {
			fmt.Fprintln(os.Stderr, "No statuses found for the selected team. Cannot set status.")
			os.Exit(1)
		}

		stateNames := []string{}
		stateMap := make(map[string]string)
		for _, state := range teamStatesResponse.Team.States.Nodes {
			stateNames = append(stateNames, state.Name)
			stateMap[state.Name] = state.ID
		}

		// Prompt to select issue status
		stateSelectPrompt := promptui.Select{
			Label: "Select Status",
			Items: stateNames,
			Searcher: func(input string, index int) bool {
				item := stateNames[index]
				return strings.Contains(strings.ToLower(item), strings.ToLower(input))
			},
		}

		_, selectedStateNameFromPrompt, err := stateSelectPrompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Status selection failed %v\n", err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			os.Exit(1)
		}

		selectedStateID := stateMap[selectedStateNameFromPrompt]
		fmt.Printf("Selected Status: %s (ID: %s)\n", selectedStateNameFromPrompt, selectedStateID)

		// Prepare assigneeId for mutation, nil if unassigned
		var assigneeID interface{}
		if selectedAssigneeID == "" {
			assigneeID = nil
		} else {
			assigneeID = selectedAssigneeID
		}

		// Issue creation mutation
		mutation := `
mutation CreateIssue(
  $title: String!, 
  $description: String, 
  $teamId: String!, 
  $assigneeId: String, 
  $stateId: String
) {
  issueCreate(
    input: {
      title: $title,
      description: $description,
      teamId: $teamId,
      assigneeId: $assigneeId,
      stateId: $stateId
    }
  ) {
    success
    issue {
      id
      title
      description
      assignee {
        id
        name
      }
      state {
        id
        name
      }
    }
  }
}
    `

		variables := map[string]interface{}{
			"title":      title,
			"teamId":     selectedTeamID,
			"assigneeId": assigneeID,
			"stateId":    selectedStateID,
		}
		if description != "" {
			variables["description"] = description
		}

		fmt.Println("Creating issue...")
		createIssueData, err := api.MakeGraphQLRequest(apiKey, mutation, variables)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error making GraphQL request to create issue: %v\n", err)
			os.Exit(1)
		}

		var createResponse IssueCreateResponseData
		if err := json.Unmarshal(createIssueData, &createResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling create issue response data: %v\n", err)
			os.Exit(1)
		}

		if createResponse.IssueCreate.Success {
			fmt.Println("Issue created successfully!")
			fmt.Printf("  ID: %s\n", createResponse.IssueCreate.Issue.ID)
			fmt.Printf("  Title: %s\n", createResponse.IssueCreate.Issue.Title)
			if createResponse.IssueCreate.Issue.Description != "" {
				fmt.Printf("  Description: %s\n", createResponse.IssueCreate.Issue.Description)
			}
			if createResponse.IssueCreate.Issue.Assignee.ID != "" {
				fmt.Printf("  Assignee: %s\n", createResponse.IssueCreate.Issue.Assignee.Name)
			}
			if createResponse.IssueCreate.Issue.State.Name != "" {
				fmt.Printf("  Status: %s\n", createResponse.IssueCreate.Issue.State.Name)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Error creating issue: API reported success: false")
			os.Exit(1)
		}
	},
}

// "issues modify (or edit)" command
var modifyCmd = &cobra.Command{
	Use:   "modify [issue-id]",
	Short: "Modify an existing Linear issue",
	Long:  `Modifies an existing Linear issue identified by its ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		issueID := args[0]
		fmt.Printf("modify issue called for ID: %s\n", issueID)
		fmt.Println("TODO: Implement issue modification logic")
	},
}

func init() {
	rootCmd.AddCommand(
		issuesRootCmd,
	)

	issuesRootCmd.AddCommand(listCmd)
	issuesRootCmd.AddCommand(createCmd)
	issuesRootCmd.AddCommand(modifyCmd)

	// command flags for filtering and limiting
	listCmd.Flags().StringP("team", "t", "", "Filter issues by Team Name")
	listCmd.Flags().
		StringP("state-type", "s", "", "Filter issues by State Type (e.g., 'started', 'completed')")
	listCmd.Flags().IntP("limit", "l", 0, "Limit the number of results")
}
