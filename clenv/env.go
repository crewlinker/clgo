// Package clenv provides functionality to work with environment variables.
package clenv

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// LoadFromGitRoot loads environment variables from a file in the root of
// the git repository.
func LoadFromGitRoot(ctx context.Context, names ...string) error {
	var errb, outb bytes.Buffer

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmd.Stderr = &errb
	cmd.Stdout = &outb

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run git rev-parse --show-toplevel: %w: %v", err, errb.String())
	}

	for i, name := range names {
		names[i] = filepath.Join(strings.TrimSpace(outb.String()), name)
	}

	return godotenv.Load(names...) //nolint: wrapcheck
}
