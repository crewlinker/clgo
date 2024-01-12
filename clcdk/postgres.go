package clcdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// PostgresTenant provides an interface to retrieve information
// on a unique tenant in a postgres database.
type PostgresTenant interface {
	DatabaseName() *string
	DatabaseUser() *string
}

// implements the postgres tenant resource.
type postgresTenant struct {
	resource awscdk.CustomResource
}

func (pt postgresTenant) DatabaseName() *string {
	return pt.resource.GetAttString(jsii.String("DatabaseName"))
}

func (pt postgresTenant) DatabaseUser() *string {
	return pt.resource.GetAttString(jsii.String("DatabaseUser"))
}

// WithPostgresTenant creates a tenant with.
func WithPostgresTenant(
	scope constructs.Construct,
	name ScopeName,
	providerToken *string,
	dbSecret awssecretsmanager.ISecret,
	tenantName string,
) PostgresTenant {
	scope = name.ChildScope(scope)

	tenant := awscdk.NewCustomResource(scope,
		jsii.String("Tenant"), &awscdk.CustomResourceProps{
			ServiceToken: providerToken,
			ResourceType: jsii.String("Custom::CrewlinkerPostgresTenant"),
			Properties: &map[string]any{
				"Name":            tenantName,
				"MasterSecretArn": *dbSecret.SecretFullArn(),
			},
		})

	return postgresTenant{resource: tenant}
}

// withPostgresInstance will ensure an AWS RDS postgres instance is either created or imported.
func WithPostgresInstance(
	scope constructs.Construct, name ScopeName, cfg Config, vpc awsec2.IVpc,
	allocatedStorageGib, maxAllocatedStorageGib float64,
) (awsrds.IDatabaseInstance, awssecretsmanager.ISecret) {
	scope, stack := name.ChildScope(scope), awscdk.Stack_Of(scope)

	const (
		// constants for setting up the instance
		monitoringIntervalSeconds = 15
		backupRetentionDays       = 7
		port                      = 5432
	)

	// setup the database secret, it is commonly passed around to support connections in various places.
	secret := awsrds.NewDatabaseSecret(scope, jsii.String("Secret"), &awsrds.DatabaseSecretProps{
		SecretName: jsii.String(*stack.StackName() + string(name) + "Secret"),
		Username:   jsii.String("postgres"),
	})

	// security group for public access
	securityGroup := awsec2.NewSecurityGroup(scope, jsii.String("SecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
	})

	securityGroup.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(port)),
		jsii.String("allow all inbound access to postgres"), jsii.Bool(false))

	// engine for parameters and clusters
	engine := awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
		Version: awsrds.PostgresEngineVersion_VER_15_3(),
	})

	// configuration as a parameter group
	// configure the instance using a parameter group
	parameters := awsrds.NewParameterGroup(scope, jsii.String("ParameterGroup"), &awsrds.ParameterGroupProps{
		Engine: engine,

		Parameters: &map[string]*string{
			// allow logical replication for Debezium-like connectors
			"rds.force_ssl":           jsii.String("1"),
			"rds.logical_replication": jsii.String("1"),
		},
	})

	// setup the instance
	instance := awsrds.NewDatabaseInstance(scope, jsii.String("Instance"), &awsrds.DatabaseInstanceProps{
		RemovalPolicy:      cfg.RemovalPolicyIfSnapshotable(),
		DeletionProtection: cfg.DeletionProtection(),

		Engine:              engine,
		InstanceType:        awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE4_GRAVITON, awsec2.InstanceSize_MICRO),
		Vpc:                 vpc,
		AllocatedStorage:    jsii.Number(allocatedStorageGib),    // GiB
		MaxAllocatedStorage: jsii.Number(maxAllocatedStorageGib), // GiB

		// We use a reference to a secret we create ourselves so we can easily look it up in other stacks (byname)
		Credentials: awsrds.Credentials_FromSecret(secret, jsii.String("postgres")),

		// Each instance will have performance insight enabled and is publicly accessible
		IamAuthentication:           jsii.Bool(true),
		AutoMinorVersionUpgrade:     jsii.Bool(true),
		VpcSubnets:                  &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC},
		PubliclyAccessible:          jsii.Bool(true),
		SecurityGroups:              securityGroup.Connections().SecurityGroups(),
		EnablePerformanceInsights:   jsii.Bool(true),
		PerformanceInsightRetention: awsrds.PerformanceInsightRetention_DEFAULT,

		// enables enhanced monitoring
		MonitoringInterval: awscdk.Duration_Seconds(jsii.Number(monitoringIntervalSeconds)),
		// update to higher-security RSA certifiacte that doesn't expire in 2024
		CaCertificate: awsrds.CaCertificate_RDS_CA_RDS4096_G1(),

		// We export postgres logs to cloudwatch so we can add alarms if we want to.
		CloudwatchLogsExports:   jsii.Strings("postgresql"),
		CloudwatchLogsRetention: cfg.LogRetention(),
		// backups for disaster recovery
		BackupRetention: awscdk.Duration_Days(jsii.Number(backupRetentionDays)),
		// we only allow tls connections since the password will travel over the public internet
		ParameterGroup: parameters,
	})

	return instance, secret
}
