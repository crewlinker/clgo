package main

import (
	"fmt"
	"os"

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

// Dev will create or replace containers used for development
func Dev() error {
	return sh.Run("docker", "compose", "-f", "docker-compose.dev.yml", "-p", "clgo-dev", "up",
		"-d", "--build", "--remove-orphans", "--force-recreate")
}

// mustBeInRoot checks that the command is run in the project root
func mustBeInRoot() {
	if _, err := os.Stat("go.mod"); err != nil {
		panic("must be in root, couldn't stat go.mod file: " + err.Error())
	}
}
