package clzap_test

import (
	"context"

	"github.com/caarlos0/env/v10"
	"github.com/crewlinker/clgo/clzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fatalHook struct{}

func (fatalHook) OnWrite(*zapcore.CheckedEntry, []zapcore.Field) {}

var _ = Describe("context logging", func() {
	var logs *zap.Logger
	var clogs *clzap.ContextLogger
	var obs *observer.ObservedLogs
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&logs, &obs), clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
		clogs = clzap.NewTraceContextLogger(logs)
	})

	It("should add span, trace and profile id", func(ctx context.Context) {
		ctx = trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: trace.TraceID{0x01},
			SpanID:  trace.SpanID{0x02},
		}))

		clogs.Info(ctx, "foo")

		Expect(obs.FilterMessage("foo").All()[0].ContextMap()).To(Equal(map[string]any{
			"span_id":  "0200000000000000",
			"trace_id": "1-01000000-000000000000000000000000",
		}))
	})
})

var _ = Describe("default context logger", func() {
	var logs *zap.Logger
	var ctxlogs *clzap.ContextLogger
	var obs *observer.ObservedLogs

	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&logs, &obs), clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
		ctxlogs = clzap.NewContextLogger(logs)
	})
	It("should log", func(ctx SpecContext) {
		ctxlogs.Info(ctx, "foo")
		Expect(obs.FilterMessage("foo").All()).To(HaveLen(1))
	})
})

var _ = Describe("context logger", func() {
	var logs *zap.Logger
	var ctxlogs *clzap.ContextLogger
	var obs *observer.ObservedLogs
	var msg string
	var fields []zap.Field
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&logs, &obs),
			fx.Supply(env.Options{Environment: map[string]string{"CLZAP_LEVEL": "debug"}}),
			fx.Decorate(func(l *zap.Logger) *zap.Logger {
				return l.WithOptions(zap.WithFatalHook(fatalHook{}))
			}),
			clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
		ctxlogs = clzap.NewContextLogger(logs, func(ctx context.Context, f []zap.Field) []zap.Field {
			return append(f, zap.String("hook", "kooh"))
		})
		msg, fields = "my message", []zap.Field{zap.Int64("dar", 10)}
	})

	It("should report level correctly", func() {
		Expect(ctxlogs.Level()).To(Equal(zapcore.DebugLevel))
	})

	DescribeTable("added context fields", func(ctx context.Context, lvl zapcore.Level) {
		logs := ctxlogs.With(zap.String("foo", "bar"))

		switch lvl {
		case zapcore.DebugLevel:
			logs.Debug(ctx, msg, fields...)
		case zapcore.InfoLevel:
			logs.Info(ctx, msg, fields...)
		case zapcore.WarnLevel:
			logs.Warn(ctx, msg, fields...)
		case zapcore.ErrorLevel:
			logs.Error(ctx, msg, fields...)
		case zapcore.PanicLevel:
			Expect(func() {
				logs.Panic(ctx, msg, fields...)
			}).To(Panic())
		case zapcore.DPanicLevel:
			logs.DPanic(ctx, msg, fields...)
		case zapcore.FatalLevel:
			logs.Fatal(ctx, msg, fields...)
		case zapcore.InvalidLevel:
			Fail("invalid")
		default:
			Fail("unsupported")
		}

		Expect(logs.Check(lvl, "some message")).ToNot(BeNil())
		Expect(logs.Sync()).To(Succeed())

		all := obs.FilterMessage(msg).All()
		Expect(all).To(HaveLen(1))
		Expect(all[0].Level).To(Equal(lvl))
		Expect(all[0].ContextMap()).To(Equal(map[string]any{
			"foo":  "bar",
			"dar":  int64(10),
			"hook": "kooh",
		}))
	},
		Entry("debug", zapcore.DebugLevel),
		Entry("info", zapcore.InfoLevel),
		Entry("warn", zapcore.WarnLevel),
		Entry("error", zapcore.ErrorLevel),
		Entry("panic", zapcore.PanicLevel),
		Entry("dpanic", zapcore.DPanicLevel),
		Entry("fatal", zapcore.FatalLevel),
	)

	It("should log with variable level", func(ctx context.Context) {
		logs := ctxlogs.With(zap.String("foo", "bar"))
		logs.Log(ctx, zapcore.InfoLevel, msg, fields...)
		all := obs.FilterMessage(msg).All()
		Expect(all).To(HaveLen(1))
		Expect(all[0].Level).To(Equal(zap.InfoLevel))
		Expect(all[0].ContextMap()).To(Equal(map[string]any{
			"foo":  "bar",
			"dar":  int64(10),
			"hook": "kooh",
		}))
	})
})
