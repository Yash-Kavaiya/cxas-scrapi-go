package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
)

func init() {
	rootCmd.AddCommand(branchCmd)
	branchCmd.Flags().StringP("app", "a", "", "Source app display name or resource name")
	branchCmd.Flags().StringP("name", "n", "", "Display name for the new branch app")
}

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Create a copy (branch) of an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		appFlag, _ := cmd.Flags().GetString("app")
		newName, _ := cmd.Flags().GetString("name")
		if newName == "" {
			return fmt.Errorf("--name is required")
		}

		client, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}

		// Export the source app.
		src, err := resolveApp(context.Background(), client, appFlag)
		if err != nil {
			return err
		}
		content, err := client.ExportApp(context.Background(), apps.ExportAppRequest{AppName: src.Name})
		if err != nil {
			return err
		}

		// Create the new app.
		newApp, err := client.CreateApp(context.Background(), apps.CreateAppRequest{DisplayName: newName})
		if err != nil {
			return err
		}

		// Import the exported content.
		_, err = client.ImportApp(context.Background(), apps.ImportAppRequest{
			AppName:          newApp.Name,
			AppContent:       content,
			ConflictStrategy: "OVERWRITE",
		})
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Branched %s → %s\n", src.Name, newApp.Name)
		return nil
	},
}
