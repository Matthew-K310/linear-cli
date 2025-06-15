package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "linear-cli",
	Short: "A command line interface for interacting with the Linear API",
	Long:  `A simple CLI tool to fetch and manage Linear data via its GraphQL API.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
}
