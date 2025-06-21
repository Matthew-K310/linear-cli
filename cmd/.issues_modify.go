package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Matthew-K310/linear-cli/internal/api"
	"github.com/Matthew-K310/linear-cli/internal/config"
	"github.com/Matthew-K310/linear-cli/internal/linear"
)

// query teams
// get projects
// get issue titles and status
// have prompt to select issue from list of titles and IDs (maybe status too)
// figure out a method to select fields to modify

// modifyCmd represents the modify command
var modifyCmd = &cobra.Command{
	Use:   "modify [issue-id]",
	Short: "Modify an existing Linear issue",
	Long:  `Modifies an existing Linear issue identified by its ID.`,
	Args:  cobra.ExactArgs(1), // Requires exactly one argument (the issue ID)
	Run: func(cmd *cobra.Command, args []string) {
		issueID := args[0]
		fmt.Printf("Modify issue called for ID: %s\n", issueID)
		fmt.Println("TODO: Implement issue modification logic")
		// Exit with a non-zero status code to indicate unimplemented state
		os.Exit(1)
		teamNameFromFlag, _ := cmd.Flags().GetString("team")
		projectNameFromFlag, _ := cmd.Flags().GetString("project")
		stateType, _ := cmd.Flags().GetString("state-type")
		limit, _ := cmd.Flags().GetInt("limit")

		apiKey := config.GetAPIKey()
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: API_KEY is not set.")
			os.Exit(1)
		}

		teamID := ""

		// get team name
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
			teamVars := map[string]any{
				"teamName": teamNameFromFlag,
			}

			fmt.Printf("Looking up Team ID for name '%s'...\n", teamNameFromFlag)
			teamData, err := api.MakeGraphQLRequest(apiKey, teamQuery, teamVars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error looking up team '%s': %v\n", teamNameFromFlag, err)
				os.Exit(1)
			}

			var teamsResponse linear.TeamsResponseData
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

		projectNameFromFlag, _ = cmd.Flags().
			GetString("project")
			// Get value from new --project flag

		projectID := "" // Variable to store the found Project ID

		// If a project name is provided, find its ID within the selected team
		if projectNameFromFlag != "" {
			// Require --team flag if --project is used
			if teamID == "" {
				fmt.Fprintln(
					os.Stderr,
					"Error: --project flag requires the --team flag to be specified first.",
				)
				os.Exit(1)
			}

			// GraphQL query to find a project by name within a specific team
			// This query is similar to the one used in the createCmd for fetching projects for a team,
			// but with an added filter by project name.
			projectLookupQuery := `
			query TeamProjectsFiltered($teamId: String!, $projectName: String!) {
				team(id: $teamId) {
					projects(filter: {name: {eq: $projectName}}) {
						nodes {
							id
							name
						}
					}
				}
			}
			`

			projectVars := map[string]any{
				"teamId":      teamID, // Use the ID of the already found/specified team
				"projectName": projectNameFromFlag,
			}

			fmt.Printf(
				"Looking up Project ID for name '%s' within team (ID: %s)...\n",
				projectNameFromFlag,
				teamID,
			)
			projectData, err := api.MakeGraphQLRequest(apiKey, projectLookupQuery, projectVars)
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"Error looking up project '%s' in team (ID: %s): %v\n",
					projectNameFromFlag,
					teamID,
					err,
				)
				os.Exit(1)
			}

			var teamProjectsResponse linear.TeamProjectsResponseData // Use the struct for projects within a team
			if err := json.Unmarshal(projectData, &teamProjectsResponse); err != nil {
				fmt.Fprintf(os.Stderr, "Error unmarshalling project lookup data: %v\n", err)
				os.Exit(1)
			}

			// Check the response: Ensure the team was found and has the projects field populated,
			// and check the number of matching projects.
			if teamProjectsResponse.Team.Projects.Nodes == nil ||
				len(teamProjectsResponse.Team.Projects.Nodes) == 0 {
				fmt.Fprintf(
					os.Stderr,
					"Error: Project '%s' not found in team (ID: %s).\n",
					projectNameFromFlag,
					teamID,
				)
				os.Exit(1)
			} else if len(teamProjectsResponse.Team.Projects.Nodes) > 1 {
				// Linear project names are unique within a team, so this case should ideally not happen
				// unless there's an unexpected API response or search behavior.
				fmt.Fprintf(os.Stderr, "Warning: Multiple projects found with name '%s' in team (ID: %s). Using the first one found.\n", projectNameFromFlag, teamID)
				projectID = teamProjectsResponse.Team.Projects.Nodes[0].ID
			} else {
				// Exactly one project found
				projectID = teamProjectsResponse.Team.Projects.Nodes[0].ID
				fmt.Printf("Found Project ID: %s\n", projectID)
			}
		}

		query := `
		`

		variables := map[string]any{}
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
		// Add projectID to variables ONLY if it was found
		if projectID != "" {
			variables["projectId"] = projectID
		}
	},
}

func init() {
	// Define flags for the list command
	listCmd.Flags().StringP("team", "t", "", "Filter issues by Team Name")
	listCmd.Flags().StringP("project", "p", "", "Filter issues by Project")
	listCmd.Flags().
		StringP("state-type", "s", "", "Filter issues by State Type (e.g., 'started', 'completed')")
	listCmd.Flags().IntP("limit", "l", 0, "Limit the number of results")
}

// init function is not strictly necessary if you register commands in issues.go's init
/*
func init() {
	// No specific flags defined for the modify command in the original code
}
*/
