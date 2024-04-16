package clcedard_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/crewlinker/clgo/clcedard"
	"github.com/crewlinker/clgo/clzap"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

func TestClcedard(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clcedard")
}

var _ = BeforeSuite(func() {
	godotenv.Load(filepath.Join("..", "test.env"))
})

func serve(w http.ResponseWriter, r *http.Request) {
	var input clcedard.Input

	lo.Must0(json.NewDecoder(r.Body).Decode(&input))

	switch input.Principal {
	case `unauthorized`:
		w.WriteHeader(http.StatusUnauthorized)
	case `cedar-errors`:
		lo.Must0(json.NewEncoder(w).Encode(clcedard.Output{
			Decision:      "Deny",
			ErrorMessages: []string{"some error1", "some error2"},
		}))

	default:
		lo.Must0(json.NewEncoder(w).Encode(clcedard.Output{
			Decision: "Allow",
		}))
	}
}

var _ = Describe("clcedard", func() {
	var ccl *clcedard.Client
	var srv *httptest.Server

	BeforeEach(func(ctx context.Context) {
		srv = httptest.NewServer(http.HandlerFunc(serve))

		app := fx.New(
			fx.Populate(&ccl),
			fx.Decorate(func(c clcedard.Config) clcedard.Config {
				c.BaseURL = srv.URL

				return c
			}),
			clcedard.Provide(),
			clzap.TestProvide(),
			fx.Supply(http.DefaultClient),
		)

		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	Describe("IsAuthorized", func() {
		It("should allow", func(ctx context.Context) {
			res, err := ccl.IsAuthorized(ctx, &clcedard.Input{})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})

		It("should fail unauthorized", func(ctx context.Context) {
			res, err := ccl.IsAuthorized(ctx, &clcedard.Input{InputItem: clcedard.InputItem{Principal: "unauthorized"}})
			Expect(err).To(MatchError(MatchRegexp("401")))
			Expect(res).To(BeFalse())
		})

		It("should fail with errors", func(ctx context.Context) {
			res, err := ccl.IsAuthorized(ctx, &clcedard.Input{InputItem: clcedard.InputItem{Principal: "cedar-errors"}})
			Expect(err).To(MatchError(MatchRegexp("error1")))
			Expect(res).To(BeFalse())
		})
	})
})
