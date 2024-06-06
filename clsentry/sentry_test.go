package clsentry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crewlinker/clgo/clbuildinfo"
	"github.com/crewlinker/clgo/clsentry"
	"github.com/crewlinker/clgo/clzap"
	sentry "github.com/getsentry/sentry-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func TestModel(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clsentry")
}

var _ = Describe("zap sentry", func() {
	var logs *zap.Logger
	var sent chan string
	var count uint64

	BeforeEach(func(ctx context.Context) {
		beforeSend := clsentry.BeforeSendFunc(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			atomic.AddUint64(&count, 1)

			return event
		})

		sent = make(chan string, 1)
		app := fx.New(fx.Populate(&logs), Provide(sent), fx.Supply(beforeSend))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should log zap errors to sentry", func() {
		logs.Error("some error for sentry")
		Eventually(sent).Should(Receive(ContainSubstring(`"message":"some error for sentry"`)))
		Expect(atomic.LoadUint64(&count)).To(BeNumerically(">", 0))
	})
})

var _ = Describe("fx zap sentry", Serial, func() {
	var logs *zap.Logger
	var hdl http.Handler
	var sent chan string

	BeforeEach(func(ctx context.Context) {
		clsentry.FxErrorShutdownDelay = time.Millisecond * 100

		sent = make(chan string, 1)
		app := fx.New(fx.Populate(&logs, &hdl), Provide(sent))
		app.Start(ctx)
	})

	It("should receive the failed fx event", func() {
		Eventually(sent).Should(Receive(ContainSubstring(`"error":"missing dependencies for function`)))
	})
})

func Provide(sent chan string) fx.Option {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := lo.Must(io.ReadAll(r.Body))

		sent <- string(body)
	}))

	// note: don't close the server as it will cause race conditions

	loc, _ := url.Parse(srv.URL)
	loc.User = url.UserPassword("someuser", "")

	return fx.Options(
		fx.Decorate(func(c clsentry.Config) clsentry.Config {
			c.DSN = loc.String() + "/someproject"

			return c
		}),
		clbuildinfo.TestProvide(),
		clsentry.Provide(),
		clzap.TestProvide(),
	)
}
