// Package cli provides the cxas command-line interface.
package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	oauthToken   string
	projectID    string
	location     string
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "cxas",
	Short: "CXAS SCRAPI — CX Agent Studio Developer Toolkit",
	Long:  "cxas is a CLI for managing Google Cloud CX Agent Studio resources.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if oauthToken != "" {
			os.Setenv("CXAS_OAUTH_TOKEN", oauthToken)
		}
		return nil
	},
}

// Execute runs the root command.
func Execute() error { return rootCmd.Execute() }

func init() {
	rootCmd.PersistentFlags().StringVar(&oauthToken, "oauth-token", "", "OAuth2 token (overrides CXAS_OAUTH_TOKEN env var)")
	rootCmd.PersistentFlags().StringVarP(&projectID, "project", "p", "", "GCP project ID")
	rootCmd.PersistentFlags().StringVarP(&location, "location", "l", "us", "CXAS location (default: us)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
}
