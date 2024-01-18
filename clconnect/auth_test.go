package clconnect_test

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clauthn"
	"github.com/crewlinker/clgo/clauthz"
	"github.com/crewlinker/clgo/clconnect"
	clconnectv1 "github.com/crewlinker/clgo/clconnect/v1"
	"github.com/crewlinker/clgo/clconnect/v1/clconnectv1connect"
	"github.com/lestrrat-go/jwx/v2/jwt/openid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"
)

var _ = Describe("auth", func() {
	var hdl http.Handler
	var rwc clconnectv1connect.ReadWriteServiceClient
	var roc clconnectv1connect.ReadOnlyServiceClient
	var obs *observer.ObservedLogs
	var authn *clauthn.Authn

	BeforeEach(func(ctx context.Context) {
		policies := map[string]string{
			"main.rego": `
				package authz
				import rego.v1

				default allow := false

				allow if {
					input.open_id.sub == "sub2"
					input.env.foo = "bar"
				}
`,
		}

		app := fx.New(
			fx.Populate(fx.Annotate(&hdl, fx.ParamTags(`name:"clconnect"`)), &rwc, &roc, &obs, &authn),
			fx.Decorate(func(c clconnect.Config) clconnect.Config {
				c.AuthzPolicyEnvInput = `{"foo":"bar"}` // set some environment for the policy to work

				return c
			}),
			fx.Decorate(func(b clauthz.MockBundle) clauthz.MockBundle {
				return clauthz.MockBundle(policies)
			}),

			ProvideEnt(),
		)

		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("anonymous permission denied", func(ctx context.Context) {
		req := &connect.Request[clconnectv1.FooRequest]{Msg: &clconnectv1.FooRequest{}}

		_, err := roc.Foo(ctx, req)
		Expect(err).To(MatchError(MatchRegexp(`unauthorized, subject: ''`)))

		var cerr *connect.Error
		Expect(errors.As(err, &cerr)).To(BeTrue())
		Expect(cerr.Code()).To(Equal(connect.CodePermissionDenied))
	})

	It("invalid token unauthenticated", func(ctx context.Context) {
		req := &connect.Request[clconnectv1.FooRequest]{Msg: &clconnectv1.FooRequest{}}
		req.Header().Set("Authorization", "Bearer "+"eyJhbGciOiJub25lIn0.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ")

		_, err := roc.Foo(ctx, req)
		Expect(err).To(MatchError(MatchRegexp(`invalid JWT`)))

		var cerr *connect.Error
		Expect(errors.As(err, &cerr)).To(BeTrue())
		Expect(cerr.Code()).To(Equal(connect.CodeUnauthenticated))
	})

	Describe("with openid token", func() {
		var tok1, tok2 string
		BeforeEach(func(ctx context.Context) {
			tok1 = string(lo.Must(authn.SignJWT(ctx, lo.Must(openid.NewBuilder().
				Subject("sub1").
				Build()))))
			tok2 = string(lo.Must(authn.SignJWT(ctx, lo.Must(openid.NewBuilder().
				Subject("sub2").
				GivenName("John").
				FamilyName("Doe").
				Build()))))
		})

		It("valid token but no permission", func(ctx context.Context) {
			req := &connect.Request[clconnectv1.FooRequest]{Msg: &clconnectv1.FooRequest{}}
			req.Header().Set("Authorization", "Bearer "+tok1)

			_, err := roc.Foo(ctx, req)
			Expect(err).To(MatchError(MatchRegexp(`unauthorized, subject: 'sub1'`)))
		})

		It("valid token with permission", func(ctx context.Context) {
			req := &connect.Request[clconnectv1.FooRequest]{Msg: &clconnectv1.FooRequest{}}
			req.Header().Set("Authorization", "Bearer "+tok2)

			resp, err := roc.Foo(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Msg.GetBar()).To(Equal(`Name: John Doe`))
		})
	})
})
