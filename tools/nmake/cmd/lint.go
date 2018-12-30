package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use: "lint",
	Aliases: []string{
		"l",
	},
	Short: "run lints",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running linters")

		env := []string{}

		lintArgs := []string{
			"run",
		}

		if err := execPipe(env, "golangci-lint", lintArgs, os.Stdout, os.Stderr); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
