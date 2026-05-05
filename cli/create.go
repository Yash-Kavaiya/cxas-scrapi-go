package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
)

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("name", "n", "", "Display name for the new app")
	createCmd.Flags().StringP("description", "d", "", "Description for the new app")
	createCmd.Flags().String("app-id", "", "Custom app ID (optional)")
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new CXAS app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		name, _ := cmd.Flags().GetString("name")
		desc, _ := cmd.Flags().GetString("description")
		appID, _ := cmd.Flags().GetString("app-id")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		client, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}

		app, err := client.CreateApp(context.Background(), apps.CreateAppRequest{
			DisplayName: name,
			Description: desc,
			AppID:       appID,
		})
		if err != nil {
			return err
		}
		return printJSON(app)
	},
}
