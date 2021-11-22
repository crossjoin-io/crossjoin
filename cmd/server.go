package cmd

import (
	"os"
	"path/filepath"

	"github.com/crossjoin-io/crossjoin/server"
	"github.com/spf13/cobra"
)

var (
	listenAddr string
	dataDir    string
)

// serverCmd represents the runner command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start a crossjoin server",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := server.NewServer(listenAddr, dataDir)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		err = s.Start()
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&listenAddr, "listen", ":8000", "listen address")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	serverCmd.Flags().StringVar(&dataDir, "data-dir", filepath.Join(homeDir, ".crossjoin"), "data directory")
}
