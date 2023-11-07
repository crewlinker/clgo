package postgresresource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// TenantProperties that configure the custom resource.
type TenantProperties struct {
	// Name of the postgres Tenant and user that will be created
	Name TenantName `mapstructure:"Name" validate:"required,resource_ident"`
	// MasterSecretArn must provide the secret for the database's master user
	MasterSecretArn string `mapstructure:"MasterSecretArn" validate:"required"`
}

// TenantName type provides ways to turn it into identifiers.
type TenantName string

// ToPhysicalResourceID encodes the properties to physical resource id.
func (n TenantName) ToPhysicalResourceID() string {
	return fmt.Sprintf("tenant-%s", n)
}

// ToDatabaseName allows consistent database naming based on the tenant.
func (n TenantName) ToDatabaseName() string {
	return fmt.Sprintf("db_tenant_%s", n)
}

// ToDatabaseUser allows consistent database naming based on the tenant.
func (n TenantName) ToDatabaseUser() string {
	return fmt.Sprintf("db_tenant_usr_%s", n)
}

// ToOutputData consitently produces the output data from the tenant name.
func (n TenantName) ToOutputData() map[string]any {
	return map[string]any{
		"DatabaseName": n.ToDatabaseName(),
		"DatabaseUser": n.ToDatabaseUser(),
	}
}

// TenantNameFromPhysicalResourceID allows us to use the physical resource id as a storage medium of
// the main property we care about: the tenant name.
func TenantNameFromPhysicalResourceID(prid string) TenantName {
	return TenantName(strings.TrimPrefix(prid, "tenant-"))
}

// handleMigrationCreates handles the migration creation. If a Create event fails, the resource provider
// framework will automatically IGNORE the subsequent Delete operation issued by AWS CloudFormation. If
// your create involves multiple distinct operations, it is your responsibility to catch and rethrow and
// clean up any partial updates that have already been performed. Make sure your API call timeouts and Lambda
// timeouts allow for this.
func (h Handler) handleTenantCreate(ctx context.Context, _ Input, props TenantProperties) (out Output, err error) {
	if err := h.runOnDB(ctx, props.MasterSecretArn, "postgres",
		func(ctx context.Context, isRDS bool, conn *pgx.Conn) error {
			if err = h.createNewTenantOrCleanup(ctx, conn, isRDS, props.Name); err != nil {
				return fmt.Errorf("failed to create new tenant: %w", err)
			}

			return nil
		}); err != nil {
		return out, fmt.Errorf("failed to execute sql on master db: %w", err)
	}

	out.PhysicalResourceID = props.Name.ToPhysicalResourceID()
	out.Data = props.Name.ToOutputData()

	return out, nil
}

// ErrMasterArnUpdateNotSupported is returned when the master arn is updated, which is not supported.
var ErrMasterArnUpdateNotSupported = errors.New("changing the MasterSecretArn for an update is not supported")

