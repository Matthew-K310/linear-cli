package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "linear",
	Short: "linear-cli is a tool for managing your linear.app projects from the comfort and safety of your terminal.",
	Long: `linear-cli is a command-line tool for managing Linear.app projects, issues, and workflows directly from your terminal. It provides a fast and scriptable way to interact with your workspace.

Key actions include:
* Creating new issues
* Viewing existing issues and their details
* Updating issue status, assignees, and other properties
* Listing projects and teams
* Searching for specific issues

Integrate Linear management into your terminal workflow to enhance productivity and enable automation.`,
	Run: func(cmdn *cobra.Command, arg []string) {
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. All berries! An error while executing linear-cli '%s'\n", err)
		os.Exit(1)
	}
}
