package cmd

import (
	"os"
	"path/filepath"

	"github.com/crossjoin-io/crossjoin/server"
	"github.com/spf13/cobra"
)

var (
	listenAddr   string
	dataDir      string
	configSource string
	config       string
	startRunner  bool
)

// serverCmd represents the runner command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start a crossjoin server",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := server.NewServer(listenAddr, dataDir, configSource, config, startRunner)
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
	serverCmd.Flags().StringVar(&listenAddr, "listen", "127.0.0.1:8000", "listen address")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	serverCmd.Flags().StringVar(&dataDir, "data-dir", filepath.Join(homeDir, ".crossjoin"), "data directory")
	serverCmd.Flags().StringVar(&config, "config", "", "config location to load")
	serverCmd.Flags().StringVar(&configSource, "config-source", "file", "config source type (e.g. `file`, `http`, `github`)")
	serverCmd.Flags().BoolVar(&startRunner, "runner", false, "start a runner")
}
