package clcdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// BaseZoneAndCerts implements the logic for returing HostedZone and certificate constructs from just
// attrbutes (without lookups).
func BaseZoneAndCerts(scope constructs.Construct, name ScopeName, cfg Config) (
	awsroute53.IHostedZone,
	awscertificatemanager.ICertificate,
	awscertificatemanager.ICertificate,
) {
	scope = name.ChildScope(scope)

	zone := awsroute53.PublicHostedZone_FromPublicHostedZoneAttributes(scope, jsii.String("HostedZone"),
		&awsroute53.PublicHostedZoneAttributes{
			ZoneName:     cfg.MainDomainName(),
			HostedZoneId: cfg.MainDomainHostedZoneID(),
		})

	regional := awscertificatemanager.Certificate_FromCertificateArn(scope,
		jsii.String("RegionalCertificate"), cfg.RegionalCertificateArn())

	edge := awscertificatemanager.Certificate_FromCertificateArn(scope,
		jsii.String("EdgeCertificate"), cfg.EdgeCertificateArn())

	return zone, regional, edge
}
