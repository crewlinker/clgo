package clwebserver_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/crewlinker/clgo/clwebserver"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestWebserver(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clwebserver")
}

var _ = Describe("failed to handle webserver", func() {
	var lnr *net.TCPListener
	BeforeEach(func(ctx context.Context) {
		app := fx.New(clwebserver.Provide(),
			fx.Decorate(func(c clwebserver.Config) clwebserver.Config {
				c.BindAddrPort = "127.0.0.1:0" // random port for parallel tests

				return c
			}),
			fx.Invoke(func(s *http.Server) {}),
			fx.Populate(&lnr),
			fx.Supply(fx.Annotate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
				fx.As(new(http.Handler)))),
			clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Done)
	})

	It("should server http", func() {
		Expect(http.Get(fmt.Sprintf("http://%s", lnr.Addr()))).To(HaveHTTPStatus(200))
	})
})
