package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
)


// installCmd represents the install command
var installCmd = &cobra.Command{
	Use: "install",
	Aliases: []string{
		"i",
	},
	Short: "install",
	RunE: func(cmd *cobra.Command, args []string) error {
		env := []string{}

		info("Trying to figure out current version from git tags")
		currentGitTag, err := exec(env, "git", []string{"describe", "--abbrev=0", "--tags"})
		if err != nil {
			return err
		}

		tagged := false
		if _, err := exec(env, "git", []string{"describe", "--exact-match"}); err == nil {
			tagged = true
		}

		currentGitTag = strings.TrimSpace(currentGitTag)
		version, err := semver.NewVersion(currentGitTag)
		if err != nil {
			return err
		}

		extraInfo("* current version v%s", version.String())

		newSemVersion := version.IncPatch()
		newVersion := "v" + newSemVersion.String()

		commit, err := exec(env, "git", []string{"show", `--format="%h"`, "--no-patch"})
		if err != nil {
			return err
		}
		commit = strings.TrimSpace(commit)
		commit = strings.Trim(commit, `'"`)

		extraInfo("* current commit %s", commit)

		if !tagged {
			newVersion += "-dev+" + commit
		}
		extraInfo("* new version %s", newVersion)

		info("Building packages")
		ldflags := `-s -w -X main.Version=%s -X main.Commit=%s -X main.Date=%s`
		ldflags = fmt.Sprintf(ldflags, newVersion, strings.Trim(commit, `"`), time.Now().UTC().Format(time.RFC3339))
			iargs := []string{
				"install",
				"-tags=release",
				"-a",
				"-ldflags",
				ldflags,
				"./cmd/nimona",
			}
			if err := execPipe(env, "go", iargs, os.Stdout, os.Stderr); err != nil {
				return err
			}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
