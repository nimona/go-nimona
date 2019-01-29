package cmd

import (
	"github.com/spf13/cobra"
)

// objectCmd represents the object command
var objectCmd = &cobra.Command{
	Use: "object",
	Aliases: []string{
		"objects",
	},
	Short: "Object commands",
	Long:  "",
}

func init() {
	rootCmd.AddCommand(objectCmd)
}
