package clplatform

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	awselbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Imports construct provides access to imported resources from the
// platform stack.
//
//nolint:interfacebloat
type Imports interface {
	VPC() awsec2.IVpc
	LoadBalancerDNSName() *string
	ContainerCluster() awsecs.ICluster
	LoadBalancerListener() awselbv2.IApplicationListener
	WildcardCertificate() awscertificatemanager.ICertificate
	HostedZone() awsroute53.IHostedZone
	EnvSecret() awssecretsmanager.ISecret
	DBROHostName() *string
	DBRWHostName() *string
	DBResourceID() *string
	DBCustomProviderToken() *string
	StdSmallCapacityName() *string
}

const (
	nrOfZones    = 2
	httpsPort    = 443
	importPrefix = "ClPlatform:V2"
)

// NewImports inits the root constructs for the stack.
func NewImports(scope constructs.Construct) Imports {
	con := imports{}

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

	con.albDNSName = awscdk.Fn_ImportValue(jsii.String(importPrefix + ":LoadBalancerDnsName"))
	con.listener = awselbv2.ApplicationListener_FromApplicationListenerAttributes(scope,
		jsii.String("Listener"), &awselbv2.ApplicationListenerAttributes{
			ListenerArn: awscdk.Fn_ImportValue(jsii.String(importPrefix + ":LoadBalancerHttpsListenerArn")),
			SecurityGroup: awsec2.SecurityGroup_FromSecurityGroupId(scope, jsii.String("AlbSg"),
				awscdk.Fn_ImportValue(jsii.String(importPrefix+":LoadBalancerSgId")), nil),
			DefaultPort: jsii.Number(httpsPort),
		})

	con.cert = awscertificatemanager.Certificate_FromCertificateArn(scope, jsii.String("Cert"),
		awscdk.Fn_ImportValue(jsii.String(importPrefix+":WildcardCertificateArn")))

	con.zone = awsroute53.HostedZone_FromHostedZoneAttributes(scope, jsii.String("Zone"),
		&awsroute53.HostedZoneAttributes{
			HostedZoneId: awscdk.Fn_ImportValue(jsii.String(importPrefix + ":HostedZoneId")),
			ZoneName:     awscdk.Fn_ImportValue(jsii.String(importPrefix + ":HostedZoneName")),
		})

	con.dbROHostname = awscdk.Fn_ImportValue(jsii.String(importPrefix + ":PostgresClusterReadOnlyHostname"))
	con.dbRWHostname = awscdk.Fn_ImportValue(jsii.String(importPrefix + ":PostgresClusterReadWriteHostname"))
	con.dbResourceID = awscdk.Fn_ImportValue(jsii.String(importPrefix + ":PostgresClusterResourceId"))

	con.dbResourceProviderToken = awscdk.Fn_ImportValue(
		jsii.String(importPrefix + ":PostgresV3ResourceProviderServiceToken"))

	con.envSecret = awssecretsmanager.Secret_FromSecretCompleteArn(scope, jsii.String("EnvSecret"),
		awscdk.Fn_ImportValue(jsii.String(importPrefix+":EnvSecretFullArn")))

	con.stdSmallCapcityProviderName = awscdk.Fn_ImportValue(jsii.String(importPrefix + ":SmallStdCapacityProviderName"))

	return con
}

// imports construct.
type imports struct {
	vpc                         awsec2.IVpc
	cluster                     awsecs.ICluster
	albDNSName                  *string
	listener                    awselbv2.IApplicationListener
	cert                        awscertificatemanager.ICertificate
	zone                        awsroute53.IHostedZone
	dbROHostname                *string
	dbRWHostname                *string
	dbResourceID                *string
	dbResourceProviderToken     *string
	envSecret                   awssecretsmanager.ISecret
	stdSmallCapcityProviderName *string
}

func (c imports) VPC() awsec2.IVpc                                        { return c.vpc }
func (c imports) LoadBalancerDNSName() *string                            { return c.albDNSName }
func (c imports) LoadBalancerListener() awselbv2.IApplicationListener     { return c.listener }
func (c imports) WildcardCertificate() awscertificatemanager.ICertificate { return c.cert }
func (c imports) HostedZone() awsroute53.IHostedZone                      { return c.zone }
func (c imports) ContainerCluster() awsecs.ICluster                       { return c.cluster }
func (c imports) EnvSecret() awssecretsmanager.ISecret                    { return c.envSecret }
func (c imports) DBROHostName() *string                                   { return c.dbROHostname }
func (c imports) DBRWHostName() *string                                   { return c.dbRWHostname }
func (c imports) DBResourceID() *string                                   { return c.dbResourceID }
func (c imports) DBCustomProviderToken() *string                          { return c.dbResourceProviderToken }
func (c imports) StdSmallCapacityName() *string                           { return c.stdSmallCapcityProviderName }
