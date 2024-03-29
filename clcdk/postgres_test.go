package clcdk_test

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("postgres", func() {
	var app awscdk.App
	var stack awscdk.Stack
	var cfg clcdk.Config

	BeforeEach(func() {
		app = awscdk.NewApp(nil)
		cfg = clcdk.NewStagingConfig()
		cfg = cfg.Copy(clcdk.WithMainIPSpace(awsec2.IpAddresses_Cidr(jsii.String(`100.0.0.0/16`))))
		stack = awscdk.NewStack(app, jsii.String("Stack1"), nil)
	})

	It("should create instance", func() {
		vpc := clcdk.WithNetwork(stack, "Network1", cfg)
		clcdk.WithPostgresInstance(stack, "Postgres1", cfg, vpc, 10, 30)

		tmpl := assertions.Template_FromStack(stack, nil)

		tmpl.ResourceCountIs(jsii.String("AWS::RDS::DBInstance"), jsii.Number(1))
		tmpl.ResourceCountIs(jsii.String("AWS::RDS::DBParameterGroup"), jsii.Number(1))
		tmpl.ResourceCountIs(jsii.String("AWS::EC2::SecurityGroup"), jsii.Number(1))

		tmpl.HasResourceProperties(jsii.String("AWS::RDS::DBInstance"), map[string]any{
			"AllocatedStorage":    "10",
			"MaxAllocatedStorage": 30,
		})
	})

	It("should create a custom tenant resource", func() {
		secret := awssecretsmanager.NewSecret(stack, jsii.String("Secret1"), &awssecretsmanager.SecretProps{})

		tenant := clcdk.WithPostgresTenant(stack, "Tenant1", jsii.String("Token1"), secret, "tenant_1")

		tmpl := assertions.Template_FromStack(stack, nil)
		tmpl.ResourceCountIs(jsii.String("Custom::CrewlinkerPostgresTenant"), jsii.Number(1))

		Expect(*tenant.DatabaseName()).To(Not(BeEmpty()))
		Expect(*tenant.DatabaseUser()).To(Not(BeEmpty()))
		Expect(*tenant.DatabaseUser()).ToNot(Equal(*tenant.DatabaseName()))
	})
})
