package clcdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// WithNetwork sets up our vpc and other networking.
func WithNetwork(
	scope constructs.Construct,
	name ScopeName,
	cfg Config,
) awsec2.IVpc {
	const maxAzs = 2

	vpc := awsec2.NewVpc(scope, jsii.String("Vpc"), &awsec2.VpcProps{
		MaxAzs:      jsii.Number(maxAzs), // two is enough, in case we need to setup nat gateways (which are expensive)
		NatGateways: jsii.Number(0),      // simpler and saves costs, run everything in public subnet
		IpAddresses: cfg.MainIPSpace(),
		VpcName:     jsii.String(*awscdk.Stack_Of(scope).StackName() + string(name)),
	})

	return vpc
}
