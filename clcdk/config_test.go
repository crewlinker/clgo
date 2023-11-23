package clcdk_test

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", Serial, func() {
	It("should stringify scope names", func() {
		stag1 := clcdk.NewStagingConfig()
		stag2 := stag1.Copy(clcdk.WithLambdaApplicationLogLevel(jsii.String("INFO")))

		Expect(*stag1.LambdaApplicationLogLevel()).To(Equal("DEBUG")) // should hot have changed
		Expect(*stag2.LambdaApplicationLogLevel()).To(Equal("INFO"))  // should have changed
		Expect(*stag2.LambdaSystemLogLevel()).To(Equal("DEBUG"))      // should not have changed
	})
})
