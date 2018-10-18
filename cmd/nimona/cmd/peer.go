package cmd

import (
	"github.com/spf13/cobra"
)

// peerCmd represents the peer command
var peerCmd = &cobra.Command{
	Use: "peer",
	Aliases: []string{
		"peers",
	},
	Short: "Peer commands",
	Long:  "",
}

func init() {
	rootCmd.AddCommand(peerCmd)
}
