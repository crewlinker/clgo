// Package clcdk contains reusable infrastructruce components using AWS CDK.
package clcdk

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// StagingConfig provides static configuration based on something reasonable for staging.
type StagingConfig struct{}

// LogRetention config.
func (c StagingConfig) LogRetention() awslogs.RetentionDays { return awslogs.RetentionDays_FIVE_DAYS }

// LambdaTimeout config.
func (c StagingConfig) LambdaTimeout() awscdk.Duration {
	return awscdk.Duration_Seconds(jsii.Number(10)) //nolint:gomnd
}

// LambdaReservedConcurrency config.
func (c StagingConfig) LambdaReservedConcurrency() *float64 { return jsii.Number(5) } //nolint:gomnd

// LambdaProvisionedConcurrency config.
func (c StagingConfig) LambdaProvisionedConcurrency() *float64 { return jsii.Number(0) }

// LambdaApplicationLogLevel config.
func (c StagingConfig) LambdaApplicationLogLevel() string { return "DEBUG" }

// LambdaSystemLogLevel config.
func (c StagingConfig) LambdaSystemLogLevel() string { return "DEBUG" }

// GatewayThrottlingRateLimit config.
func (c StagingConfig) GatewayThrottlingRateLimit() *float64 { return jsii.Number(100) } //nolint:gomnd

// GatewayThrottlingBurstLimit config.
func (c StagingConfig) GatewayThrottlingBurstLimit() *float64 { return jsii.Number(200) } //nolint:gomnd

// DomainRecordTTL config.
func (c StagingConfig) DomainRecordTTL() awscdk.Duration {
	return awscdk.Duration_Seconds(jsii.Number(60)) //nolint:gomnd
}

// ScopeName is the name of a scope.
type ScopeName string

// ChildScope returns a new scope named 'name'.
func (sn ScopeName) ChildScope(parent constructs.Construct) constructs.Construct {
	return constructs.NewConstruct(parent, jsii.String(sn.String()))
}

func (sn ScopeName) String() string {
	return fmt.Sprintf("[%s]", string(sn))
}
