package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/Matthew-K310/linear-cli/internal/api"
	"github.com/Matthew-K310/linear-cli/internal/config"
	"github.com/Matthew-K310/linear-cli/internal/linear"
)

// Helper prompt functions

func promptForString(label, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:   fmt.Sprintf("%s (leave empty to keep current)", label),
		Default: defaultValue,
	}
	return prompt.Run()
}

func promptForSelect(label string, items []string, defaultIndex int) (int, error) {
	prompt := promptui.Select{
		Label:     label,
		Items:     items,
		CursorPos: defaultIndex,
		Searcher: func(input string, index int) bool {
			item := items[index]
			return strings.Contains(strings.ToLower(item), strings.ToLower(input))
		},
	}

	index, _, err := prompt.Run()
	return index, err
}

var modifyCmd = &cobra.Command{
	Use:   "modify [issue-id]",
	Short: "Modify an existing Linear issue",
	Long:  `Modifies an existing Linear issue identified by its ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := config.GetAPIKey()
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: API_KEY is not set.")
			os.Exit(1)
		}

		// Fetch teams
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
		teamsData, err := api.MakeGraphQLRequest(apiKey, teamsQuery, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching teams: %v\n", err)
			os.Exit(1)
		}

		var teamsResponse linear.TeamsResponseData
		if err := json.Unmarshal(teamsData, &teamsResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling teams data: %v\n", err)
			os.Exit(1)
		}

		if len(teamsResponse.Teams.Nodes) == 0 {
			fmt.Fprintln(os.Stderr, "No teams found. Cannot list issues.")
			os.Exit(1)
		}

		teamNames := []string{}
		teamMap := make(map[string]string)
		for _, team := range teamsResponse.Teams.Nodes {
			teamNames = append(teamNames, team.Name)
			teamMap[team.Name] = team.ID
		}

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
			fmt.Fprintf(os.Stderr, "Team selection failed: %v\n", err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			os.Exit(1)
		}
		selectedTeamID := teamMap[selectedTeamName]

		stateType, _ := cmd.Flags().GetString("state-type")
		limit, _ := cmd.Flags().GetInt("limit")

		issueQuery := `
		query Issue($teamId: ID, $stateType: String, $first: Int) {
			issues(filter: {team: {id: {eq: $teamId}}, state: {type: {eq: $stateType}}}, first: $first) {
				nodes {
					id
					identifier
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
					project {
						id
						name
					}
					team {
						id
						name
					}
				}
			}
		}
		`

		variables := map[string]any{
			"teamId": selectedTeamID,
		}
		if stateType != "" {
			variables["stateType"] = stateType
		}
		if limit > 0 {
			variables["first"] = limit
		} else {
			variables["first"] = 50
		}

		data, err := api.MakeGraphQLRequest(apiKey, issueQuery, variables)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error making GraphQL request: %v\n", err)
			os.Exit(1)
		}

		var issuesResponse linear.IssuesResponseData
		if err := json.Unmarshal(data, &issuesResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling issues data: %v\n", err)
			os.Exit(1)
		}

		var issueDisplayItems []string
		var selectableIssues []linear.IssueNode
		for _, issue := range issuesResponse.Issues.Nodes {
			if issue.State.Name == "Done" || issue.State.Name == "Canceled" {
				continue
			}
			display := fmt.Sprintf(
				"%s: %s | Status: %s",
				issue.Identifier,
				issue.Title,
				issue.State.Name,
			)
			issueDisplayItems = append(issueDisplayItems, display)
			selectableIssues = append(selectableIssues, issue)
		}

		if len(issueDisplayItems) == 0 {
			fmt.Println("No selectable issues found matching the criteria in the selected team.")
			os.Exit(0)
		}

		issueSelectPrompt := promptui.Select{
			Label: "Select Issue to Modify",
			Items: issueDisplayItems,
			Searcher: func(input string, index int) bool {
				item := issueDisplayItems[index]
				return strings.Contains(strings.ToLower(item), strings.ToLower(input))
			},
		}

		selectedIndex, _, err := issueSelectPrompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Issue selection failed: %v\n", err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			os.Exit(1)
		}

		selectedIssue := selectableIssues[selectedIndex]

		// Fetch detailed info for selected issue
		selectedIssueQuery := `
		query IssueByID($id: String!) {
			issue(id: $id) {
				id
				identifier
				title
				description
				project {
					id
					name
				}
				assignee {
					id
					name
				}
				state {
					id
					name
					type
				}
			}
		}
		`

		issueVariables := map[string]any{
			"id": selectedIssue.ID,
		}

		data, err = api.MakeGraphQLRequest(apiKey, selectedIssueQuery, issueVariables)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching issue details: %v\n", err)
			os.Exit(1)
		}

		var issueDetailResponse struct {
			Issue linear.IssueNode `json:"issue"`
		}

		if err := json.Unmarshal(data, &issueDetailResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling issue details: %v\n", err)
			os.Exit(1)
		}

		detailedIssue := issueDetailResponse.Issue

		projectName := ""
		if detailedIssue.Project != nil {
			projectName = detailedIssue.Project.Name
		}
		assigneeName := ""
		if detailedIssue.Assignee != nil {
			assigneeName = detailedIssue.Assignee.Name
		}

		fmt.Println("--------------------")
		fmt.Printf(
			"Current Issue Details:\n ID: %s\n Title: %s\n Description: %s\n Project: %s\n Assignee: %s\n Status: %s\n",
			detailedIssue.Identifier,
			detailedIssue.Title,
			detailedIssue.Description,
			projectName,
			assigneeName,
			detailedIssue.State.Name,
		)
		fmt.Println("--------------------")

		// Fetch projects for team to select new project
		projectsQuery := `
		query TeamProjects($teamId: String!) {
			team(id: $teamId) {
				projects {
					nodes {
						id
						name
					}
				}
			}
		}
		`
		projectVars := map[string]any{"teamId": selectedTeamID}
		projectData, err := api.MakeGraphQLRequest(apiKey, projectsQuery, projectVars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching projects: %v\n", err)
			os.Exit(1)
		}

		var projectsResponse struct {
			Team struct {
				Projects struct {
					Nodes []linear.ProjectNode `json:"nodes"`
				} `json:"projects"`
			} `json:"team"`
		}
		if err := json.Unmarshal(projectData, &projectsResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling projects data: %v\n", err)
			os.Exit(1)
		}

		// Fetch users (assignees) for team
		usersQuery := `
		query TeamUsers($teamId: String!) {
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
		userData, err := api.MakeGraphQLRequest(apiKey, usersQuery, projectVars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching users: %v\n", err)
			os.Exit(1)
		}

		var usersResponse struct {
			Team struct {
				Members struct {
					Nodes []linear.UserNode `json:"nodes"`
				} `json:"members"`
			} `json:"team"`
		}
		if err := json.Unmarshal(userData, &usersResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling users data: %v\n", err)
			os.Exit(1)
		}

		// Fetch states for team
		statesQuery := `
		query TeamStates($teamId: String!) {
			team(id: $teamId) {
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
		stateData, err := api.MakeGraphQLRequest(apiKey, statesQuery, projectVars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching states: %v\n", err)
			os.Exit(1)
		}

		var statesResponse struct {
			Team struct {
				States struct {
					Nodes []linear.StateNode `json:"nodes"`
				} `json:"states"`
			} `json:"team"`
		}
		if err := json.Unmarshal(stateData, &statesResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling states data: %v\n", err)
			os.Exit(1)
		}

		// Prompt user for new values, allowing to keep existing values

		newTitle, err := promptForString("Title", detailedIssue.Title)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed: %v\n", err)
			os.Exit(1)
		}

		newDescription, err := promptForString("Description", detailedIssue.Description)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed: %v\n", err)
			os.Exit(1)
		}

		// Project selection
		projectNames := make([]string, len(projectsResponse.Team.Projects.Nodes))
		projectDefault := 0
		for i, p := range projectsResponse.Team.Projects.Nodes {
			projectNames[i] = p.Name
			if detailedIssue.Project != nil && p.ID == detailedIssue.Project.ID {
				projectDefault = i
			}
		}
		selectedProjectIndex, err := promptForSelect("Select Project", projectNames, projectDefault)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Project selection failed: %v\n", err)
			os.Exit(1)
		}
		newProjectID := projectsResponse.Team.Projects.Nodes[selectedProjectIndex].ID

		// Assignee selection (include Unassigned option)
		assigneeNames := make([]string, len(usersResponse.Team.Members.Nodes)+1)
		assigneeNames[0] = "Unassigned"
		assigneeDefault := 0
		for i, u := range usersResponse.Team.Members.Nodes {
			assigneeNames[i+1] = u.Name
			if detailedIssue.Assignee != nil && u.ID == detailedIssue.Assignee.ID {
				assigneeDefault = i + 1
			}
		}
		selectedAssigneeIndex, err := promptForSelect(
			"Select Assignee",
			assigneeNames,
			assigneeDefault,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Assignee selection failed: %v\n", err)
			os.Exit(1)
		}

		var newAssigneeID *string
		if selectedAssigneeIndex == 0 {
			// Unassigned
			newAssigneeID = nil
		} else {
			id := usersResponse.Team.Members.Nodes[selectedAssigneeIndex-1].ID
			newAssigneeID = &id
		}

		// State selection
		stateNames := make([]string, len(statesResponse.Team.States.Nodes))
		stateDefault := 0
		for i, s := range statesResponse.Team.States.Nodes {
			stateNames[i] = s.Name
			if detailedIssue.State.ID == s.ID {
				stateDefault = i
			}
		}
		selectedStateIndex, err := promptForSelect("Select Status", stateNames, stateDefault)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Status selection failed: %v\n", err)
			os.Exit(1)
		}
		newStateID := statesResponse.Team.States.Nodes[selectedStateIndex].ID

		// Prepare mutation
		updateMutation := `
		mutation IssueUpdate(
			$id: String!,
			$title: String,
			$description: String,
			$projectId: String,
			$assigneeId: String,
			$stateId: String
		) {
			issueUpdate(
				id: $id,
				input: {
					title: $title,
					description: $description,
					projectId: $projectId,
					assigneeId: $assigneeId,
					stateId: $stateId
				}
			) {
				success
				issue {
					identifier
					title
					description
					project {
						name
					}
					assignee {
						name
					}
					state {
						name
					}
				}
			}
		}
		`

		mutationVariables := map[string]any{
			"id":          detailedIssue.ID,
			"title":       newTitle,
			"description": newDescription,
			"projectId":   newProjectID,
			"assigneeId":  newAssigneeID,
			"stateId":     newStateID,
		}

		mutationData, err := api.MakeGraphQLRequest(apiKey, updateMutation, mutationVariables)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating issue: %v\n", err)
			os.Exit(1)
		}

		var mutationResponse struct {
			IssueUpdate struct {
				Success bool             `json:"success"`
				Issue   linear.IssueNode `json:"issue"`
			} `json:"issueUpdate"`
		}

		if err := json.Unmarshal(mutationData, &mutationResponse); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling mutation response: %v\n", err)
			os.Exit(1)
		}

		if mutationResponse.IssueUpdate.Success {
			fmt.Println("Issue updated successfully!")
			fmt.Printf("Updated Issue:\n Identifier: %s\n Title: %s\n Status: %s\n",
				mutationResponse.IssueUpdate.Issue.Identifier,
				mutationResponse.IssueUpdate.Issue.Title,
				mutationResponse.IssueUpdate.Issue.State.Name,
			)
		} else {
			fmt.Println("Issue update failed.")
			os.Exit(1)
		}
	},
}

func init() {
	modifyCmd.Flags().
		StringP("state-type", "s", "", "Filter issues by state type (e.g., 'backlog', 'unstarted', 'started', 'completed', 'canceled')")
	modifyCmd.Flags().IntP("limit", "l", 50, "Limit the number of issues fetched")
}
