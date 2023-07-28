//go:build tools

package tools

import (
	_ "github.com/magefile/mage"
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
