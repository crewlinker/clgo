package clcdk

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/cloapi"
)

// gatewayConfig describes the config required for setting up gateways.
type gatewayConfig interface {
	LogRetention() awslogs.RetentionDays
	GatewayThrottlingRateLimit() *float64
	GatewayThrottlingBurstLimit() *float64
	DomainRecordTTL() awscdk.Duration
}

// WithOpenApiGateway creates a gateway that is defined by an OpenAPI definition while proxying all
// requests to a single Lambda Handler. This is done by treating the schema a template that will have
// certain values replaced. It will pick up changes to this schema and trigger a re-deploy, but changes
// to the templated values are not picked up (because they are still Cloudformation tokens at this stage.
func WithOpenApiGateway(
	scope constructs.Construct,
	name ScopeName,
	cfg gatewayConfig,
	handler awslambda.IFunction,
	schemaTmpl string,
) awsapigateway.IRestApi {
	scope, stack := name.ChildScope(scope), awscdk.Stack_Of(scope)

	def, sum, err := cloapi.ExecuteSchemaTmpl([]byte(schemaTmpl), cloapi.SchemaDeployment{
		Title:       fmt.Sprintf("%s%sOpenApiGateway", *stack.StackName(), string(name)),
		Description: fmt.Sprintf("%s OpenApi gateway for stack %s", string(name), *stack.StackName()),
		AwsProxyIntegrationURI: fmt.Sprintf("arn:%s:apigateway:%s:lambda:path/2015-03-31/functions/%s/invocations",
			*awscdk.Stack_Of(scope).Partition(),
			*awscdk.Stack_Of(scope).Region(),
			*handler.FunctionArn()),
	})
	if err != nil {
		panic(fmt.Errorf("failed to execute schema template: %w", err))
	}

	logs := awslogs.NewLogGroup(scope, jsii.String("Logs"), &awslogs.LogGroupProps{
		Retention: cfg.LogRetention(),
	})

	// we use the stack's asset bucket to upload our definition. It assumes the name based on the defeault
	// synthesizer.
	assets := awss3.Bucket_FromBucketName(
		scope, jsii.String("Assets"),
		jsii.String(fmt.Sprintf("cdk-%s-assets-%s-%s",
			strings.ToLower(*stack.Synthesizer().BootstrapQualifier()),
			*stack.Account(),
			*stack.Region(),
		)),
	)

	key, prefix := fmt.Sprintf("%s_%x_api_def.json", *stack.StackName(), sum[:10]),
		strings.ToLower(string(name))+"_oapi_definitions/"

	deployment := awss3deployment.NewBucketDeployment(scope, jsii.String("Deployment"),
		&awss3deployment.BucketDeploymentProps{
			Prune:                jsii.Bool(false),
			DestinationBucket:    assets,
			DestinationKeyPrefix: jsii.String(string(prefix)),
			Sources: &[]awss3deployment.ISource{
				awss3deployment.Source_Data(jsii.String(key), jsii.String(def)),
			},
		})

	gateway := awsapigateway.NewSpecRestApi(scope, jsii.String("Gateway"), &awsapigateway.SpecRestApiProps{
		CloudWatchRole: jsii.Bool(true),
		RestApiName:    jsii.String(fmt.Sprintf("%s%sOpenApiGateway", *stack.StackName(), string(name))),
		Description:    jsii.String(fmt.Sprintf("%s OpenApi gateway for stack %s", string(name), *stack.StackName())),

		ApiDefinition: awsapigateway.ApiDefinition_FromBucket(
			deployment.DeployedBucket(),
			jsii.String(fmt.Sprintf("%s%s", prefix, key)), nil),
		EndpointTypes: &[]awsapigateway.EndpointType{
			awsapigateway.EndpointType_REGIONAL,
		},

		DeployOptions: gatewayDeployOptions(cfg, logs),
	})

	handler.AddPermission(jsii.String("AllowGateway"), &awslambda.Permission{
		Principal: awsiam.NewServicePrincipal(jsii.String("apigateway.amazonaws.com"), nil),
		Action:    jsii.String("lambda:InvokeFunction"),
		SourceArn: jsii.String(fmt.Sprintf("arn:%s:execute-api:%s:%s:%s/*",
			*awscdk.Stack_Of(scope).Partition(),
			*awscdk.Stack_Of(scope).Region(),
			*awscdk.Stack_Of(scope).Account(),
			*gateway.RestApiId())),
	})

	return gateway
}

