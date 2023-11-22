package clcdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/jsii-runtime-go"
)

// Config describes the providing of resource configuration that is often convenient
// to be shared between branches of the resource tree.
type Config interface {
	LogRetention() awslogs.RetentionDays
	GatewayThrottlingRateLimit() *float64
	GatewayThrottlingBurstLimit() *float64
	DomainRecordTTL() awscdk.Duration
	LambdaTimeout() awscdk.Duration
	LambdaReservedConcurrency() *float64
	LambdaProvisionedConcurrency() *float64
	LambdaApplicationLogLevel() string
	LambdaSystemLogLevel() string
}

type config struct {
	logRetention                 awslogs.RetentionDays
	gatewayThrottlingRateLimit   *float64
	gatewayThrottlingBurstLimit  *float64
	domainRecordTTL              awscdk.Duration
	lambdaTimeout                awscdk.Duration
	lambdaReservedConcurrency    *float64
	lambdaProvisionedConcurrency *float64
	lambdaApplicationLogLevel    string
	lambdaSystemLogLevel         string
}

// NewConfig initializes a config implementation given the provided values.
func NewConfig(
	logRetention awslogs.RetentionDays,
	gatewayThrottlingRateLimit *float64,
	gatewayThrottlingBurstLimit *float64,
	domainRecordTTL awscdk.Duration,
	lambdaTimeout awscdk.Duration,
	lambdaReservedConcurrency *float64,
	lambdaProvisionedConcurrency *float64,
	lambdaApplicationLogLevel string,
	lambdaSystemLogLevel string,
) Config {
	return config{
		logRetention:                 logRetention,
		gatewayThrottlingRateLimit:   gatewayThrottlingRateLimit,
		gatewayThrottlingBurstLimit:  gatewayThrottlingBurstLimit,
		domainRecordTTL:              domainRecordTTL,
		lambdaTimeout:                lambdaTimeout,
		lambdaReservedConcurrency:    lambdaReservedConcurrency,
		lambdaProvisionedConcurrency: lambdaProvisionedConcurrency,
		lambdaApplicationLogLevel:    lambdaApplicationLogLevel,
		lambdaSystemLogLevel:         lambdaSystemLogLevel,
	}
}

// LogRetention config.
func (c config) LogRetention() awslogs.RetentionDays { return c.logRetention }

// LambdaTimeout config.
func (c config) LambdaTimeout() awscdk.Duration { return c.lambdaTimeout }

// LambdaReservedConcurrency config.
func (c config) LambdaReservedConcurrency() *float64 { return c.lambdaReservedConcurrency }

// LambdaProvisionedConcurrency config.
func (c config) LambdaProvisionedConcurrency() *float64 { return c.lambdaProvisionedConcurrency }

// LambdaApplicationLogLevel config.
func (c config) LambdaApplicationLogLevel() string { return c.lambdaApplicationLogLevel }

// LambdaSystemLogLevel config.
func (c config) LambdaSystemLogLevel() string { return c.lambdaSystemLogLevel }

// GatewayThrottlingRateLimit config.
func (c config) GatewayThrottlingRateLimit() *float64 { return c.gatewayThrottlingRateLimit }

// GatewayThrottlingBurstLimit config.
func (c config) GatewayThrottlingBurstLimit() *float64 { return c.gatewayThrottlingBurstLimit }

// DomainRecordTTL config.
func (c config) DomainRecordTTL() awscdk.Duration { return c.domainRecordTTL }

// NewStagingConfig provides a config that provides easy-to-use defeaults for a staging environment.
func NewStagingConfig() Config {
	return NewConfig(
		awslogs.RetentionDays_FIVE_DAYS,
		jsii.Number(100),                         //nolint:gomnd
		jsii.Number(200),                         //nolint:gomnd
		awscdk.Duration_Seconds(jsii.Number(60)), //nolint:gomnd
		awscdk.Duration_Seconds(jsii.Number(10)), //nolint:gomnd
		jsii.Number(5),                           //nolint:gomnd
		jsii.Number(0),
		"DEBUG",
		"DEBUG",
	)
}
