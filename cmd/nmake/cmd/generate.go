package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use: "generate",
	Aliases: []string{
		"g",
	},
	Short: "generate",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running go generate")

		env := []string{}

		genArgs := []string{
			"generate",
			"./...",
		}

		return execPipe(env, "go", genArgs, os.Stdout, os.Stderr)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
