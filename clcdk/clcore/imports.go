package clcore

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type imports struct {
	vpc        awsec2.IVpc
	cluster    awsecs.ICluster
	hostedZone awsroute53.IHostedZone
	repository awsecr.IRepository
}

// Imports describe resources imported from the "core" stack.
type Imports interface {
	VPC() awsec2.IVpc
	Cluster() awsecs.ICluster
	HostedZone() awsroute53.IHostedZone
	Repository() awsecr.IRepository
}

const (
	nrOfZones = 2
)

// NewImports inits imported resources from the "Core".
func NewImports(scope constructs.Construct, importPrefix string) Imports {
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

	con.repository = awsecr.Repository_FromRepositoryAttributes(scope, jsii.String("Repository"),
		&awsecr.RepositoryAttributes{
			RepositoryArn: jsii.Sprintf("arn:aws:ecr:%s:%s:repository/%s",
				*stack.Region(),
				*stack.Account(),
				*awscdk.Fn_ImportValue(jsii.String(importPrefix + ":ContainerRepositoryName"))),
			RepositoryName: awscdk.Fn_ImportValue(jsii.String(importPrefix + ":ContainerRepositoryName")),
		})

	return con
}

func (con imports) Cluster() awsecs.ICluster           { return con.cluster }
func (con imports) VPC() awsec2.IVpc                   { return con.vpc }
func (con imports) HostedZone() awsroute53.IHostedZone { return con.hostedZone }
func (con imports) Repository() awsecr.IRepository     { return con.repository }

// ImportCapacityProviderName will use the importPrefix and instanceID to import its capacity provider name.
func ImportCapacityProviderName(importPrefix, instanceID string) *string {
	return awscdk.Fn_ImportValue(jsii.String(importPrefix + ":" + instanceID + ":CapacityProviderName"))
}

type instanceImports struct {
	capacityProviderName *string
	securityGroup        awsec2.ISecurityGroup
}

// InstanceImports are imports for deploying something on an instance via an
// ECS cluster with a capacity provider.
type InstanceImports interface {
	CapacityProviderName() *string
	SecurityGroup() awsec2.ISecurityGroup
}

// NewInstanceImports inits imports for a capacity instance.
func NewInstanceImports(scope constructs.Construct, importPrefix, instanceID string) InstanceImports {
	scope, con := constructs.NewConstruct(scope, jsii.String("InstanceImport"+instanceID)), instanceImports{}

	con.capacityProviderName = awscdk.Fn_ImportValue(
		jsii.String(importPrefix + ":" + instanceID + ":CapacityProviderName"))
	con.securityGroup = awsec2.SecurityGroup_FromSecurityGroupId(
		scope,
		jsii.String("SecurityGroup"),
		awscdk.Fn_ImportValue(jsii.String(importPrefix+":"+instanceID+":SecurityGroupId")), nil)

	return con
}

func (con instanceImports) CapacityProviderName() *string {
	return con.capacityProviderName
}

func (con instanceImports) SecurityGroup() awsec2.ISecurityGroup {
	return con.securityGroup
}
