package clsentry_test

import (
	"context"
	"testing"
	"time"

	"github.com/crewlinker/clgo/clbuildinfo"
	"github.com/crewlinker/clgo/clsentry"
	"github.com/crewlinker/clgo/clzap"
	sentry "github.com/getsentry/sentry-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestModel(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clsentry")
}

var _ = Describe("just sentry", func() {
	var hub *sentry.Hub
	var obs *clsentry.ObservedEvents

	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&hub, &obs), Provide(make(chan string)))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should convert event id", func() {
		Expect(clsentry.StringFromEventID(nil)).To(BeNil())
		id := sentry.EventID("some-id")
		Expect(clsentry.StringFromEventID(&id)).To(HaveValue(Equal("some-id")))
	})

	for range 10 {
		It("should receive the failed fx event", func() {
			Expect(hub).NotTo(BeNil())
			hub.CaptureMessage("some message")
			hub.Flush(time.Second)
			Expect(obs.Events()).To(HaveLen(1))
			Expect(obs.Events()[0].Message).To(Equal("some message"))
		})
	}
})

func Provide(sent chan string) fx.Option {
	return fx.Options(
		clbuildinfo.TestProvide(),
		clsentry.TestProvide(),
		clzap.TestProvide(),
	)
}
