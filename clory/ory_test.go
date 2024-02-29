package clory_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crewlinker/clgo/clory"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	orysdk "github.com/ory/client-go"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
)

func TestOryauth(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clory")
}

var _ = Describe("di", func() {
	var ory *clory.Ory
	var front *MockFrontendAPI

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&ory),
			clory.Provide(),
			clzap.TestProvide(),
			WithMocked(&front))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should provide dep", func() {
		Expect(ory).ToNot(BeNil())
	})

	It("should return login url", func() {
		Expect(ory.BrowserLoginURL().String()).To(Equal("http://localhost:4000/self-service/login/browser"))
	})

	Describe("private middleware", func() {
		var hdl http.Handler
		BeforeEach(func() {
			hdl = ory.Private(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "hello: %s", clory.Session(r.Context()).Id)
			}))
		})

		It("should server private with active session", func() {
			front.EXPECT().ToSession(mock.Anything).Return(orysdk.FrontendAPIToSessionRequest{})
			front.EXPECT().ToSessionExecute(mock.Anything).Return(&orysdk.Session{
				Id:     "some_session_id",
				Active: orysdk.PtrBool(true),
			}, nil, nil)

			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			hdl.ServeHTTP(resp, req)

			Expect(resp.Result().StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Body.String()).To(Equal(`hello: some_session_id`))
		})

		It("should unauthorized when error", func() {
			front.EXPECT().ToSession(mock.Anything).Return(orysdk.FrontendAPIToSessionRequest{})
			front.EXPECT().ToSessionExecute(mock.Anything).Return(nil, nil, errors.New("foo"))

			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			hdl.ServeHTTP(resp, req)

			Expect(resp.Result().StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(resp.Header().Get("X-Browser-Login-URL")).To(Equal("http://localhost:4000/self-service/login/browser"))
		})

		It("should unauthorized when session inactive", func() {
			front.EXPECT().ToSession(mock.Anything).Return(orysdk.FrontendAPIToSessionRequest{})
			front.EXPECT().ToSessionExecute(mock.Anything).Return(&orysdk.Session{
				Id:     "some_session_id",
				Active: orysdk.PtrBool(false),
			}, nil, nil)

			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			hdl.ServeHTTP(resp, req)

			Expect(resp.Result().StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(resp.Header().Get("X-Browser-Login-URL")).To(Equal("http://localhost:4000/self-service/login/browser"))
		})
	})
})

// WithMocked is a test helper that mocks handler dependencies.
func WithMocked(front **MockFrontendAPI) fx.Option {
	return fx.Options(
		fx.Decorate(func(clory.FrontendAPI) clory.FrontendAPI {
			mock := NewMockFrontendAPI(GinkgoT())
			*front = mock

			return mock
		}),
	)
}