// handleTenantUpdate handles the Tenant update. If an Update event fails, CloudFormation will issue
// an additional Update with the previous properties. If you return a different PhysicalResourceId, you
// will subsequently receive a Delete event to clean up the previous state of the resource.
func (h Handler) handleTenantUpdate(
	ctx context.Context, _ Input, newProps, oldProps TenantProperties,
) (out Output, err error) {
	if newProps.MasterSecretArn != oldProps.MasterSecretArn {
		return out, ErrMasterArnUpdateNotSupported
	}

	newDB, newUsr := newProps.Name.ToDatabaseName(), newProps.Name.ToDatabaseUser()
	oldDB, oldUsr := oldProps.Name.ToDatabaseName(), oldProps.Name.ToDatabaseUser()

	if err := h.runOnDB(ctx, newProps.MasterSecretArn, "postgres", func(
		ctx context.Context, b bool, conn *pgx.Conn,
	) error {
		// unfortunately we can only rename the database if there are no connectsion. So we close them down with
		// a lot of force which can lead to data loss/inconsistency. This is the lesser of two evils for now.
		if _, err := conn.Exec(ctx,
			`SELECT pg_terminate_backend(pg_stat_activity.pid) `+
				`FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, oldDB,
		); err != nil {
			return fmt.Errorf("failed to termate connections: %w", err)
		}

		if _, err := conn.Exec(ctx, fmt.Sprintf(`ALTER DATABASE "%s" RENAME TO "%s"`, oldDB, newDB)); err != nil {
			return fmt.Errorf("failed to rename database: %w", err)
		}

		if _, err := conn.Exec(ctx, fmt.Sprintf(`ALTER USER "%s" RENAME TO "%s"`, oldUsr, newUsr)); err != nil {
			return fmt.Errorf("failed to rename user: %w", err)
		}

		return nil
	}); err != nil {
		return out, err
	}

	// return the new resource id because it is required to determine the right database name/user in the delete. This
	// will cause a DELETE to be isseus which should do nothing since we've renamed the database and user.
	out.PhysicalResourceID = newProps.Name.ToPhysicalResourceID()
	out.Data = newProps.Name.ToOutputData()

	return out, nil
}

// handleTenantDelete handles the migration delete. If a Delete event fails, CloudFormation will abandon this resource.
// The delete can also be called implicitly if the update returns a new physical resource id.
func (h Handler) handleTenantDelete(ctx context.Context, in Input, props TenantProperties) (out Output, err error) {
	if err := h.runOnDB(ctx, props.MasterSecretArn, "postgres", func(ctx context.Context, b bool, c *pgx.Conn) error {
		if err := h.dropTenantIfExists(ctx, c, props.Name); err != nil {
			return fmt.Errorf("failed to drop tenant if exist: %w", err)
		}

		return nil
	}); err != nil {
		return out, err
	}

	out.PhysicalResourceID = in.PhysicalResourceID // must always be the same, or cfn will error

	return
}

// dropTenantIfExists either.
func (h Handler) dropTenantIfExists(ctx context.Context, conn *pgx.Conn, name TenantName) error {
	h.logs.Info("drop tenant if exists", zap.String("tenant_name", string(name)))

	if err := h.dropDatabaseIfExists(ctx, conn, name); err != nil {
		return err
	}

	return h.dropUserIfExists(ctx, conn, name)
}

// runOnDB turns a arn to the master database secret into pgx connection config.
func (h Handler) runOnDB(ctx context.Context, arn, db string, runf func(context.Context, bool, *pgx.Conn) error) error {
	val, err := h.smc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: aws.String(arn)})
	if err != nil {
		return fmt.Errorf("failed to get secret value: %w", err)
	}

	var secret struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Port     int    `json:"port"`
		Host     string `json:"host"`
	}

	if err := json.Unmarshal([]byte(*val.SecretString), &secret); err != nil {
		return fmt.Errorf("failed to unmarshal secret string: %w", err)
	}

	pcfg, err := pgx.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		secret.Username,
		url.QueryEscape(secret.Password),
		net.JoinHostPort(secret.Host, fmt.Sprintf("%d", secret.Port)),
		db))
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	conn, err := pgx.ConnectConfig(ctx, pcfg)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	defer conn.Close(ctx)

	if err := runf(ctx, strings.HasSuffix(pcfg.Host, "rds.amazonaws.com"), conn); err != nil {
		return fmt.Errorf("failed to run on master database: %w", err)
	}

	return nil
}

// dropDatabaseIfExists will drop the database but won't complain if it doesn't exist.
func (h Handler) dropDatabaseIfExists(ctx context.Context, conn *pgx.Conn, tname TenantName) error {
	if _, err := conn.Exec(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" (FORCE)`, tname.ToDatabaseName())); err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	return nil
}

// dropUserIfExists drops a database user but won't error if it doesn't exist.
func (h Handler) dropUserIfExists(ctx context.Context, conn *pgx.Conn, tname TenantName) error {
	if _, err := conn.Exec(ctx, fmt.Sprintf(`DROP USER IF EXISTS "%s"`, tname.ToDatabaseUser())); err != nil {
		return fmt.Errorf("failed to drop user: %w", err)
	}

	return nil
}

// handleCleanUp will call each of 'fns' when 'err' is non-nil.
func (h Handler) handleCleanUp(
	ctx context.Context,
	err error,
	conn *pgx.Conn,
	tname TenantName,
	fns ...func(ctx context.Context, conn *pgx.Conn, tname TenantName) error,
) error {
	if err == nil {
		return nil // no error
	}

	h.logs.Info("handling cleanup",
		zap.NamedError("for_err", err), zap.String("tenant", string(tname)), zap.Int("num_clean_fns", len(fns)))

	for idx, fn := range fns {
		if err := fn(ctx, conn, tname); err != nil {
			h.logs.Error("clean up error", zap.Int("fn_idx", idx), zap.NamedError("clean_up_err", err))
		}
	}

	return err
}

// createNewTenant wil create the tenant or clean it up if there is an error.
func (h Handler) createNewTenantOrCleanup(
	ctx context.Context, conn *pgx.Conn, isRDS bool, tname TenantName,
) (err error) {
	h.logs.Info("create new tenant or clean up", zap.String("tenant_name", string(tname)))

	dbName, usrName := tname.ToDatabaseName(), tname.ToDatabaseUser()

	// create database for tenant
	if _, err := conn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, dbName)); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// create user for tenant
	if _, err = conn.Exec(ctx, fmt.Sprintf(`CREATE USER "%s"`, usrName)); err != nil {
		return h.handleCleanUp(ctx,
			fmt.Errorf("failed to create user: %w", err), conn, tname,
			h.dropDatabaseIfExists)
	}

	// if we're migrating a rds database we would like to grand iam access, so uncomment it
	rdsComment := "--"
	if isRDS {
		rdsComment = ""
	}

	// grant the user all privileges to the configured database, and rds_iam access
	if _, err = conn.Exec(ctx, fmt.Sprintf(`
		GRANT ALL PRIVILEGES ON DATABASE "%s" TO "%s"; 
		%sGRANT rds_iam TO "%s";
	`, dbName, usrName, rdsComment, usrName)); err != nil {
		return h.handleCleanUp(ctx,
			fmt.Errorf("failed to grant IAM login privileges: %w", err), conn, tname,
			h.dropDatabaseIfExists,
			h.dropUserIfExists)
	}

	// also enable the lambda extension (if this is run on aws's RDS)
	if err := h.OnOtherDB(ctx, conn.Config(), dbName, func(ctx context.Context, conn *pgx.Conn) (err error) {
		if isRDS {
			_, err = conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS aws_lambda CASCADE;`)
		}

		return err //nolint:wrapcheck
	}); err != nil {
		return h.handleCleanUp(ctx,
			fmt.Errorf("failed to install aws_lambda extension: %w", err),
			conn,
			tname,
			h.dropDatabaseIfExists,
			h.dropUserIfExists)
	}

	return nil
}

// OnOtherDB runs code with a connection to another database on the cluster.
func (h Handler) OnOtherDB(
	ctx context.Context, dbcfg *pgx.ConnConfig, dbName string, run func(ctx context.Context, conn *pgx.Conn) error,
) error {
	tcfg := dbcfg.Copy()
	tcfg.Database = dbName

	// we want to make the user owner of the public schema, but is only possible with another connection
	tconn, err := pgx.ConnectConfig(ctx, tcfg)
	if err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}

	// always close the connection
	defer func() {
		if err := tconn.Close(ctx); err != nil {
			h.logs.Error("failed to close connection to tenant database", zap.Error(err))
		}
	}()

	return run(ctx, tconn)
}
