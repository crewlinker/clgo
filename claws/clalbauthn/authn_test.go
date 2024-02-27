package clalbauthn_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crewlinker/clgo/claws/clalbauthn"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestClalbauthn(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "claws/clalbauthn")
}

var _ = Describe("middleware", func() {
	It("should add context values", func() {
		resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Set("X-Amzn-Oidc-Accesstoken", "access_token_1")
		req.Header.Set("x-amzn-Oidc-identity", "identity_1")
		req.Header.Set("x-amzn-oidc-data", "data_1")

		var accessToken, identity, claimData string
		clalbauthn.New(clalbauthn.Config{}, http.Error).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessToken, identity, claimData = clalbauthn.AccessToken(r.Context()),
				clalbauthn.Identity(r.Context()),
				clalbauthn.ClaimData(r.Context())
		})).ServeHTTP(resp, req)

		Expect(accessToken).To(Equal(`access_token_1`))
		Expect(identity).To(Equal(`identity_1`))
		Expect(claimData).To(Equal(`data_1`))
	})

	It("should return unauth without all header", func() {
		resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Set("X-Amzn-Oidc-Accesstoken", "access_token_1")
		req.Header.Set("x-amzn-Oidc-identity", "identity_1")

		clalbauthn.New(clalbauthn.Config{}, http.Error).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})).ServeHTTP(resp, req)

		Expect(resp).To(HaveHTTPStatus(http.StatusUnauthorized))
		Expect(resp.Body.String()).To(Equal("Unauthorized\n"))
	})

	It("should return 200 if unauth is allowed", func() {
		resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Set("X-Amzn-Oidc-Accesstoken", "access_token_1")
		req.Header.Set("x-amzn-Oidc-identity", "identity_1")

		clalbauthn.New(clalbauthn.Config{AllowUnauthenticated: true}, http.Error).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})).ServeHTTP(resp, req)

		Expect(resp).To(HaveHTTPStatus(http.StatusOK))
	})
})

var _ = Describe("di", func() {
	var authn *clalbauthn.Authentication
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&authn), clalbauthn.Provide(), clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should have setup di", func() {
		Expect(authn).ToNot(BeNil())
	})
})
