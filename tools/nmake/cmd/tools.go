package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// toolsCmd represents the tools command
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		env := []string{
			"GOBIN=" + filepath.Join(cwd, "bin"),
		}

		tools := []string{
			"github.com/cheekybits/genny",
			"github.com/goreleaser/goreleaser",
			"github.com/golangci/golangci-lint/cmd/golangci-lint",
			"github.com/vektra/mockery/cmd/mockery",
			"nimona.io/tools/nmake",
			"nimona.io/tools/objectify",
			"nimona.io/tools/generators/community",
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
