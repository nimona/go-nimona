package cmd

import (
	"fmt"

	"github.com/shurcooL/vfsgen"
	"github.com/spf13/cobra"

	"nimona.io/internal/api"
)

// vfsgenCmd represents the vfsgen command
var vfsgenCmd = &cobra.Command{
	Use:   "vfsgen",
	Short: "vfsgen",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running vfsgen")

		err := vfsgen.Generate(api.Assets, vfsgen.Options{
			PackageName:  "api",
			BuildTags:    "!dev",
			VariableName: "Assets",
		})
		return err
	},
}

func init() {
	rootCmd.AddCommand(vfsgenCmd)
}
