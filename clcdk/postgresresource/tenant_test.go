package postgresresource_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/crewlinker/clgo/clcdk/postgresresource"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("tenant identifiers", func() {
	It("from tenant name", func() {
		Expect(postgresresource.TenantName("foo").ToPhysicalResourceID()).To(Equal("tenant-foo"))
	})

	It("to tenant name", func() {
		Expect(postgresresource.TenantNameFromPhysicalResourceID("tenant-foo")).To(Equal(postgresresource.TenantName("foo")))
	})

	It("to database name", func() {
		Expect(postgresresource.TenantName("foo1").ToDatabaseName()).To(Equal("db_tenant_foo1"))
	})

	It("to database user", func() {
		Expect(postgresresource.TenantName("foo1").ToDatabaseUser()).To(Equal("db_tenant_usr_foo1"))
	})
})

var _ = Describe("tenant", func() {
	var obs *observer.ObservedLogs
	var hdl *postgresresource.Handler
	var smc *secretsmanager.Client
	var rw *pgxpool.Pool
	var dbcfg *pgxpool.Config

	var name, updName postgresresource.TenantName

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&hdl, &obs, &smc, &rw, &dbcfg),
			clpostgres.Test(),
			postgresresource.Test())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)

		b := make([]byte, 6)
		Expect(rand.Read(b)).To(Equal(6))

		name, updName = postgresresource.TenantName(fmt.Sprintf("ten_%x", b)),
			postgresresource.TenantName(fmt.Sprintf("upd_ten_%x", b))

		_ = updName
	})

	AfterEach(func(ctx context.Context) {
		Expect(rw.Exec(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" (FORCE)`, name.ToDatabaseName()))).ToNot(BeNil())
		Expect(rw.Exec(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" (FORCE)`, updName.ToDatabaseName()))).ToNot(BeNil())

		Expect(rw.Exec(ctx, fmt.Sprintf(`DROP USER IF EXISTS "%s"`, name.ToDatabaseUser()))).ToNot(BeNil())
		Expect(rw.Exec(ctx, fmt.Sprintf(`DROP USER IF EXISTS "%s"`, updName.ToDatabaseUser()))).ToNot(BeNil())

		ExpectDatabaseCountWithName(ctx, rw, updName.ToDatabaseName(), Equal(0))
		ExpectUserCountWithName(ctx, rw, updName.ToDatabaseUser(), Equal(0))
		ExpectDatabaseCountWithName(ctx, rw, name.ToDatabaseName(), Equal(0))
		ExpectUserCountWithName(ctx, rw, name.ToDatabaseUser(), Equal(0))
	})

	DescribeTable("validation", func(ctx context.Context, in postgresresource.Input, expErr, expResp OmegaMatcher) {
		resp, err := hdl.Handle(ctx, in)
		Expect(err).To(expErr)
		Expect(resp).To(expResp)
	},
		Entry(`unsupported resource`, postgresresource.Input{},
			MatchError(MatchRegexp(`unsupported resource`)), Equal(postgresresource.Output{})),

		Entry(`validate required`, postgresresource.Input{
			ResourceType: "Custom::CrewlinkerPostgresTenant",
		}, MatchError(MatchRegexp(`'Name' failed on the 'required' tag`)), Equal(postgresresource.Output{})),

		Entry(`validate identifier`, postgresresource.Input{
			ResourceType: "Custom::CrewlinkerPostgresTenant",
			ResourceProperties: map[string]interface{}{
				"Name":            "!asf",
				"MasterSecretArn": "arn:aws:secretsmanager:eu-central-1:860345245734:secret:bogus-asdf",
			},
		}, MatchError(MatchRegexp(`resource_ident`)), Equal(postgresresource.Output{})),

		Entry(`invalid request type`, postgresresource.Input{
			ResourceType: "Custom::CrewlinkerPostgresTenant",
			ResourceProperties: map[string]interface{}{
				"Name":            "foo",
				"MasterSecretArn": "arn:aws:secretsmanager:eu-central-1:860345245734:secret:bogus-asdf",
			},
		}, MatchError(MatchRegexp(`unsupported request type`)), Equal(postgresresource.Output{})),
	)

	Describe("integration", Ordered, func() {
		var secretArn string

		BeforeAll(func(ctx context.Context) {
			secret, err := smc.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
				Name: aws.String("ClCoreTesting_" + string(name)),
				SecretString: aws.String(`{
					"username":"postgres",
					"password":"postgres",
					"port":` + os.Getenv("CLPOSTGRES_PORT") + `,
					"host":"localhost"					
				}`),
			})
			Expect(err).ToNot(HaveOccurred())
			secretArn = *secret.ARN
		})

		AfterAll(func(ctx context.Context) {
			_, err := smc.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
				SecretId:             &secretArn,
				RecoveryWindowInDays: aws.Int64(7),
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create", func(ctx context.Context) {
			create, err := hdl.Handle(ctx, postgresresource.Input{
				ResourceType: "Custom::CrewlinkerPostgresTenant",
				RequestType:  cfn.RequestCreate,
				ResourceProperties: map[string]interface{}{
					"Name":            name,
					"MasterSecretArn": secretArn,
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(create).To(Equal(postgresresource.Output{
				PhysicalResourceID: name.ToPhysicalResourceID(),
				Data: map[string]any{
					"DatabaseName": name.ToDatabaseName(),
					"DatabaseUser": name.ToDatabaseUser(),
				},
			}))

			ExpectDatabaseCountWithName(ctx, rw, name.ToDatabaseName(), Equal(1))
			ExpectUserCountWithName(ctx, rw, name.ToDatabaseUser(), Equal(1))
		})

		It("should cleanup if create fails halfway", func(ctx context.Context) {
			By("causing user exists error", func() {
				Expect(rw.Exec(ctx, `CREATE USER `+name.ToDatabaseUser())).ToNot(BeNil())
			})

			resp, err := hdl.Handle(ctx, postgresresource.Input{
				ResourceType: "Custom::CrewlinkerPostgresTenant",
				RequestType:  cfn.RequestCreate,
				ResourceProperties: map[string]interface{}{
					"Name":            name,
					"MasterSecretArn": secretArn,
				},
			})

			Expect(err).To(MatchError(MatchRegexp(`already exists`)))
			Expect(resp).To(Equal(postgresresource.Output{}))

			ExpectDatabaseCountWithName(ctx, rw, name.ToDatabaseName(), Equal(0))
			ExpectUserCountWithName(ctx, rw, name.ToDatabaseUser(), Equal(1)) // should not have removed original user
		})

		Describe("with existing tenant", func() {
			BeforeEach(func(ctx context.Context) {
				Expect(hdl.Handle(ctx, postgresresource.Input{
					ResourceType: "Custom::CrewlinkerPostgresTenant",
					RequestType:  cfn.RequestCreate,
					ResourceProperties: map[string]interface{}{
						"Name":            name,
						"MasterSecretArn": secretArn,
					},
				})).To(HaveField("PhysicalResourceID", name.ToPhysicalResourceID()))

				By("check that we can fill up the tenant database", func() {
					Expect(hdl.OnOtherDB(ctx, dbcfg.ConnConfig, name.ToDatabaseName(), func(ctx context.Context, conn *pgx.Conn) error {
						Expect(conn.Exec(ctx, "CREATE SCHEMA foo")).ToNot(BeNil())
						Expect(conn.Exec(ctx, "CREATE TABLE foo.bar (id UUID PRIMARY KEY)")).ToNot(BeNil())

						return nil
					})).To(Succeed())
				})

				// simulate an open connection to the database, all logic should stil work
				tcfg := dbcfg.Copy()
				tcfg.ConnConfig.Database = name.ToDatabaseName()
				tconn, err := pgx.ConnectConfig(ctx, tcfg.ConnConfig)
				Expect(err).ToNot(HaveOccurred())
				DeferCleanup(tconn.Close)
			})

			It("should error if the master arn is updated", func(ctx context.Context) {
				_, err := hdl.Handle(ctx, postgresresource.Input{
					ResourceType: "Custom::CrewlinkerPostgresTenant",
					RequestType:  cfn.RequestUpdate,
					OldResourceProperties: map[string]interface{}{
						"Name":            name,
						"MasterSecretArn": secretArn,
					},
					ResourceProperties: map[string]interface{}{
						"Name":            updName,
						"MasterSecretArn": secretArn + "_upd",
					},
				})
				Expect(err).To(MatchError(MatchRegexp(`not supported`)))
			})

			It("should allow update", func(ctx context.Context) {
				Expect(hdl.Handle(ctx, postgresresource.Input{
					ResourceType: "Custom::CrewlinkerPostgresTenant",
					RequestType:  cfn.RequestUpdate,
					OldResourceProperties: map[string]interface{}{
						"Name":            name,
						"MasterSecretArn": secretArn,
					},
					ResourceProperties: map[string]interface{}{
						"Name":            updName,
						"MasterSecretArn": secretArn,
					},
				})).To(Equal(postgresresource.Output{
					PhysicalResourceID: updName.ToPhysicalResourceID(),
					Data: map[string]any{
						"DatabaseName": updName.ToDatabaseName(),
						"DatabaseUser": updName.ToDatabaseUser(),
					},
				}))

				ExpectDatabaseCountWithName(ctx, rw, updName.ToDatabaseName(), Equal(1))
				ExpectUserCountWithName(ctx, rw, updName.ToDatabaseUser(), Equal(1))

				ExpectDatabaseCountWithName(ctx, rw, name.ToDatabaseName(), Equal(0))
				ExpectUserCountWithName(ctx, rw, name.ToDatabaseUser(), Equal(0))
			})

			It("should handle failed update", func(ctx context.Context) {
				By("causing user exists error", func() {
					Expect(rw.Exec(ctx, `CREATE USER `+updName.ToDatabaseUser())).ToNot(BeNil())
				})

				// this will fail because we can't rename the old user to the new user since it already exists.
				_, err := hdl.Handle(ctx, postgresresource.Input{
					ResourceType: "Custom::CrewlinkerPostgresTenant",
					RequestType:  cfn.RequestUpdate,
					OldResourceProperties: map[string]interface{}{
						"Name":            name,
						"MasterSecretArn": secretArn,
					},
					ResourceProperties: map[string]interface{}{
						"Name":            updName,
						"MasterSecretArn": secretArn,
					},
				})
				Expect(err).To(MatchError(MatchRegexp(`already exists`)))

				ExpectDatabaseCountWithName(ctx, rw, updName.ToDatabaseName(), Equal(1))
				ExpectUserCountWithName(ctx, rw, updName.ToDatabaseUser(), Equal(1)) // we've created it at the start of the test

				// At this point the update has succeeded partially. The old database has been renamed to the
				// new database name. And the old user is NOT renamed, is now owner of the new database. Cloudformation will issue an
				// update with the parameters reversed.
				//
				// This will cause the renamed database to be named the old database (as expected). And it will try to rename the upd
				// user (which we created manually) to the non-update user (that already exists).
				_, err = hdl.Handle(ctx, postgresresource.Input{
					ResourceType: "Custom::CrewlinkerPostgresTenant",
					RequestType:  cfn.RequestUpdate,
					OldResourceProperties: map[string]interface{}{
						"Name":            updName,
						"MasterSecretArn": secretArn,
					},
					ResourceProperties: map[string]interface{}{
						"Name":            name,
						"MasterSecretArn": secretArn,
					},
				})
				Expect(err).To(MatchError(MatchRegexp(`already exists`)))

				// At this point the situation is locked and manual intervention is required. And the developer will need to
				// do "Continue Update Rollback" in the cdn console.
			})

			It("should delete cleanly", func(ctx context.Context) {
				Expect(hdl.Handle(ctx, postgresresource.Input{
					PhysicalResourceID: name.ToPhysicalResourceID(),
					ResourceType:       "Custom::CrewlinkerPostgresTenant",
					RequestType:        cfn.RequestDelete,
					ResourceProperties: map[string]interface{}{
						"Name":            name,
						"MasterSecretArn": secretArn,
					},
				})).To(Equal(postgresresource.Output{
					PhysicalResourceID: name.ToPhysicalResourceID(),
				}))

				ExpectDatabaseCountWithName(ctx, rw, name.ToDatabaseName(), Equal(0))
				ExpectUserCountWithName(ctx, rw, name.ToDatabaseUser(), Equal(0))
			})
		})
	})
})

func ExpectDatabaseCountWithName(ctx context.Context, rw *pgxpool.Pool, name string, exp OmegaMatcher) {
	GinkgoHelper()

	var count int

	Expect(rw.QueryRow(ctx, `SELECT COUNT(*) FROM pg_database WHERE datname = $1;`, name).Scan(&count)).To(Succeed())
	Expect(count).To(exp)
}

func ExpectUserCountWithName(ctx context.Context, rw *pgxpool.Pool, name string, exp OmegaMatcher) {
	GinkgoHelper()

	var count int

	Expect(rw.QueryRow(ctx, `SELECT COUNT(*) FROM pg_user WHERE usename = $1;`, name).Scan(&count)).To(Succeed())
	Expect(count).To(exp)
}
