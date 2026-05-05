package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/sessions"
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("app", "a", "", "App display name or resource name")
	runCmd.Flags().StringP("session", "s", "default", "Session ID")
	runCmd.Flags().StringP("text", "t", "", "Text input for the session")
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a single session turn against a CXAS app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--project is required")
		}
		appFlag, _ := cmd.Flags().GetString("app")
		sessID, _ := cmd.Flags().GetString("session")
		text, _ := cmd.Flags().GetString("text")

		if text == "" && len(args) > 0 {
			text = args[0]
		}
		if text == "" {
			return fmt.Errorf("--text or positional argument required")
		}

		appsClient, err := apps.NewClient(context.Background(), projectID, location, authConfig())
		if err != nil {
			return err
		}
		app, err := resolveApp(context.Background(), appsClient, appFlag)
		if err != nil {
			return err
		}

		sessClient, err := sessions.NewClient(context.Background(), authConfig())
		if err != nil {
			return err
		}

		out, err := sessClient.Run(context.Background(), sessions.RunSessionRequest{
			AppName:   app.Name,
			SessionID: sessID,
			Input:     sessions.SessionInput{Text: text},
		})
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, out.Text)
		if len(out.ToolCalls) > 0 {
			fmt.Fprintf(os.Stdout, "[tool calls: ")
			for _, tc := range out.ToolCalls {
				fmt.Fprintf(os.Stdout, "%s ", tc.ToolName)
			}
			fmt.Fprintln(os.Stdout, "]")
		}
		if out.AgentTransfer != "" {
			fmt.Fprintf(os.Stdout, "[transferred to: %s]\n", out.AgentTransfer)
		}
		return nil
	},
}
