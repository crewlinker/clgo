package clbackbone

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	awselbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"
)

// webService implementation.
type webService struct {
	logs          awslogs.ILogGroup
	definition    awsecs.TaskDefinition
	mainContainer awsecs.ContainerDefinition
	service       awsecs.Ec2Service
	targetGroup   awselbv2.ApplicationTargetGroup
	listenerRule  awselbv2.ApplicationListenerRule
}

// WebService construct.
type WebService interface {
	TargetGroup() awselbv2.IApplicationTargetGroup
	TaskDefinition() awsecs.ITaskDefinition
}

// NewWebService creates a web service construct.
//
//nolint:funlen
func NewWebService(
	scope constructs.Construct,
	idSuffix string,
	vpc awsec2.IVpc,
	cluster awsecs.ICluster,
	loadBalancerListener awselbv2.IApplicationListener,
	image awsecs.ContainerImage,
	capacityProviderName *string,
	memoryReservationMiB int,
	listenerPriority int,
	containerPort int,
	healthCheckPath string,
	desiredCount int,
	minHealthPercent int,
	maxHealthPercent int,
	hostHeaderCondition string,
	pathPatternCondition string,
	environment *map[string]*string,
	secrets *map[string]awsecs.Secret,
	healthCheckCommand []string,
	authenticateOidcOptions *awselbv2.AuthenticateOidcOptions,
) WebService {
	con, scope := webService{}, constructs.NewConstruct(scope, jsii.String("WebService"+idSuffix))
	qual, instance := clcdk.QualifierFromScope(scope), clcdk.InstanceFromScope(scope)
	serviceName := strings.ToLower(fmt.Sprintf("%s%d%s", qual, instance, idSuffix))

	con.logs = awslogs.NewLogGroup(scope, jsii.String("Logs"), &awslogs.LogGroupProps{
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Retention:     awslogs.RetentionDays_FIVE_DAYS,
	})

	con.definition = awsecs.NewEc2TaskDefinition(scope, jsii.String("Definition"),
		&awsecs.Ec2TaskDefinitionProps{})

	const (
		containerHealthCheckIntervalSecs    = 30
		containerHealthCheckTimeoutSecs     = 6
		containerHealthCheckRetries         = 3
		containerHealthCheckStartPeriodSecs = 1
	)

	var containerHealthCheck *awsecs.HealthCheck
	if len(healthCheckCommand) > 0 {
		containerHealthCheck = &awsecs.HealthCheck{
			Command:     jsii.Strings(healthCheckCommand...),
			Interval:    awscdk.Duration_Seconds(jsii.Number(containerHealthCheckIntervalSecs)),
			Timeout:     awscdk.Duration_Seconds(jsii.Number(containerHealthCheckTimeoutSecs)),
			Retries:     jsii.Number(containerHealthCheckRetries),
			StartPeriod: awscdk.Duration_Seconds(jsii.Number(containerHealthCheckStartPeriodSecs)),
		}
	}

	con.mainContainer = con.definition.AddContainer(jsii.String("Main"),
		&awsecs.ContainerDefinitionOptions{
			Image:                image,
			MemoryReservationMiB: jsii.Number(memoryReservationMiB),
			PortMappings: &[]*awsecs.PortMapping{
				{ContainerPort: jsii.Number(containerPort)},
			},
			Logging: awsecs.AwsLogDriver_AwsLogs(&awsecs.AwsLogDriverProps{
				StreamPrefix: jsii.String(serviceName), // e.g: clatsback1rpc
				LogGroup:     con.logs,
			}),
			Environment: environment,
			Secrets:     secrets,
			HealthCheck: containerHealthCheck,
		})

	con.service = awsecs.NewEc2Service(scope, jsii.String("Service"), &awsecs.Ec2ServiceProps{
		Cluster:           cluster,
		TaskDefinition:    con.definition,
		DesiredCount:      jsii.Number(desiredCount),
		MinHealthyPercent: jsii.Number(minHealthPercent),
		MaxHealthyPercent: jsii.Number(maxHealthPercent),
		CapacityProviderStrategies: &[]*awsecs.CapacityProviderStrategy{
			{
				CapacityProvider: capacityProviderName,
				Weight:           jsii.Number(1),
			},
		},
	})

	const (
		healthCheckIntervalSec    = 5
		healthCheckTimeoutSec     = 2
		healthCheckThresholdCount = 2
	)

	con.targetGroup = awselbv2.NewApplicationTargetGroup(scope, jsii.String("TargetGroup"),
		&awselbv2.ApplicationTargetGroupProps{
			Vpc:      vpc,
			Protocol: awselbv2.ApplicationProtocol_HTTP,
			HealthCheck: &awselbv2.HealthCheck{
				Path: jsii.String(healthCheckPath),

				// "You can speed up the health-check process if your service starts up and stabilizes in under 10
				// seconds. To speed up the process, reduce the number of checks and the interval between the checks."
				// https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/load-balancer-healthcheck.html
				Interval:                awscdk.Duration_Seconds(jsii.Number(healthCheckIntervalSec)),
				Timeout:                 awscdk.Duration_Seconds(jsii.Number(healthCheckTimeoutSec)),
				HealthyThresholdCount:   jsii.Number(healthCheckThresholdCount),
				UnhealthyThresholdCount: jsii.Number(healthCheckThresholdCount),
			},
			Targets: &[]awselbv2.IApplicationLoadBalancerTarget{
				con.service.LoadBalancerTarget(&awsecs.LoadBalancerTargetOptions{
					ContainerName: jsii.String("Main"),
					ContainerPort: jsii.Number(containerPort),
				}),
			},
		})

	conditions := []awselbv2.ListenerCondition{}

	// quicker deregistration, see:
	// https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/load-balancer-connection-draining.html
	con.targetGroup.SetAttribute(jsii.String("deregistration_delay.timeout_seconds"), jsii.String("5"))

	if hostHeaderCondition != "" {
		conditions = append(conditions, awselbv2.ListenerCondition_HostHeaders(
			jsii.Strings(hostHeaderCondition),
		))
	}

	if pathPatternCondition != "" {
		conditions = append(conditions, awselbv2.ListenerCondition_PathPatterns(
			jsii.Strings(pathPatternCondition),
		))
	}

	// by default, the action is just to use the target group
	action := awselbv2.ListenerAction_Forward(&[]awselbv2.IApplicationTargetGroup{
		con.targetGroup,
	}, nil)

	// if authentication is set, the default action is wrapped
	if authenticateOidcOptions != nil {
		authenticateOidcOptions.Next = action

		action = awselbv2.ListenerAction_AuthenticateOidc(authenticateOidcOptions)
	}

	con.listenerRule = awselbv2.NewApplicationListenerRule(scope, jsii.String("ListenerRule"),
		&awselbv2.ApplicationListenerRuleProps{
			Listener:   loadBalancerListener,
			Priority:   jsii.Number(listenerPriority),
			Conditions: &conditions,
			Action:     action,
		})

	return con
}

func (con webService) TargetGroup() awselbv2.IApplicationTargetGroup { return con.targetGroup }
func (con webService) TaskDefinition() awsecs.ITaskDefinition        { return con.definition }
