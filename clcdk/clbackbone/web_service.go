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
type WebService interface{}

// NewWebService creates a web service construct.
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
	hostHeaderCondition string,
	pathPatternCondition string,
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
		})

	con.service = awsecs.NewEc2Service(scope, jsii.String("Service"), &awsecs.Ec2ServiceProps{
		Cluster:           cluster,
		TaskDefinition:    con.definition,
		DesiredCount:      jsii.Number(desiredCount),
		MinHealthyPercent: jsii.Number(minHealthPercent),
		CapacityProviderStrategies: &[]*awsecs.CapacityProviderStrategy{
			{
				CapacityProvider: capacityProviderName,
				Weight:           jsii.Number(1),
			},
		},
	})

	con.targetGroup = awselbv2.NewApplicationTargetGroup(scope, jsii.String("TargetGroup"),
		&awselbv2.ApplicationTargetGroupProps{
			Vpc:      vpc,
			Protocol: awselbv2.ApplicationProtocol_HTTP,
			HealthCheck: &awselbv2.HealthCheck{
				Path: jsii.String(healthCheckPath),
			},
			Targets: &[]awselbv2.IApplicationLoadBalancerTarget{
				con.service.LoadBalancerTarget(&awsecs.LoadBalancerTargetOptions{
					ContainerName: jsii.String("Main"),
					ContainerPort: jsii.Number(containerPort),
				}),
			},
		})

	conditions := []awselbv2.ListenerCondition{}

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

	con.listenerRule = awselbv2.NewApplicationListenerRule(scope, jsii.String("ListenerRule"),
		&awselbv2.ApplicationListenerRuleProps{
			Listener:   loadBalancerListener,
			Priority:   jsii.Number(listenerPriority),
			Conditions: &conditions,
			Action: awselbv2.ListenerAction_Forward(&[]awselbv2.IApplicationTargetGroup{
				con.targetGroup,
			}, nil),
		})

	return con
}
