package cmd

import (
	"github.com/spf13/cobra"
)

// identityCmd represents the identity command
var identityCmd = &cobra.Command{
	Use: "identity",
	Aliases: []string{
		"id",
	},
	Short: "Identity commands",
	Long:  "",
}

func init() {
	rootCmd.AddCommand(identityCmd)
}
