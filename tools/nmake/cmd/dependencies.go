package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// depsCmd represents the dependencies command
var depsCmd = &cobra.Command{
	Use: "dependencies",
	Aliases: []string{
		"deps",
	},
	Short: "deps",
	RunE: func(cmd *cobra.Command, args []string) error {
		env := []string{}

		info("Running `go mod vendor`")
		if err := execPipe(env, "go", []string{"mod", "vendor"}, os.Stdout, os.Stderr); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(depsCmd)
}
