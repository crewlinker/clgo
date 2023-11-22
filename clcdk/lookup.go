package clcdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type baseInputs struct {
	mainDomainName  string
	regionalCertArn string
	edgeCertArn     string
}

// NewBaseInputs creates a base input from its resource references.
func NewBaseInputs(
	mainDomainName string,
	regionalCertArn string,
	edgeCertArn string,
) BaseInputs {
	return baseInputs{
		mainDomainName:  mainDomainName,
		regionalCertArn: regionalCertArn,
		edgeCertArn:     edgeCertArn,
	}
}

func (bi baseInputs) MainDomainName() string  { return bi.mainDomainName }
func (bi baseInputs) RegionalCertArn() string { return bi.regionalCertArn }
func (bi baseInputs) EdgeCertArn() string     { return bi.edgeCertArn }

// BaseInputs are references to resources that are expected to exist in the
// AWS account before the stack is created.
type BaseInputs interface {
	MainDomainName() string
	RegionalCertArn() string
	EdgeCertArn() string
}

// LookupBaseZoneAndCerts implements the logic for fetching the main public hosted zone, the zone's regional wildcard
// certificate and the zone's edge certificate. These must be setup manually and provide to the stack.
func LookupBaseZoneAndCerts(scope constructs.Construct, name ScopeName, inp BaseInputs) (
	awsroute53.IHostedZone,
	awscertificatemanager.ICertificate,
	awscertificatemanager.ICertificate,
) {
	scope = name.ChildScope(scope)

	zone := awsroute53.PublicHostedZone_FromLookup(scope, jsii.String("HostedZone"),
		&awsroute53.HostedZoneProviderProps{
			DomainName: jsii.String(inp.MainDomainName()),
		})

	regional := awscertificatemanager.Certificate_FromCertificateArn(scope,
		jsii.String("RegionalCertificate"), jsii.String(inp.RegionalCertArn()))

	edge := awscertificatemanager.Certificate_FromCertificateArn(scope,
		jsii.String("EdgeCertificate"), jsii.String(inp.EdgeCertArn()))

	return zone, regional, edge
}