// WithProxyGateway will setup a gateway that proxies all requests to a Lambda, with logging and
// tracing enabled.
func WithProxyGateway(
	scope constructs.Construct, name ScopeName, cfg gatewayConfig, handler awslambda.IFunction,
) awsapigateway.IRestApi {
	scope, stack := name.ChildScope(scope), awscdk.Stack_Of(scope)

	logs := awslogs.NewLogGroup(scope, jsii.String("Logs"), &awslogs.LogGroupProps{
		Retention: cfg.LogRetention(),
	})

	return awsapigateway.NewLambdaRestApi(scope, jsii.String("Gateway"), &awsapigateway.LambdaRestApiProps{
		CloudWatchRole: jsii.Bool(true),
		RestApiName:    jsii.String(fmt.Sprintf("%s%sProxyGateway", *stack.StackName(), string(name))),
		Description:    jsii.String(fmt.Sprintf("%s Proxy gateway for stack %s", string(name), *stack.StackName())),
		Handler:        handler,
		EndpointTypes: &[]awsapigateway.EndpointType{
			awsapigateway.EndpointType_REGIONAL,
		},
		DeployOptions: gatewayDeployOptions(cfg, logs),
	})
}

// WithGatewayDomain will setup a domain for the gateway on the provided hosted zone.
func WithGatewayDomain(
	scope constructs.Construct,
	name ScopeName,
	cfg gatewayConfig,
	gateway awsapigateway.IRestApi,
	zone awsroute53.IHostedZone,
	subDomain string,
	basePath *string,
) awsapigateway.IDomainName {
	scope = name.ChildScope(scope)

	fullDomain := subDomain + "." + *zone.ZoneName()

	cert := awscertificatemanager.NewCertificate(scope, jsii.String("Certificate"),
		&awscertificatemanager.CertificateProps{
			DomainName: jsii.String(fullDomain),
			Validation: awscertificatemanager.CertificateValidation_FromDns(zone),
		})

	domain := awsapigateway.NewDomainName(scope, jsii.String("Domain"), &awsapigateway.DomainNameProps{
		DomainName:     jsii.String(fullDomain),
		Certificate:    cert,
		EndpointType:   awsapigateway.EndpointType_REGIONAL,
		SecurityPolicy: awsapigateway.SecurityPolicy_TLS_1_2,
	})

	domain.AddApiMapping(gateway.DeploymentStage(), &awsapigateway.ApiMappingOptions{
		BasePath: basePath,
	})

	awsroute53.NewCnameRecord(scope, jsii.String("DnsRecord"), &awsroute53.CnameRecordProps{
		Zone:       zone,
		DomainName: domain.DomainNameAliasDomainName(),
		RecordName: jsii.String(subDomain),
		Ttl:        cfg.DomainRecordTTL(),
	})

	return domain
}

func gatewayDeployOptions(cfg gatewayConfig, logs awslogs.ILogGroup) *awsapigateway.StageOptions {
	return &awsapigateway.StageOptions{
		ThrottlingRateLimit:  cfg.GatewayThrottlingRateLimit(),
		ThrottlingBurstLimit: cfg.GatewayThrottlingBurstLimit(),
		TracingEnabled:       jsii.Bool(true),
		DataTraceEnabled:     jsii.Bool(true),
		LoggingLevel:         awsapigateway.MethodLoggingLevel_INFO,
		AccessLogDestination: awsapigateway.NewLogGroupLogDestination(logs),
		AccessLogFormat: awsapigateway.AccessLogFormat_JsonWithStandardFields(
			&awsapigateway.JsonWithStandardFieldProps{
				Caller:         jsii.Bool(true),
				HttpMethod:     jsii.Bool(true),
				Ip:             jsii.Bool(true),
				Protocol:       jsii.Bool(true),
				RequestTime:    jsii.Bool(true),
				ResourcePath:   jsii.Bool(true),
				ResponseLength: jsii.Bool(true),
				Status:         jsii.Bool(true),
				User:           jsii.Bool(true),
			},
		),
	}
}
