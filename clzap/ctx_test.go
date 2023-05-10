package clzap_test

import (
	"context"

	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

var _ = Describe("context", func() {
	var ctx1 context.Context
	var ctx2 context.Context
	var logs *zap.Logger
	var obs *observer.ObservedLogs

	BeforeEach(func() {
		ctx1 = context.Background()
		app := fx.New(clzap.Test(), fx.Populate(&logs, &obs))
		ctx2 = clzap.WithLogger(context.Background(), logs)
		Expect(app.Start(ctx1)).To(Succeed())
		DeferCleanup(app.Stop, ctx1)
	})

	It("should return false if no logger", func() {
		logs1, ok := clzap.LoggerFromContext(ctx1)
		Expect(ok).To(BeFalse())
		Expect(logs1).To(BeNil())
	})

	It("should return nop logger", func() {
		logs1 := clzap.Log(ctx1)
		Expect(logs1).To(Equal(zap.NewNop()))

		logs1.Info("foo")
		Expect(obs.FilterMessage("foo").Len()).To(Equal(0))
	})

	It("should return regular logger", func() {
		clzap.Log(ctx2).Info("foo")
		Expect(obs.FilterMessage("foo").Len()).To(Equal(1))
	})

	It("should log span/trace info", func() {
		ctx2 = trace.ContextWithSpanContext(ctx2, trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: trace.TraceID{0x01},
			SpanID:  trace.SpanID{0x02},
		}))

		clzap.Log(ctx2).Info("foo")
		entries := obs.FilterMessage("foo")
		Expect(entries.Len()).To(Equal(1))

		Expect(entries.All()[0].ContextMap()).To(
			HaveKeyWithValue("trace_id", "1-01000000-000000000000000000000000"))
		Expect(entries.All()[0].ContextMap()).To(
			HaveKeyWithValue("span_id", "0200000000000000"))
	})
})
