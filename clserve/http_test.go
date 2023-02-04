package clserve_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClServe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clserve")
}
