package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/evals/turn"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/sessions"
)

func init() {
	rootCmd.AddCommand(evalCmd)
	evalCmd.AddCommand(evalRunCmd)
	evalRunCmd.Flags().StringP("app", "a", "", "App display name or resource name")
	evalRunCmd.Flags().StringP("file", "f", "", "Path to turn-eval YAML file")
}

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Run evaluations against a CXAS app",
}

var evalRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run turn-based evaluations from a YAML file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		appFlag, _ := cmd.Flags().GetString("app")
		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file is required")
		}

		// Load eval file.
		f, err := turn.LoadFile(filePath)
		if err != nil {
			return err
		}
		turn.ApplyGlobalConfig(f)

		// Resolve app.
		appsClient, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}
		app, err := resolveApp(context.Background(), appsClient, appFlag)
		if err != nil {
			return err
		}

		// Create sessions client.
		sessClient, err := sessions.NewClient(context.Background(), authConfig())
		if err != nil {
			return err
		}

		runner := func(ctx context.Context, appName, sessID, text string) (*sessions.SessionOutput, error) {
			return sessClient.Run(ctx, sessions.RunSessionRequest{
				AppName:   appName,
				SessionID: sessID,
				Input:     sessions.SessionInput{Text: text},
			})
		}

		result, err := turn.RunTests(context.Background(), f.AllTests(), app.Name, runner)
		if err != nil {
			return err
		}

		for _, r := range result.Results {
			status := "PASS"
			if !r.Passed {
				status = "FAIL"
			}
			fmt.Fprintf(os.Stdout, "[%s] %s turn %d: %s\n", status, r.TestID, r.TurnIndex, r.Expectation)
			if !r.Passed && r.Reason != "" {
				fmt.Fprintf(os.Stdout, "       reason: %s\n", r.Reason)
			}
		}

		fmt.Fprintf(os.Stdout, "\nTotal: %d  Passed: %d  Failed: %d\n",
			result.Total, result.Passed, result.Failed)

		if result.Failed > 0 {
			return fmt.Errorf("%d evaluation(s) failed", result.Failed)
		}
		return nil
	},
}
