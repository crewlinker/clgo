package clcore

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"
)

// postgresTenant private.
type postgresTenant struct {
	database         awscdk.CustomResource
	user             awscdk.CustomResource
	password, secret awssecretsmanager.ISecret
}

// PostgresTenant construct.
type PostgresTenant interface {
	Secret() awssecretsmanager.ISecret
	SecretFromReplicated(scope constructs.Construct) awssecretsmanager.ISecret
}

// Secret itself for the local region.
func (con postgresTenant) Secret() awssecretsmanager.ISecret {
	return con.secret
}

// SecretFromReplicated returns a secret that is configured to be replicated.
func (con postgresTenant) SecretFromReplicated(scope constructs.Construct) awssecretsmanager.ISecret {
	return awssecretsmanager.Secret_FromSecretNameV2(scope,
		jsii.String("DbSecretFromCrossAcountRef"), con.secret.SecretName())
}

// NewPostgresTenant makes sure that the tenant.
func NewPostgresTenant(
	scope constructs.Construct,
	idSuffix string,
	providerToken *string,
	replicaRegions []string,
) PostgresTenant {
	con, scope, stack := postgresTenant{},
		constructs.NewConstruct(scope, jsii.String("PostgresTenant"+idSuffix)),
		awscdk.Stack_Of(scope)
	qual, instance := clcdk.QualifierFromScope(scope), clcdk.InstanceFromScope(scope)

	const passwordLength = 40

	// generate a password for our tenant user/database
	con.password = awssecretsmanager.NewSecret(scope, jsii.String("Password"), &awssecretsmanager.SecretProps{
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			ExcludeCharacters:  jsii.String("/@\""),
			ExcludePunctuation: jsii.Bool(true),
			PasswordLength:     jsii.Number(passwordLength),
		},
	})

	// use our custom resource to create an isolated database on the rds instance
	con.database = awscdk.NewCustomResource(scope,
		jsii.String("Database"), &awscdk.CustomResourceProps{
			ServiceToken: providerToken,
			ResourceType: jsii.String("Custom::CrewlinkerPostgresDatabase"),
			Properties: &map[string]any{
				"Name": fmt.Sprintf("%s%ddb", strings.ToLower(qual), instance),
			},
		})

	// use our custom resource to create a dedicated user for this instance/project.
	con.user = awscdk.NewCustomResource(scope,
		jsii.String("User"), &awscdk.CustomResourceProps{
			ServiceToken: providerToken,
			ResourceType: jsii.String("Custom::CrewlinkerPostgresUser"),
			Properties: &map[string]any{
				"SecretArn": con.password.SecretFullArn(),
				"Name":      fmt.Sprintf("%s%d", strings.ToLower(qual), instance),
				"Database":  con.database.GetAttString(jsii.String("DatabaseName")),
				"Role":      con.database.GetAttString(jsii.String("ReadWriteRoleName")),
			},
		})

	// store the created credential information in a secret that is replicated across regions.
	smgrReplicaRegions := []*awssecretsmanager.ReplicaRegion{}
	for _, r := range replicaRegions {
		smgrReplicaRegions = append(smgrReplicaRegions, &awssecretsmanager.ReplicaRegion{
			Region: jsii.String(r),
		})
	}

	con.secret = awssecretsmanager.NewSecret(scope, jsii.String("Secret"), &awssecretsmanager.SecretProps{
		SecretName:     jsii.Sprintf("%s%dPostgresTenantSecret", qual, instance),
		ReplicaRegions: &smgrReplicaRegions,
		SecretObjectValue: &map[string]awscdk.SecretValue{
			"database": awscdk.SecretValue_UnsafePlainText(con.database.GetAttString(jsii.String("DatabaseName"))),
			"user":     awscdk.SecretValue_UnsafePlainText(con.user.GetAttString(jsii.String("RoleName"))),
			"password": con.password.SecretValue(), // this will not update if the secret value is changed/rotated
		},
	})

	// export the tenant info so migration scripts can read and use it.
	awscdk.Stack_Of(scope).ExportValue(con.database.GetAttString(jsii.String("DatabaseName")),
		&awscdk.ExportValueOptions{
			Name: jsii.String(*stack.StackName() + ":PostgresDatabaseName" + idSuffix),
		})

	awscdk.Stack_Of(scope).ExportValue(con.user.GetAttString(jsii.String("RoleName")),
		&awscdk.ExportValueOptions{
			Name: jsii.String(*stack.StackName() + ":PostgresDatabaseUser" + idSuffix),
		})

	awscdk.Stack_Of(scope).ExportValue(con.password.SecretName(),
		&awscdk.ExportValueOptions{
			Name: jsii.String(*stack.StackName() + ":PostgresDatabasePasswordSecretName" + idSuffix),
		})

	return con
}
