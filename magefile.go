//+build mage

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/magefile/mage/sh"
)

// Build builds all main packages
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

	ldflags := `-s -w -X main.Version=v%s -X main.Commit=%s -X main.Date=%s`
	ldflags = fmt.Sprintf(ldflags, newVersion.String(), strings.Trim(commit, `"`), now)
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
		fmt.Println(strings.Join(args, " "))
		fmt.Printf("Building '%s' as %s with version v%s\n", pkg, bn, newVersion.String())
		if err := sh.Run("go", args...); err != nil {
			return err
		}
	}

	return nil
}

// Generate runs go generate
func Generate() error {
	fmt.Println("Running go generate")
	_, err := sh.Exec(nil, os.Stdout, os.Stderr, "go", "generate", "./...")
	return err
}

// Test runs all tests with coverage
func Test() error {
	fmt.Println("Running tests")

	env := map[string]string{
		"LOG_LEVEL":    "debug",
		"DEBUG_BLOCKS": "true",
		"BIND_LOCAL":   "true",
		"UPNP":         "false",
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

	if _, err := sh.Exec(env, os.Stdout, os.Stderr, "go", testArgs...); err != nil {
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

	if _, err := sh.Exec(env, os.Stdout, os.Stderr, "go", coverArgs...); err != nil {
		return err
	}

	return nil
}
