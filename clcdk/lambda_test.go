package clcdk_test

import (
	"path/filepath"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("lambda creation", func() {
	var app awscdk.App
	var stack awscdk.Stack
	var code awslambda.AssetCode
	var cfg clcdk.Config

	BeforeEach(func() {
		app = awscdk.NewApp(nil)
		cfg = clcdk.NewStagingConfig()
		stack = awscdk.NewStack(app, jsii.String("Stack1"), nil)
		code = awslambda.AssetCode_FromAsset(jsii.String(
			filepath.Join("testdata", "pkg1.zip")), nil)
	})

	It("should create native lambda", func() {
		clcdk.WithNativeLambda(stack, "Lambda1", cfg, code, nil, nil)

		tmpl := assertions.Template_FromStack(stack, nil)

		tmpl.ResourceCountIs(jsii.String("AWS::Lambda::Function"), jsii.Number(1))
		tmpl.ResourceCountIs(jsii.String("AWS::Lambda::Alias"), jsii.Number(1))
		tmpl.ResourceCountIs(jsii.String("AWS::Logs::LogGroup"), jsii.Number(1))

		tmpl.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
			"Handler": jsii.String("bootstrap"),
			"Runtime": jsii.String("provided.al2023"),
		})
	})

	It("should create Node lambda", func() {
		clcdk.WithNodeLambbda(stack, "Lambda2", cfg, code, nil, nil)

		tmpl := assertions.Template_FromStack(stack, nil)

		tmpl.ResourceCountIs(jsii.String("AWS::Lambda::Function"), jsii.Number(1))
		tmpl.ResourceCountIs(jsii.String("AWS::Lambda::Alias"), jsii.Number(1))
		tmpl.ResourceCountIs(jsii.String("AWS::Logs::LogGroup"), jsii.Number(1))

		tmpl.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
			"Handler": jsii.String("index.handler"),
			"Runtime": jsii.String("nodejs20.x"),
		})
	})

	It("should re-use provided logs", func() {
		logs := awslogs.NewLogGroup(stack, jsii.String("Logs1"), nil)
		clcdk.WithNodeLambbda(stack, "Dar", cfg, code, nil, logs)

		tmpl := assertions.Template_FromStack(stack, nil)

		tmpl.ResourceCountIs(jsii.String("AWS::Logs::LogGroup"), jsii.Number(1))
	})
})
