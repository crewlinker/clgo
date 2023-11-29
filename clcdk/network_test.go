package clcdk_test

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("network creation", func() {
	var app awscdk.App
	var stack awscdk.Stack
	var cfg clcdk.Config

	BeforeEach(func() {
		app = awscdk.NewApp(nil)
		cfg = clcdk.NewStagingConfig()
		cfg = cfg.Copy(clcdk.WithMainIPSpace(awsec2.IpAddresses_Cidr(jsii.String(`100.0.0.0/16`))))
		stack = awscdk.NewStack(app, jsii.String("Stack1"), nil)
	})

	It("should create network", func() {
		clcdk.WithNetwork(stack, "Network1", cfg)

		tmpl := assertions.Template_FromStack(stack, nil)

		tmpl.ResourceCountIs(jsii.String("AWS::EC2::VPC"), jsii.Number(1))

		tmpl.HasResourceProperties(jsii.String("AWS::EC2::VPC"), map[string]any{
			"CidrBlock": jsii.String(`100.0.0.0/16`),
			"Tags": []map[string]any{
				{
					"Key":   jsii.String("Name"),
					"Value": jsii.String("Stack1Network1"),
				},
			},
		})
	})
})
