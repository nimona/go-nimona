package cmd

import (
	"github.com/spf13/cobra"
)

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a daemon",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := restClient.R().Post("/api/v1/stop")
		return err
	},
}

func init() {
	daemon.AddCommand(daemonStopCmd)
}
