package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/magefile/mage/sh"
)

// init performs some sanity checks before running anything
func init() {
	mustBeInRoot()
}

// Test perform the whole project's unit tests
func Test() error {
	if err := Dev(); err != nil {
		return fmt.Errorf("failed to setup dev environment: %w", err)
	}

	return sh.Run(
		"go", "run", "-mod=readonly", "github.com/onsi/ginkgo/v2/ginkgo",
		"-p", "-randomize-all", "-repeat=5", "--fail-on-pending", "--race", "--trace",
		"--junit-report=test-report.xml", "./...",
	)
}

// Bench run any benchmarks in the project
func Bench() error {
	return sh.Run("go", "test", "-bench='.*'", "-test.run=notests", "./...")
}

// Dev will create or replace containers used for development
func Dev() error {
	return sh.Run("docker", "compose", "-f", "docker-compose.dev.yml", "-p", "clgo-dev", "up",
		"-d", "--build", "--remove-orphans", "--force-recreate")
}

// Release tags a new version and pushes it
func Release(version string) error {
	if !regexp.MustCompile(`^v([0-9]+).([0-9]+).([0-9]+)$`).Match([]byte(version)) {
		return fmt.Errorf("version must be in format vX,Y,Z")
	}

	if err := sh.Run("git", "tag", version); err != nil {
		return fmt.Errorf("failed to tag version: %w", err)
	}
	if err := sh.Run("git", "push", "origin", version); err != nil {
		return fmt.Errorf("failed to push version tag: %w", err)
	}

	return nil
}

// mustBeInRoot checks that the command is run in the project root
func mustBeInRoot() {
	if _, err := os.Stat("go.mod"); err != nil {
		panic("must be in root, couldn't stat go.mod file: " + err.Error())
	}
}
