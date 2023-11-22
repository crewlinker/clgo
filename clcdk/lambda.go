package clcdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// sub-set of the total interface for lamba config.
type lambdaConfig interface {
	LogRetention() awslogs.RetentionDays
	LambdaTimeout() awscdk.Duration
	LambdaReservedConcurrency() *float64
	LambdaProvisionedConcurrency() *float64
	LambdaApplicationLogLevel() string
	LambdaSystemLogLevel() string
}

// WithNativeLambda creates a lambda for code that compiles Natively (such as Go).
func WithNativeLambda(
	scope constructs.Construct,
	name ScopeName,
	cfg lambdaConfig,
	code awslambda.AssetCode,
	env *map[string]*string,
	logs awslogs.ILogGroup,
) awslambda.IFunction {
	scope = name.ChildScope(scope)

	return withLambda(
		scope, cfg, code, jsii.String("bootstrap"), awslambda.Runtime_PROVIDED_AL2023(), env, logs,
	)
}

// WithNodeLambbda creates a NodeJS lambda with a default alias.
func WithNodeLambbda(
	scope constructs.Construct,
	name ScopeName,
	cfg lambdaConfig,
	code awslambda.AssetCode,
	env *map[string]*string,
	logs awslogs.ILogGroup,
) awslambda.IFunction {
	scope = name.ChildScope(scope)

	return withLambda(
		scope, cfg, code,
		jsii.String("index.handler"),
		awslambda.Runtime_NODEJS_20_X(), env, logs,
	)
}

// withLambda creates a standard lambda with a default alias.
func withLambda(
	scope constructs.Construct,
	cfg lambdaConfig,
	code awslambda.AssetCode,
	hdlr *string,
	runtime awslambda.Runtime,
	env *map[string]*string,
	logs awslogs.ILogGroup,
) awslambda.IFunction {
	if logs == nil {
		logs = awslogs.NewLogGroup(scope, jsii.String("Logs"), &awslogs.LogGroupProps{
			Retention: cfg.LogRetention(),
		})
	}

	handler := awslambda.NewFunction(scope, jsii.String("Handler"), &awslambda.FunctionProps{
		Code:                         code,
		Handler:                      hdlr,
		Runtime:                      runtime,
		Timeout:                      cfg.LambdaTimeout(),
		ReservedConcurrentExecutions: cfg.LambdaReservedConcurrency(),
		Architecture:                 awslambda.Architecture_ARM_64(),
		Tracing:                      awslambda.Tracing_ACTIVE,

		LogGroup:            logs,
		LogFormat:           jsii.String("JSON"),
		ApplicationLogLevel: jsii.String(cfg.LambdaApplicationLogLevel()),
		SystemLogLevel:      jsii.String(cfg.LambdaSystemLogLevel()),
		Environment:         env,
	})

	return awslambda.NewAlias(scope, jsii.String("Alias"), &awslambda.AliasProps{
		AliasName:                       jsii.String("Default"),
		Version:                         handler.CurrentVersion(),
		ProvisionedConcurrentExecutions: cfg.LambdaProvisionedConcurrency(),
	})
}
