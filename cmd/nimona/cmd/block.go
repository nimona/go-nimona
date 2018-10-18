package cmd

import (
	"github.com/spf13/cobra"
)

// blockCmd represents the block command
var blockCmd = &cobra.Command{
	Use: "block",
	Aliases: []string{
		"blocks",
	},
	Short: "Block commands",
	Long:  "",
}

func init() {
	rootCmd.AddCommand(blockCmd)
}
