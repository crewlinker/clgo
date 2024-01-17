package clauthn_test

import (
	"context"
	"testing"

	"github.com/crewlinker/clgo/clauthn"
	"github.com/crewlinker/clgo/clzap"
	"github.com/lestrrat-go/jwx/v2/jwt/openid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestAuthn(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clauthn")
}

var _ = Describe("authn", func() {
	var authn *clauthn.Authn

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&authn),
			clauthn.TestProvide(),
			clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should sign and authenticate", func(ctx context.Context) {
		tok1, err := openid.NewBuilder().Email("foo@foo.bar").Build()
		Expect(err).ToNot(HaveOccurred())

		out, err := authn.SignJWT(ctx, tok1)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).ToNot(BeEmpty())

		tok2, err := authn.AuthenticateJWT(ctx, out)
		Expect(err).ToNot(HaveOccurred())

		Expect(tok2.Email()).To(Equal("foo@foo.bar"))
	})

	DescribeTable("AuthenticateJWT", func(ctx SpecContext, inp []byte, expErr OmegaMatcher) {
		_, err := authn.AuthenticateJWT(ctx, inp)
		Expect(err).To(expErr)
	},
		Entry("empty", nil,
			MatchError(MatchRegexp(`invalid byte sequence`))),
		Entry("no key id", []byte(`eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c`),
			MatchError(MatchRegexp(`no key ID`))),
	)
})

var _ = Describe("authnprovide", func() {
	var authn *clauthn.Authn

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&authn),
			clauthn.Provide(),
			clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should init", func() {
		Expect(authn).ToNot(BeNil())
	})
})
