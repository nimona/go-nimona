package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// toolsCmd represents the tools command
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		env := []string{
			"GO111MODULE=off",
		}

		tools := []string{
			"github.com/goreleaser/goreleaser",
			"github.com/golangci/golangci-lint/cmd/golangci-lint",
		}

		info("Installing tools")
		for _, tool := range tools {
			extraInfo("* %s", tool)
			if err := execPipe(env, "go", []string{"get", tool}, os.Stdout, os.Stderr); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(toolsCmd)
}
