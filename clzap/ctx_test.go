package clzap_test

import (
	"context"

	"github.com/aws/aws-lambda-go/lambdacontext"
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
		app := fx.New(clzap.TestProvide(), fx.Populate(&logs, &obs))
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

	It("should return fallback logger", func() {
		logs1 := clzap.Log(ctx1, logs)
		Expect(logs1).To(Equal(logs))

		logs1.Info("backup")
		Expect(obs.FilterMessage("backup").Len()).To(Equal(1))
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

		ctx2 = lambdacontext.NewContext(ctx2, &lambdacontext.LambdaContext{
			AwsRequestID: "79b4f56e-95b1-4643-9700-2807f4e68189",
		})

		clzap.Log(ctx2).Info("foo")
		entries := obs.FilterMessage("foo")
		Expect(entries.Len()).To(Equal(1))

		Expect(entries.All()[0].ContextMap()).To(
			HaveKeyWithValue("trace_id", "1-01000000-000000000000000000000000"))
		Expect(entries.All()[0].ContextMap()).To(
			HaveKeyWithValue("span_id", "0200000000000000"))
		Expect(entries.All()[0].ContextMap()).To(
			HaveKeyWithValue("requestId", "79b4f56e-95b1-4643-9700-2807f4e68189"))
	})
})
