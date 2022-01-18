package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "DEV"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version prints the version of this program.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
