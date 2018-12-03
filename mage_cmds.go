//+build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/magefile/mage/sh"
)

// Test code
func Test() error {
	fmt.Println("Running tests")

	env := map[string]string{
		"LOG_LEVEL":    "debug",
		"DEBUG_BLOCKS": "true",
		"BIND_LOCAL":   "true",
		"UPNP":         "false",
	}

	args := []string{
		"test",
		"-v",
		"./...",
	}

	_, err := sh.Exec(env, os.Stdout, os.Stderr, "go", args...)
	return err
}

// Build binaries
func Build() error {
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

	if err := sh.Run("dep", "ensure"); err != nil {
		return err
	}

	pkgs, err := getPackages()
	if err != nil {
		return err
	}

	currentGitTag, err := sh.Output("git", "describe", "--abbrev=0", "--tags")
	if err != nil {
		return err
	}

	version, err := semver.NewVersion(currentGitTag)
	if err != nil {
		return err
	}

	newVersion := version.IncPatch()

	commit, err := sh.Output("git", "show", `--format="%h"`, "--no-patch")
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)

	ldflags := "-s -w -X main.version=v%s -X main.commit=%s -X main.date=%s"
	ldflags = fmt.Sprintf(ldflags, newVersion.String(), commit, now)
	for _, pkg := range pkgs {
		_, bn := filepath.Split(pkg)
		args := []string{
			"build",
			"-tags=release",
			"-a",
			"-o",
			"bin/" + bn,
			"-ldflags",
			ldflags,
		}
		fmt.Printf("Building '%s' as %s with version v%s\n", pkg, bn, newVersion.String())
		if err := sh.Run("go", args...); err != nil {
			return err
		}
	}

	return nil
}
