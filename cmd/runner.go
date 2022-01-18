package cmd

import (
	"os"

	"github.com/crossjoin-io/crossjoin/runner"
	"github.com/spf13/cobra"
)

// runnerCmd represents the runner command
var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "start a crossjoin runner",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := runner.NewRunner("")
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
}
