package cmd

import (
	"os"

	"github.com/crossjoin-io/crossjoin/runner"
	"github.com/spf13/cobra"
)

var (
	apiURL string
)

// runnerCmd represents the runner command
var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "start a crossjoin runner",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := runner.NewRunner(apiURL)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		err = r.Start()
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runnerCmd)
	runnerCmd.Flags().StringVar(&apiURL, "api-url", "http://127.0.0.1:8000", "URL to API server instance")
}
