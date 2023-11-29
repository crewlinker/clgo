package clcdk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewSingletonStack requires a "instance" context variable to allow different copies of the stack
// to exist in the same AWS account.
func NewSingletonStack(scope constructs.Construct, conv Conventions) awscdk.Stack {
	env := EnvironmentFromScope(scope)
	if env == "" {
		env = "<none>"
	}

	return awscdk.NewStack(scope,
		jsii.String(conv.SingletonStackName()),
		&awscdk.StackProps{
			CrossRegionReferences: jsii.Bool(true),
			Env: &awscdk.Environment{
				Account: jsii.String(conv.Account()),
				Region:  jsii.String(conv.MainRegion()),
			},
			Description: jsii.String(fmt.Sprintf("%s (env: %s, singleton)",
				conv.Qualifier(), env)),
			Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
				Qualifier: jsii.String(strings.ToLower(conv.Qualifier())),
			}),
		})
}

// NewInstancedStack requires a "instance" context variable to allow different copies of the stack
// to exist in the same AWS account.
func NewInstancedStack(scope constructs.Construct, conv Conventions) awscdk.Stack {
	instance, env := InstanceFromScope(scope), EnvironmentFromScope(scope)
	if env == "" {
		env = "<none>"
	}

	return awscdk.NewStack(scope,
		jsii.String(conv.InstancedStackName(instance)),
		&awscdk.StackProps{
			CrossRegionReferences: jsii.Bool(true),
			Env: &awscdk.Environment{
				Account: jsii.String(conv.Account()),
				Region:  jsii.String(conv.MainRegion()),
			},
			Description: jsii.String(fmt.Sprintf("%s (env: %s, instance: %d)",
				conv.Qualifier(), env, instance)),
			Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
				Qualifier: jsii.String(strings.ToLower(conv.Qualifier())),
			}),
		})
}

// EnvironmentFromScope retrieves the instance name from the context or an empty string.
func EnvironmentFromScope(s constructs.Construct) string {
	v, _ := s.Node().TryGetContext(jsii.String("environment")).(string)

	return v
}

// InstanceFromScope retrieves the instance name from the context or an empty string.
func InstanceFromScope(s constructs.Construct) int {
	return tryGetCtxNr(s, "instance")
}

// tryGetCtxNr reads a contextual nr by the provided 'name'.
func tryGetCtxNr(s constructs.Construct, name string) int {
	nrv, _ := s.Node().TryGetContext(jsii.String(name)).(string)
	if nrv == "" {
		nrv = "0"
	}

	n, err := strconv.Atoi(nrv)
	if err != nil {
		panic("instance number isn't a number: " + nrv)
	}

	return n
}
