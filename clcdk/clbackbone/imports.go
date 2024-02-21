// Package clbackbone provdes constructs for interacting with the backbone platform infrastructure.
package clbackbone

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// imports implementation.
type imports struct {
	vpc                  awsec2.IVpc
	cluster              awsecs.ICluster
	hostedZone           awsroute53.IHostedZone
	mainRepository       awsecr.IRepository
	loadBalancerListener awselasticloadbalancingv2.IApplicationListener

	postgresReadWriteHostname           *string
	postgresReadOnlyHostname            *string
	t3aSmallCapacityProviderName        *string
	postgresCustomResourceProviderToken *string
}

const (
	nrOfZones = 2
	httpsPort = 443
)

// Imports construct.
type Imports interface {
	VPC() awsec2.IVpc
	ContainerCluster() awsecs.ICluster
	LoadBalancerListener() awselasticloadbalancingv2.IApplicationListener
	HostedZone() awsroute53.IHostedZone

	DBROHostName() *string
	DBRWHostName() *string

	DBCustomProviderToken() *string
	T3aSmallCapacityProviderName() *string
}

// NewMainRegionImport inits the imports construct.
func NewMainRegionImport(scope constructs.Construct, importPrefix string) Imports {
	con, stack := imports{}, awscdk.Stack_Of(scope)

	con.vpc = awsec2.Vpc_FromVpcAttributes(scope, jsii.String("Vpc"), &awsec2.VpcAttributes{
		VpcId: awscdk.Fn_ImportValue(jsii.String(importPrefix + ":VpcId")),
		AvailabilityZones: awscdk.Fn_Split(jsii.String(","),
			awscdk.Fn_ImportValue(jsii.String(importPrefix+":AvailabilityZones")), jsii.Number(nrOfZones)),
		PublicSubnetIds: awscdk.Fn_Split(jsii.String(","),
			awscdk.Fn_ImportValue(jsii.String(importPrefix+":PublicSubnetIds")), jsii.Number(nrOfZones)),
		PublicSubnetRouteTableIds: awscdk.Fn_Split(jsii.String(","),
			awscdk.Fn_ImportValue(jsii.String(importPrefix+":PublicSubnetRtIds")), jsii.Number(nrOfZones)),
	})

	con.cluster = awsecs.Cluster_FromClusterAttributes(scope, jsii.String("Cluster"), &awsecs.ClusterAttributes{
		ClusterName:    awscdk.Fn_ImportValue(jsii.String(importPrefix + ":ClusterName")),
		Vpc:            con.vpc,
		HasEc2Capacity: jsii.Bool(true),
	})

	con.hostedZone = awsroute53.HostedZone_FromHostedZoneAttributes(scope, jsii.String("HostedZone"),
		&awsroute53.HostedZoneAttributes{
			HostedZoneId: awscdk.Fn_ImportValue(jsii.String(importPrefix + ":HostedZoneId")),
			ZoneName:     awscdk.Fn_ImportValue(jsii.String(importPrefix + ":HostedZoneName")),
		})

	con.mainRepository = awsecr.Repository_FromRepositoryAttributes(scope, jsii.String("MainRepository"),
		&awsecr.RepositoryAttributes{
			RepositoryArn: jsii.Sprintf("arn:aws:ecr:%s:%s:repository/%s",
				*stack.Region(),
				*stack.Account(),
				*awscdk.Fn_ImportValue(jsii.String(importPrefix + ":ContainerRegistryMainRepositoryName"))),
			RepositoryName: awscdk.Fn_ImportValue(jsii.String(importPrefix + ":ContainerRegistryMainRepositoryName")),
		})

	con.postgresReadOnlyHostname = awscdk.Fn_ImportValue(jsii.String(importPrefix + ":PostgresReadOnlyHostname"))
	con.postgresReadWriteHostname = awscdk.Fn_ImportValue(jsii.String(importPrefix + ":PostgresReadWriteHostname"))

	con.t3aSmallCapacityProviderName = awscdk.Fn_ImportValue(
		jsii.String(importPrefix + ":CapacityT3aSmall:CapacityProviderName"))
	con.postgresCustomResourceProviderToken = awscdk.Fn_ImportValue(
		jsii.String(importPrefix + ":PostgresCustomResourceProviderToken"))

	con.loadBalancerListener = awselasticloadbalancingv2.ApplicationListener_FromApplicationListenerAttributes(scope,
		jsii.String("LoadBalancerListener"), &awselasticloadbalancingv2.ApplicationListenerAttributes{
			ListenerArn: awscdk.Fn_ImportValue(jsii.String(importPrefix + ":LoadBalancerHttpsListenerArn")),
			SecurityGroup: awsec2.SecurityGroup_FromSecurityGroupId(scope, jsii.String("AlbSg"),
				awscdk.Fn_ImportValue(jsii.String(importPrefix+":LoadBalancerSgId")), nil),
			DefaultPort: jsii.Number(httpsPort),
		})

	return con
}

func (c imports) VPC() awsec2.IVpc { return c.vpc }

func (c imports) LoadBalancerListener() awselasticloadbalancingv2.IApplicationListener {
	return c.loadBalancerListener
}

func (c imports) HostedZone() awsroute53.IHostedZone    { return c.hostedZone }
func (c imports) ContainerCluster() awsecs.ICluster     { return c.cluster }
func (c imports) DBROHostName() *string                 { return c.postgresReadOnlyHostname }
func (c imports) DBRWHostName() *string                 { return c.postgresReadWriteHostname }
func (c imports) DBCustomProviderToken() *string        { return c.postgresCustomResourceProviderToken }
func (c imports) T3aSmallCapacityProviderName() *string { return c.t3aSmallCapacityProviderName }
