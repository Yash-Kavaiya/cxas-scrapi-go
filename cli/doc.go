// Package cli provides the cxas command-line interface.
package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "cxas",
	Short: "CXAS SCRAPI — CX Agent Studio Developer Toolkit",
}

// Execute runs the root cobra command.
func Execute() error {
	return rootCmd.Execute()
}
