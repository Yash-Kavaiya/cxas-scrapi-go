package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
)

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringP("app", "a", "", "App display name or resource name (omit to create new)")
	pushCmd.Flags().StringP("file", "f", "", "Path to app ZIP or directory")
	pushCmd.Flags().String("display-name", "", "Display name for new app (required when --app is omitted)")
	pushCmd.Flags().String("conflict", "OVERWRITE", "Conflict strategy: OVERWRITE or SKIP")
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Import (push) an app from a local file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		appFlag, _ := cmd.Flags().GetString("app")
		filePath, _ := cmd.Flags().GetString("file")
		displayName, _ := cmd.Flags().GetString("display-name")
		conflict, _ := cmd.Flags().GetString("conflict")

		if filePath == "" {
			return fmt.Errorf("--file is required")
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		client, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}

		if appFlag != "" {
			app, err := resolveApp(context.Background(), client, appFlag)
			if err != nil {
				return err
			}
			updated, err := client.ImportApp(context.Background(), apps.ImportAppRequest{
				AppName:          app.Name,
				AppContent:       content,
				ConflictStrategy: conflict,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Imported into %s\n", updated.Name)
		} else {
			if displayName == "" {
				return fmt.Errorf("--display-name is required when creating a new app")
			}
			app, err := client.CreateApp(context.Background(), apps.CreateAppRequest{
				DisplayName: displayName,
			})
			if err != nil {
				return err
			}
			_, err = client.ImportApp(context.Background(), apps.ImportAppRequest{
				AppName:          app.Name,
				AppContent:       content,
				ConflictStrategy: conflict,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Created and imported app %s\n", app.Name)
		}
		return nil
	},
}
