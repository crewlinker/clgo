package clcdk_test

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("lookup", Serial, func() {
	var app awscdk.App
	var stack awscdk.Stack
	var inp clcdk.BaseInputs
	BeforeEach(func() {
		app = awscdk.NewApp(nil)
		stack = awscdk.NewStack(app, jsii.String("Stack1"), &awscdk.StackProps{Env: &awscdk.Environment{
			Account: jsii.String("1111111"),
			Region:  jsii.String("eu-bogus-1"),
		}})

		inp = clcdk.NewBaseInputs("foo.example.com", "regional:arn", "edge:arn")
	})

	It("should lookup zone and certs", func() {
		zone, regionalCert, edgeCert := clcdk.LookupBaseZoneAndCerts(stack, "ZoneAndCerts", inp)

		awscdk.NewCfnOutput(stack, jsii.String("ZoneName"), &awscdk.CfnOutputProps{Value: zone.ZoneName()})
		awscdk.NewCfnOutput(stack, jsii.String("RegCert"), &awscdk.CfnOutputProps{Value: regionalCert.CertificateArn()})
		awscdk.NewCfnOutput(stack, jsii.String("EdgeCert"), &awscdk.CfnOutputProps{Value: edgeCert.CertificateArn()})

		tmpl := assertions.Template_FromStack(stack, nil)
		tmpl.HasOutput(jsii.String("ZoneName"), map[string]any{"Value": jsii.String("foo.example.com")})
		tmpl.HasOutput(jsii.String("RegCert"), map[string]any{"Value": jsii.String("regional:arn")})
		tmpl.HasOutput(jsii.String("EdgeCert"), map[string]any{"Value": jsii.String("edge:arn")})
	})
})
