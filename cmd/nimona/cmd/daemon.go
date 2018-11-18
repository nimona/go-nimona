package cmd

import (
	"github.com/spf13/cobra"
)

// daemon represents the daemon command
var daemon = &cobra.Command{
	Use: "daemon",
	Aliases: []string{
		"daemons",
	},
	Short: "Daemon commands",
	Long:  "",
}

func init() {
	rootCmd.AddCommand(daemon)
}
