package clcdk_test

import (
	"path/filepath"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("gateway", func() {
	var app awscdk.App
	var stack awscdk.Stack
	var cfg clcdk.Config
	var code awslambda.AssetCode
	var zone awsroute53.IHostedZone

	BeforeEach(func() {
		cfg = clcdk.NewStagingConfig()
		app = awscdk.NewApp(nil)
		stack = awscdk.NewStack(app, jsii.String("Stack1"), nil)
		code = awslambda.AssetCode_FromAsset(jsii.String(
			filepath.Join("testdata", "pkg1.zip")), nil)
		zone = awsroute53.NewPublicHostedZone(stack, jsii.String("Zone1"),
			&awsroute53.PublicHostedZoneProps{
				ZoneName: jsii.String("stag.example.com"),
			})
	})

	It("proxy gateway", func() {
		handler := clcdk.WithNativeLambda(stack, "Lambda1", cfg, code, nil, nil)
		gateway := clcdk.WithProxyGateway(stack, "Gateway1", cfg, handler)
		clcdk.WithGatewayDomain(stack, "Domain1", cfg, gateway, zone, "api", jsii.String("v1"))
		tmpl := assertions.Template_FromStack(stack, nil)

		tmpl.ResourceCountIs(jsii.String("AWS::ApiGateway::RestApi"), jsii.Number(1))
		tmpl.ResourceCountIs(jsii.String("AWS::Logs::LogGroup"), jsii.Number(2))

		tmpl.HasResourceProperties(jsii.String("AWS::ApiGateway::RestApi"), map[string]any{
			"Name":        jsii.String("Stack1Gateway1ProxyGateway"),
			"Description": jsii.String("Gateway1 Proxy gateway for stack Stack1"),
		})

		tmpl.HasResourceProperties(jsii.String("AWS::ApiGateway::DomainName"), map[string]any{
			"DomainName": jsii.String("api.stag.example.com"),
		})

		tmpl.HasResourceProperties(jsii.String("AWS::CertificateManager::Certificate"), map[string]any{
			"DomainName": jsii.String("api.stag.example.com"),
		})

		tmpl.HasResourceProperties(jsii.String("AWS::ApiGatewayV2::ApiMapping"), map[string]any{
			"ApiMappingKey": jsii.String("v1"),
		})

		tmpl.HasResourceProperties(jsii.String("AWS::Route53::RecordSet"), map[string]any{
			"Name": jsii.String("api.stag.example.com."),
			"Type": jsii.String("CNAME"),
		})
	})

	It("openapi gateway", func() {
		handler := clcdk.WithNativeLambda(stack, "Lambda1", cfg, code, nil, nil)
		clcdk.WithOpenApiGateway(stack, "Gateway1", cfg, handler, ``)
		tmpl := assertions.Template_FromStack(stack, nil)

		tmpl.HasResourceProperties(jsii.String("AWS::ApiGateway::RestApi"), map[string]any{
			"Name": jsii.String("Stack1Gateway1OpenApiGateway"),
			"BodyS3Location": map[string]any{
				"Key": jsii.String("gateway1_oapi_definitions/Stack1_e3b0c44298fc1c149afb_api_def.json"),
			},
		})
	})
})
