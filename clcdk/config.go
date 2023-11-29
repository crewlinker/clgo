package clcdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/jsii-runtime-go"
	"github.com/mitchellh/copystructure"
)

// Config describes the providing of resource configuration that is often convenient
// to be shared between branches of the resource tree.
//
//nolint:interfacebloat
type Config interface {
	Copy(opts ...ConfigOpt) Config

	LogRetention() awslogs.RetentionDays
	DomainRecordTTL() awscdk.Duration

	GatewayThrottlingRateLimit() *float64
	GatewayThrottlingBurstLimit() *float64
	GatewayDisableExecuteApi() *bool

	LambdaTimeout() awscdk.Duration
	LambdaReservedConcurrency() *float64
	LambdaProvisionedConcurrency() *float64
	LambdaApplicationLogLevel() *string
	LambdaSystemLogLevel() *string

	MainDomainName() *string
	RegionalCertificateArn() *string
	EdgeCertificateArn() *string
	MainIPSpace() awsec2.IIpAddresses
}

type config struct {
	LogRetentionVal                 awslogs.RetentionDays `copy:"shallow"`
	DomainRecordTTLVal              awscdk.Duration       `copy:"shallow"`
	LambdaTimeoutVal                awscdk.Duration       `copy:"shallow"`
	GatewayThrottlingRateLimitVal   *float64
	GatewayThrottlingBurstLimitVal  *float64
	LambdaReservedConcurrencyVal    *float64
	LambdaProvisionedConcurrencyVal *float64
	LambdaApplicationLogLevelVal    *string
	LambdaSystemLogLevelVal         *string
	GatewayDisableExecuteApiVal     *bool

	MainDomainNameVal         *string
	RegionalCertificateArnVal *string
	EdgeCertificateArnVal     *string
	MainIPSpaceVal            awsec2.IIpAddresses
}

// ConfigOpts describes a configuration option.
type ConfigOpt func(*config)

// WithMainIPSpace config.
func WithMainIPSpace(v awsec2.IIpAddresses) ConfigOpt {
	return func(c *config) { c.MainIPSpaceVal = v }
}

// WithMainDomainName config.
func WithMainDomainName(v *string) ConfigOpt {
	return func(c *config) { c.MainDomainNameVal = v }
}

// WithRegionalCertificateArn config.
func WithRegionalCertificateArn(v *string) ConfigOpt {
	return func(c *config) { c.RegionalCertificateArnVal = v }
}

// WithEdgeCertificateArn config.
func WithEdgeCertificateArn(v *string) ConfigOpt {
	return func(c *config) { c.EdgeCertificateArnVal = v }
}

// WithLogRetention config.
func WithLogRetention(v awslogs.RetentionDays) ConfigOpt {
	return func(c *config) { c.LogRetentionVal = v }
}

// WithGatewayThrottlingRateLimit config.
func WithGatewayThrottlingRateLimit(v *float64) ConfigOpt {
	return func(c *config) { c.GatewayThrottlingRateLimitVal = v }
}

// WithGatewayThrottlingBurstLimit config.
func WithGatewayThrottlingBurstLimit(v *float64) ConfigOpt {
	return func(c *config) { c.GatewayThrottlingBurstLimitVal = v }
}

// WithDomainRecordTTL config.
func WithDomainRecordTTL(v awscdk.Duration) ConfigOpt {
	return func(c *config) { c.DomainRecordTTLVal = v }
}

// WithLambdaTimeout config.
func WithLambdaTimeout(v awscdk.Duration) ConfigOpt {
	return func(c *config) { c.LambdaTimeoutVal = v }
}

// WithLambdaReservedConcurrency config.
func WithLambdaReservedConcurrency(v *float64) ConfigOpt {
	return func(c *config) { c.LambdaReservedConcurrencyVal = v }
}

// WithLambdaProvisionedConcurrency config.
func WithLambdaProvisionedConcurrency(v *float64) ConfigOpt {
	return func(c *config) { c.LambdaProvisionedConcurrencyVal = v }
}

