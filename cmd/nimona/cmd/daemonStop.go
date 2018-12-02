package cmd

import (
	"syscall"

	"github.com/spf13/cobra"
)

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a daemon",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {

		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

		return nil
	},
}

func init() {
	daemon.AddCommand()
}
