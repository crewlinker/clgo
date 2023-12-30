package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/sourcegraph/conc/iter"
)

// Infra groups commands for infrastructure deployment.
type Infra mg.Namespace

// Build infra artifacts for deployment.
func (Infra) Build() error {
	const buildDirPerm = 0o0700

	if err := os.MkdirAll(filepath.Join("clcdk", "builds"), buildDirPerm); err != nil {
		return fmt.Errorf("failed to create build dir: %w", err)
	}

	// build go lambdas
	if err := errors.Join(iter.Map([]string{
		"github.com/crewlinker/clgo/clcdk/postgresresource:postgresresource",
	}, func(it *string) error {
		pkgp, pkgn, found := strings.Cut(*it, ":")
		if !found {
			pkgn = filepath.Base(pkgp)
		}

		return buildGoLambda(
			pkgp,
			strings.ReplaceAll(pkgn, "/", "_"),
			pkgn,
		)
	})...); err != nil {
		return fmt.Errorf("failed to build lambdas: %w", err)
	}

	return nil
}

// determineBuildVersion provides the build version.
func determineBuildVersion() (string, error) {
	version := os.Getenv("BUILD_VERSION")
	if version != "" {
		return version, nil
	}

	sha, err := sh.Output("git", "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to run git: %w", err)
	}

	return fmt.Sprintf("v0.0.0-%s", sha[:7]), nil
}

// buildGoLambda builds a single lambda function's binary.
func buildGoLambda(pkgPath, dstDirName, pkgName string) error {
	dstdir := filepath.Join("clcdk", "builds", dstDirName)
	if err := os.MkdirAll(dstdir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to mkdir destination: %w", err)
	}

	version, err := determineBuildVersion()
	if err != nil {
		return fmt.Errorf("failed to determine version: %w", err)
	}

	tmpdir, err := os.MkdirTemp("", "lambda_build_*")
	if err != nil {
		return fmt.Errorf("failed to init temp dir: %w", err)
	}

	if err = os.WriteFile(filepath.Join(dstdir, "main.go"), []byte(fmt.Sprintf(`
	package main; 
	import (
		%s "%s"
		"go.uber.org/fx"
	)
	var Version = "0.0.0"
	func main(){
		fx.New(%s.Provide(Version)).Run()
	}
	`, pkgName, pkgPath, pkgName)), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write lambda main.go: %w", err)
	}

	err = runIfNoErr(err, nil, "rm", "-f", filepath.Join(dstdir, "pkg.zip"))
	err = runIfNoErr(err, map[string]string{"GOOS": "linux", "GOARCH": "arm64"},
		"go", "build", "-trimpath", "-tags", "lambda.norpc", "-ldflags", `-X 'main.Version=`+version+"'",
		"-o", filepath.Join(tmpdir, "bootstrap"), "./"+dstdir)
	err = runIfNoErr(err, nil, "touch", "-t", "200906122350", filepath.Join(tmpdir, "bootstrap"))
	err = runIfNoErr(err, nil, "zip", "-r", "-j", "--latest-time", "-X", filepath.Join(dstdir, "pkg.zip"), tmpdir)
	err = runIfNoErr(err, nil, "rm", filepath.Join(dstdir, "main.go"))

	return err
}

// runIfNoErr will only run cmd with args if 'err' is nil, else it will return err. This allows us to
// make somewhat readable automation around scripts.
func runIfNoErr(err error, env map[string]string, cmd string, args ...string) error {
	if err != nil {
		return err
	}

	if err = sh.RunWith(env, cmd, args...); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
