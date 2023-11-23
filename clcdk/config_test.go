package clcdk_test

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", Serial, func() {
	It("should stringify scope names", func() {
		stag1 := clcdk.NewStagingConfig()
		stag2 := stag1.Copy(
			clcdk.WithLambdaApplicationLogLevel(jsii.String("INFO")),
			clcdk.WithLambdaTimeout(awscdk.Duration_Seconds(jsii.Number(100))),
		)

		Expect(*stag1.LambdaApplicationLogLevel()).To(Equal("DEBUG")) // should not have changed
		Expect(*stag2.LambdaApplicationLogLevel()).To(Equal("INFO"))  // should have changed
		Expect(*stag2.LambdaSystemLogLevel()).To(Equal("DEBUG"))      // should not have changed

		Expect(*stag1.LambdaTimeout().ToString()).To(Equal(`Duration.seconds(10)`))
		Expect(*stag2.LambdaTimeout().ToString()).To(Equal(`Duration.seconds(100)`))
	})
})
