package clauth_test

import (
	"context"

	"github.com/crewlinker/clgo/clauth"
	"github.com/crewlinker/clgo/clzap"
	"github.com/lestrrat-go/jwx/v2/jwt/openid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

var _ = Describe("authn", func() {
	var autn *clauth.Authn

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&autn),
			clauth.TestProvide(map[string]string{}),
			clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should sign and authenticate", func(ctx context.Context) {
		tok1, err := openid.NewBuilder().Email("foo@foo.bar").Build()
		Expect(err).ToNot(HaveOccurred())

		out, err := autn.SignJWT(ctx, tok1)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).ToNot(BeEmpty())

		tok2, err := autn.AuthenticateJWT(ctx, out)
		Expect(err).ToNot(HaveOccurred())

		Expect(tok2.Email()).To(Equal("foo@foo.bar"))
	})

	DescribeTable("AuthenticateJWT", func(ctx SpecContext, inp []byte, expErr OmegaMatcher) {
		_, err := autn.AuthenticateJWT(ctx, inp)
		Expect(err).To(expErr)
	},
		Entry("empty", nil,
			MatchError(MatchRegexp(`invalid byte sequence`))),
		Entry("no key id", []byte(`eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c`),
			MatchError(MatchRegexp(`no key ID`))),
	)
})
