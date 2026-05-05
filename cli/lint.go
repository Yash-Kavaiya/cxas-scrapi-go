package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/linter"
)

func init() {
	rootCmd.AddCommand(lintCmd)
	lintCmd.Flags().StringP("dir", "d", ".", "Directory to lint")
}

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint a CXAS app directory for common issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")

		l := linter.New(dir, nil)
		report, err := l.LintDirectory(dir)
		if err != nil {
			return err
		}

		for _, r := range report.Results {
			fmt.Fprintln(os.Stdout, r.String())
		}
		fmt.Fprintf(os.Stdout, "\n%d files, %d errors, %d warnings, %d info\n",
			report.Files, report.Errors, report.Warnings, report.Infos)

		if report.HasErrors() {
			return fmt.Errorf("lint failed with %d error(s)", report.Errors)
		}
		return nil
	},
}
