package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
)

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringP("app", "a", "", "App display name or resource name")
	pullCmd.Flags().StringP("out", "d", ".", "Output directory for exported app ZIP")
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Export (pull) an app to a local directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		appFlag, _ := cmd.Flags().GetString("app")
		outDir, _ := cmd.Flags().GetString("out")

		client, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}

		app, err := resolveApp(context.Background(), client, appFlag)
		if err != nil {
			return err
		}

		content, err := client.ExportApp(context.Background(), apps.ExportAppRequest{
			AppName: app.Name,
		})
		if err != nil {
			return err
		}

		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
		outPath := outDir + "/" + safeFileName(app.DisplayName) + ".zip"
		if err := os.WriteFile(outPath, content, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Exported %s to %s\n", app.Name, outPath)
		return nil
	},
}

// resolveApp finds an app by display name or resource name.
func resolveApp(ctx context.Context, client *apps.Client, nameOrDisplay string) (*apps.App, error) {
	app, err := client.GetAppByDisplayName(ctx, nameOrDisplay)
	if err != nil {
		return nil, err
	}
	if app != nil {
		return app, nil
	}
	return client.GetApp(ctx, nameOrDisplay)
}

// safeFileName replaces spaces and slashes with underscores for use in file names.
func safeFileName(s string) string {
	safe := make([]byte, len(s))
	for i, c := range []byte(s) {
		if c == ' ' || c == '/' || c == '\\' {
			safe[i] = '_'
		} else {
			safe[i] = c
		}
	}
	return string(safe)
}
