package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "run cleanups",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		info("Running cleanup")

		paths := []string{
			"./bin/*",
			"./bin",
			"coverage*",
		}

		for _, path := range paths {
			if err := remove(path); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}

func remove(path string) error {
	files, err := filepath.Glob(path)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}
