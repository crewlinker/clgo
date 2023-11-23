package clcdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// LookupBaseZoneAndCerts implements the logic for fetching the main public hosted zone, the zone's regional wildcard
// certificate and the zone's edge certificate. These must be setup manually and provide to the stack.
func LookupBaseZoneAndCerts(scope constructs.Construct, name ScopeName, cfg Config) (
	awsroute53.IHostedZone,
	awscertificatemanager.ICertificate,
	awscertificatemanager.ICertificate,
) {
	scope = name.ChildScope(scope)

	zone := awsroute53.PublicHostedZone_FromLookup(scope, jsii.String("HostedZone"),
		&awsroute53.HostedZoneProviderProps{
			DomainName: cfg.MainDomainName(),
		})

	regional := awscertificatemanager.Certificate_FromCertificateArn(scope,
		jsii.String("RegionalCertificate"), cfg.RegionalCertificateArn())

	edge := awscertificatemanager.Certificate_FromCertificateArn(scope,
		jsii.String("EdgeCertificate"), cfg.EdgeCertificateArn())

	return zone, regional, edge
}
