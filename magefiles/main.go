// Package main defines automation targets using Magefile
package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/magefile/mage/sh"
)

// init performs some sanity checks before running anything.
func init() {
	mustBeInRoot()
}

// unformatted error.
var errUnformatted = errors.New("some files were unformatted, make sure `go fmt` is run")

// Checks runs various pre-merge checks.
func Checks() error {
	if err := sh.Run("go", "vet", "./..."); err != nil {
		return fmt.Errorf("failed to run go vet: %w", err)
	}

	out, err := sh.Output("go", "fmt", "./...")
	if err != nil {
		return fmt.Errorf("failed to run gofmt: %w", err)
	}

	if out != "" {
		return errUnformatted
	}

	if err := sh.Run("go", "run", "-mod=readonly", "honnef.co/go/tools/cmd/staticcheck", "./..."); err != nil {
		return fmt.Errorf("failed to run staticcheck: %w", err)
	}

	return nil
}

// Test perform the whole project's unit tests.
func Test() error {
	if err := Dev(); err != nil {
		return fmt.Errorf("failed to setup dev environment: %w", err)
	}

	if err := sh.Run(
		"go", "run", "-mod=readonly", "github.com/onsi/ginkgo/v2/ginkgo",
		"-p", "-randomize-all", "-repeat=5", "--fail-on-pending", "--race", "--trace",
		"--junit-report=test-report.xml", "./...",
	); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}

// Bench run any benchmarks in the project.
func Bench() error {
	if err := sh.Run("go", "test", "-bench=.*", "./..."); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}

// Dev will create or replace containers used for development.
func Dev() error {
	if err := sh.Run("docker", "compose", "-f", "docker-compose.dev.yml", "-p", "clgo-dev", "up",
		"-d", "--build", "--remove-orphans", "--force-recreate"); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}

// error when wrong version format is used.
var errVersionFormat = fmt.Errorf("version must be in format vX,Y,Z")

// Release tags a new version and pushes it.
func Release(version string) error {
	if !regexp.MustCompile(`^v([0-9]+).([0-9]+).([0-9]+)$`).Match([]byte(version)) {
		return errVersionFormat
	}

	if err := sh.Run("git", "tag", version); err != nil {
		return fmt.Errorf("failed to tag version: %w", err)
	}

	if err := sh.Run("git", "push", "origin", version); err != nil {
		return fmt.Errorf("failed to push version tag: %w", err)
	}

	return nil
}

// mustBeInRoot checks that the command is run in the project root.
func mustBeInRoot() {
	if _, err := os.Stat("go.mod"); err != nil {
		panic("must be in root, couldn't stat go.mod file: " + err.Error())
	}
}
