package clcdk_test

import (
	"bytes"
	"encoding/json"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"

	. "github.com/onsi/ginkgo/v2"
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
		clcdk.WithPostgresInstance(stack, "Postgres1", cfg, vpc)

		tmpl := assertions.Template_FromStack(stack, nil)

		tmpl.ResourceCountIs(jsii.String("AWS::RDS::DBInstance"), jsii.Number(1))
		tmpl.ResourceCountIs(jsii.String("AWS::RDS::DBParameterGroup"), jsii.Number(1))
		tmpl.ResourceCountIs(jsii.String("AWS::EC2::SecurityGroup"), jsii.Number(1))
	})

	It("should create custom resource provider", func() {
		vpc := clcdk.WithNetwork(stack, "Network1", cfg)
		db, dbSecret := clcdk.WithPostgresInstance(stack, "Postgres1", cfg, vpc)
		clcdk.WithPostgresCustomResources(stack, "PgCustom1", cfg, db, dbSecret)

		tmpl := assertions.Template_FromStack(stack, nil)

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", " ")
		enc.Encode(tmpl.ToJSON())

		tmpl.ResourceCountIs(jsii.String("AWS::Lambda::Function"), jsii.Number(3))

		tmpl.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
			"Description": "AWS CDK resource provider framework - onEvent (Stack1/PgCustom1/Provider)",
		})
	})
})
