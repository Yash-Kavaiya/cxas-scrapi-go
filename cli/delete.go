package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringP("app", "a", "", "App display name or resource name")
	deleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a CXAS app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		appFlag, _ := cmd.Flags().GetString("app")
		force, _ := cmd.Flags().GetBool("force")

		client, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}

		app, err := resolveApp(context.Background(), client, appFlag)
		if err != nil {
			return err
		}

		if !force {
			fmt.Fprintf(os.Stdout, "Delete app %q (%s)? [y/N]: ", app.DisplayName, app.Name)
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Fprintln(os.Stdout, "Cancelled.")
				return nil
			}
		}

		if err := client.DeleteApp(context.Background(), app.Name); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Deleted %s\n", app.Name)
		return nil
	},
}