// WithLambdaApplicationLogLevel config.
func WithLambdaApplicationLogLevel(v *string) ConfigOpt {
	return func(c *config) { c.LambdaApplicationLogLevelVal = v }
}

// WithLambdaSystemLogLevel config.
func WithLambdaSystemLogLevel(v *string) ConfigOpt {
	return func(c *config) { c.LambdaSystemLogLevelVal = v }
}

// WithGatewayDisableExecuteApi config.
func WithGatewayDisableExecuteApi(v *bool) ConfigOpt {
	return func(c *config) { c.GatewayDisableExecuteApiVal = v }
}

// NewConfig initializes a config implementation given the provided values.
func NewConfig(opts ...ConfigOpt) Config {
	cfg := config{}
	for _, o := range opts {
		o(&cfg)
	}

	return cfg
}

// CopyReturns a copy of the config while allowing certain options to be changed.
func (c config) Copy(opts ...ConfigOpt) Config {
	v, err := copystructure.Copy(c)
	if err != nil {
		panic("clcdk: failed to deep copy: " + err.Error())
	}

	cfg, _ := v.(config)
	for _, o := range opts {
		o(&cfg)
	}

	return cfg
}

// LogRetention config.
func (c config) LogRetention() awslogs.RetentionDays { return c.LogRetentionVal }

// LambdaTimeout config.
func (c config) LambdaTimeout() awscdk.Duration { return c.LambdaTimeoutVal }

// LambdaReservedConcurrency config.
func (c config) LambdaReservedConcurrency() *float64 { return c.LambdaReservedConcurrencyVal }

// LambdaProvisionedConcurrency config.
func (c config) LambdaProvisionedConcurrency() *float64 { return c.LambdaProvisionedConcurrencyVal }

// LambdaApplicationLogLevel config.
func (c config) LambdaApplicationLogLevel() *string { return c.LambdaApplicationLogLevelVal }

// LambdaSystemLogLevel config.
func (c config) LambdaSystemLogLevel() *string { return c.LambdaSystemLogLevelVal }

// GatewayThrottlingRateLimit config.
func (c config) GatewayThrottlingRateLimit() *float64 { return c.GatewayThrottlingRateLimitVal }

// GatewayThrottlingBurstLimit config.
func (c config) GatewayThrottlingBurstLimit() *float64 { return c.GatewayThrottlingBurstLimitVal }

// GatewayDisableExecuteApi config.
func (c config) GatewayDisableExecuteApi() *bool { return c.GatewayDisableExecuteApiVal }

// DomainRecordTTL config.
func (c config) DomainRecordTTL() awscdk.Duration { return c.DomainRecordTTLVal }

// MainDomainName config.
func (c config) MainDomainName() *string { return c.MainDomainNameVal }

// MainIPSpace config.
func (c config) MainIPSpace() awsec2.IIpAddresses { return c.MainIPSpaceVal }

// RegionalCertificateArn config.
func (c config) RegionalCertificateArn() *string { return c.RegionalCertificateArnVal }

// EdgeCertificateArnArn config.
func (c config) EdgeCertificateArn() *string { return c.EdgeCertificateArnVal }

// NewStagingConfig provides a config that provides easy-to-use defeaults for a staging environment.
func NewStagingConfig() Config {
	return NewConfig(
		WithLogRetention(awslogs.RetentionDays_FIVE_DAYS),
		WithGatewayThrottlingRateLimit(jsii.Number(100)),              //nolint:gomnd
		WithGatewayThrottlingBurstLimit(jsii.Number(200)),             //nolint:gomnd
		WithDomainRecordTTL(awscdk.Duration_Seconds(jsii.Number(60))), //nolint:gomnd
		WithLambdaTimeout(awscdk.Duration_Seconds(jsii.Number(10))),   //nolint:gomnd
		WithLambdaReservedConcurrency(jsii.Number(5)),                 //nolint:gomnd
		WithLambdaProvisionedConcurrency(jsii.Number(0)),
		WithLambdaApplicationLogLevel(jsii.String("DEBUG")),
		WithLambdaSystemLogLevel(jsii.String("DEBUG")),
		WithGatewayDisableExecuteApi(jsii.Bool(false)),
	)
}
