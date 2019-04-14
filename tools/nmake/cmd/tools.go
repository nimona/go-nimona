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
		env := []string{}

		tools := []string{
			"github.com/cheekybits/genny",
			"github.com/golangci/golangci-lint/cmd/golangci-lint",
			"github.com/goreleaser/goreleaser",
			"github.com/golangci/golangci-lint/cmd/golangci-lint",
			"github.com/shurcooL/vfsgen/cmd/vfsgendev",
			"github.com/vektra/mockery/cmd/mockery",
		}

		info("Installing tools")
		for _, tool := range tools {
			extraInfo("* %s", tool)
			if err := execPipe(
				env,
				"go",
				[]string{
					"install",
					tool,
				},
				os.Stdout,
				os.Stderr,
			); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(toolsCmd)
}
