package clcdk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewInstancedStack requires a "instance" context variable to allow different copies of the stack
// to exist in the same AWS account.
func NewInstancedStack(scope constructs.Construct, conv Conventions, account string) awscdk.Stack {
	instance := InstanceFromScope(scope)

	return awscdk.NewStack(scope,
		jsii.String(conv.InstancedStackName(instance)),
		&awscdk.StackProps{
			CrossRegionReferences: jsii.Bool(true),
			Env: &awscdk.Environment{
				Account: jsii.String(account),
				Region:  jsii.String(conv.MainRegion()),
			},
			Description: jsii.String(fmt.Sprintf("%s (instance: %d)",
				conv.Qualifier(), instance)),
			Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
				Qualifier: jsii.String(strings.ToLower(conv.Qualifier())),
			}),
		})
}

// InstanceFromScope retrieves the instance name from the context or an empty string.
func InstanceFromScope(s constructs.Construct) int {
	return tryGetCtxNr(s, "instance")
}

// tryGetCtxNr reads a contextual nr by the provided 'name'.
func tryGetCtxNr(s constructs.Construct, name string) int {
	nrv, _ := s.Node().TryGetContext(jsii.String(name)).(string)
	if nrv == "" {
		panic("instance number not in context")
	}

	n, err := strconv.Atoi(nrv)
	if err != nil {
		panic("instance number isn't a number: " + nrv)
	}

	return n
}
