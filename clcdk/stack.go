package clcdk

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// RegionAcronyms holds acronyms for building CDK identifiers.
var RegionAcronyms = map[string]string{
	"eu-central-1":   "Fra",
	"eu-west-1":      "Ire",
	"ap-southeast-1": "Sin",
}

// FormatDockerImageTag standardizes how docker images in the repositories are called given the project
// qualifier, the dockerTarget name and the version.
func FormatDockerImageTag(qualifier, targetName, version string) string {
	return fmt.Sprintf("%s_%s_%s", strings.ToLower(qualifier), targetName, version)
}

// NewRegionalInstancedStack represents a stack of which multiple may exist per region.
func NewRegionalInstancedStack(app awscdk.App, region, idSuffix string) awscdk.Stack {
	qual, instance, env := QualifierFromScope(app), InstanceFromScope(app), EnvironmentFromScope(app)

	return awscdk.NewStack(app, jsii.String(qual+idSuffix+strconv.Itoa(instance)),
		&awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
				Region:  jsii.String(region),
			},
			CrossRegionReferences: jsii.Bool(true),
			Description: jsii.String(fmt.Sprintf("%s (env: %s, instance: %d, %s)",
				qual, env, instance, region)),
			Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
				Qualifier: jsii.String(strings.ToLower(qual)),
			}),
		})
}

// NewRegionalSingletonStack represents a stack of which only one exists per region but
// multiple may exist per account.
func NewRegionalSingletonStack(app awscdk.App, region, idSuffix string) awscdk.Stack {
	qual, env := QualifierFromScope(app), EnvironmentFromScope(app)

	return awscdk.NewStack(app, jsii.String(qual+idSuffix),
		&awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
				Region:  jsii.String(region),
			},
			CrossRegionReferences: jsii.Bool(true),
			Description: jsii.String(fmt.Sprintf("%s (env: %s, singleton, %s)",
				qual, env, region)),
			Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
				Qualifier: jsii.String(strings.ToLower(qual)),
			}),
		})
}

// NewSingletonStackV1 requires a "instance" context variable to allow different copies of the stack
// to exist in the same AWS account.
func NewSingletonStackV1(scope constructs.Construct, conv Conventions) awscdk.Stack {
	env := EnvironmentFromScope(scope)
	if env == "" {
		env = "<none>"
	}

	return awscdk.NewStack(scope,
		jsii.String(conv.SingletonStackName()),
		&awscdk.StackProps{
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

// NewInstancedStackV1 requires a "instance" context variable to allow different copies of the stack
// to exist in the same AWS account.
// Deprecated: use [NewInstancedStack].
func NewInstancedStackV1(scope constructs.Construct, conv Conventions) awscdk.Stack {
	instance, env := InstanceFromScope(scope), EnvironmentFromScope(scope)
	if env == "" {
		env = "<none>"
	}

	return awscdk.NewStack(scope,
		jsii.String(conv.InstancedStackName(instance)),
		&awscdk.StackProps{
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

// NewInstancedStack standardizes the creation of a stack based on three context string
// parameters: qualifier, instance and environment.
func NewInstancedStack(app awscdk.App) awscdk.Stack {
	qual, instance, env := QualifierFromScope(app), InstanceFromScope(app), EnvironmentFromScope(app)

	return awscdk.NewStack(app, jsii.String(qual+strconv.Itoa(instance)),
		&awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
				Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			},
			Description: jsii.String(fmt.Sprintf("%s (env: %s, instance: %d)",
				qual, env, instance)),
			Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
				Qualifier: jsii.String(strings.ToLower(qual)),
			}),
		})
}

// VersionFromScope retrieves the instance name from the context or an empty string.
func VersionFromScope(s constructs.Construct) string {
	v, _ := s.Node().TryGetContext(jsii.String("version")).(string)

	return v
}

// QualifierFromScope retrieves the qualifier from the context or an empty string.
func QualifierFromScope(s constructs.Construct) string {
	v, _ := s.Node().TryGetContext(jsii.String("qualifier")).(string)

	return v
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
