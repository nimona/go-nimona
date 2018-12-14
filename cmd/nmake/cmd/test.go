package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use: "test",
	Aliases: []string{
		"t",
	},
	Short: "run tests",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running tests")

		env := []string{
			"LOG_LEVEL=debug",
			"DEBUG_BLOCKS=true",
			"BIND_LOCAL=true",
			"UPNP=false",
		}

		testArgs := []string{
			"test",
			"-v",
			"-race",
			"-covermode=atomic",
			"-coverprofile=coverage.out",
			"./...",
		}

		fmt.Println("Cleaning up coverage file for generated files")

		if err := execPipe(env, "go", testArgs, os.Stdout, os.Stderr); err != nil {
			return err
		}

		coverFile, err := os.Open("coverage.out")
		if err != nil {
			return err
		}

		newLines := []string{}
		scanner := bufio.NewScanner(coverFile)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "_generated.go") || strings.Contains(line, "_mock.go") {
				continue
			}
			newLines = append(newLines, line)
		}

		coverFile.Close()

		output := strings.Join(newLines, "\n")
		err = ioutil.WriteFile("coverage.out", []byte(output), 0644)
		if err != nil {
			return err
		}

		coverArgs := []string{
			"tool",
			"cover",
			"-func=coverage.out",
		}

		if err := execPipe(env, "go", coverArgs, os.Stdout, os.Stderr); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
