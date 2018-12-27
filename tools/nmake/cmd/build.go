package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
)

// buildCmd represents the daemon command
var buildCmd = &cobra.Command{
	Use: "build",
	Aliases: []string{
		"b",
	},
	Short: "build",
	RunE: func(cmd *cobra.Command, args []string) error {
		getPackages := func() ([]string, error) {
			pkgs, err := filepath.Glob("./*/*/main.go")
			if err != nil {
				return nil, err
			}

			for i := range pkgs {
				pkgs[i] = strings.Replace(pkgs[i], "/main.go", "", -1)
			}

			return pkgs, nil
		}

		env := []string{}

		info("Trying to find packages to build")
		pkgs, err := getPackages()
		if err != nil {
			return err
		}

		if len(pkgs) == 0 {
			return errors.New("couldn't find any packages to build")
		}

		for _, pkg := range pkgs {
			extraInfo("* %s", pkg)
		}

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
		for _, pkg := range pkgs {
			_, bn := filepath.Split(pkg)
			args := []string{
				"build",
				"-tags=release",
				"-a",
				"-ldflags",
				ldflags,
				"-o",
				"bin/" + bn,
				"./" + pkg,
			}
			extraInfo("* building '%s' as '%s'", pkg, "bin/"+bn)
			if err := execPipe(env, "go", args, os.Stdout, os.Stderr); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
