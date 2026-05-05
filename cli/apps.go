package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
)

func init() {
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsGetCmd)
	appsCmd.AddCommand(appsDeleteCmd)
}

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage CXAS apps",
}

var appsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all apps in a project/location",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		client, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}
		list, err := client.ListApps(context.Background())
		if err != nil {
			return err
		}
		return printJSON(list)
	},
}

var appsGetCmd = &cobra.Command{
	Use:   "get <app-name>",
	Short: "Get a single app by resource name or display name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		client, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}
		app, err := client.GetAppByDisplayName(context.Background(), args[0])
		if err != nil {
			return err
		}
		if app == nil {
			// Try as resource name.
			app, err = client.GetApp(context.Background(), args[0])
			if err != nil {
				return err
			}
		}
		return printJSON(app)
	},
}

var appsDeleteCmd = &cobra.Command{
	Use:   "delete <app-name>",
	Short: "Delete an app by resource name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		client, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}
		if err := client.DeleteApp(context.Background(), args[0]); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Deleted app %s\n", args[0])
		return nil
	},
}

// authConfig builds an auth.Config from CLI flags / environment.
func authConfig() auth.Config {
	return auth.Config{}
}

// printJSON marshals v to stdout as indented JSON.
func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
