package cmd

import (
	"github.com/spf13/cobra"
)

var issuesRootCmd = &cobra.Command{
	Use:   "issues",
	Short: "Manage Linear issues",
	Long:  `Provides commands to create, list, and modify Linear issues.`,
}

func init() {
	rootCmd.AddCommand(
		issuesRootCmd,
	)

	issuesRootCmd.AddCommand(listCmd)
	issuesRootCmd.AddCommand(createCmd)
	issuesRootCmd.AddCommand(modifyCmd)

	// command flags for filtering and limiting
	// listCmd.Flags().StringP("team", "t", "", "Filter issues by Team Name")
	// listCmd.Flags().
	// 	StringP("state-type", "s", "", "Filter issues by State Type (e.g., 'started', 'completed')")
	// listCmd.Flags().IntP("limit", "l", 0, "Limit the number of results")
}

var IssuesRootCmd = issuesRootCmd
