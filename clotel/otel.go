package clotel

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-logr/zapr"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"

	ecsdetector "go.opentelemetry.io/contrib/detectors/aws/ecs"
)

// Config configures the code in this package.
type Config struct {
	// DetectorDetectTimeout bound the time it may take to init a trace provider
	DetectorDetectTimeout time.Duration `env:"DETECTOR_DETECT_TIMEOUT" envDefault:"100ms"`
	// ExporterTimeout overwrites the timeout for exporting spans. This can be usefull in tests to speed
	// them up
	ExporterTimeout time.Duration `env:"EXPORTER_TIMEOUT" envDefault:"10s"`
	// ExporterEndpoint configures where otel span exporter will send data to
	ExporterEndpoint string `env:"EXPORTER_ENDPOINT" envDefault:"localhost:4317"`
}

// new inits a tracer provider
func New(
	cfg Config,
	logs *zap.Logger,
	exp sdktrace.SpanExporter,
	det resource.Detector,
	idg sdktrace.IDGenerator,
	pr propagation.TextMapPropagator,
) (*sdktrace.TracerProvider, error) {

	// detect the resource with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DetectorDetectTimeout)
	defer cancel()
	res, err := det.Detect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource: %w", err)
	}

	// we handle otel errors by logging it with our zap logger. This is unfortunately a global
	// setting so it may confuse testing setups
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logs.Error("otel error", zap.Error(err))
	}))

	// for sdk logging we also need to set a global value
	otel.SetLogger(zapr.NewLogger(logs))
	// set the global text map propagator
	otel.SetTextMapPropagator(pr)

	// finally, init the actual provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exp),
		sdktrace.WithIDGenerator(idg),
	)

	// set it globally, but code should prefer to inject it during construction
	otel.SetTracerProvider(tp)

	// set default htt transport and client to use tracing
	http.DefaultTransport = otelhttp.NewTransport(http.DefaultTransport,
		otelhttp.WithPropagators(pr),
		otelhttp.WithTracerProvider(tp))
	http.DefaultClient = &http.Client{Transport: http.DefaultTransport}
	return tp, nil
}

// moduleName for naming conventions
const moduleName = "clotel"

// shared module with di setup shared between test and prod environment
var shared = fx.Module(moduleName,
	// the incoming logger will be named after the module
	fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
	// provide the environment configuration
	fx.Provide(fx.Annotate(
		func(o env.Options) (c Config, err error) {
			o.Prefix = strings.ToUpper(moduleName) + "_"
			return c, env.Parse(&c, o)
		},
		fx.ParamTags(`optional:"true"`))),

	// we can use the xray id generator in all cases
	fx.Provide(fx.Annotate(xray.NewIDGenerator, fx.As(new(sdktrace.IDGenerator)))),
	// we also provide an xray propagator for anywhere it code we need this
	fx.Provide(func() propagation.TextMapPropagator { xp := xray.Propagator{}; return xp }),
	// provide the tracer provider
	fx.Provide(fx.Annotate(New,
		fx.OnStop(func(ctx context.Context, tp *sdktrace.TracerProvider) error { return tp.Shutdown(ctx) }),
	)),

	// also provide as more generic interface
	fx.Provide(func(tp *sdktrace.TracerProvider) trace.TracerProvider { return tp }),
)

// newGrcpExporter rturns the grpc exporter
func newGrpcExporter(cfg Config) *otlptrace.Exporter {
	return otlptracegrpc.NewUnstarted(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(cfg.ExporterTimeout),
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
	)
}

// Service provides otel dependencies for container services
var Service = fx.Options(shared,
	// service will export traces over grpc
	fx.Provide(fx.Annotate(newGrpcExporter,
		fx.OnStart(func(ctx context.Context, e *otlptrace.Exporter) error { return e.Start(ctx) }),
		fx.OnStop(func(ctx context.Context, e *otlptrace.Exporter) error { return e.Shutdown(ctx) }),
	)),
	fx.Provide(func(e *otlptrace.Exporter) sdktrace.SpanExporter { return e }),
	fx.Provide(ecsdetector.NewResourceDetector),
)

// Test configures the DI for a test environment
var Test = fx.Options(shared,
	fx.Provide(fx.Annotate(tracetest.NewInMemoryExporter)),
	fx.Provide(func(e *tracetest.InMemoryExporter) sdktrace.SpanExporter { return e }),
	fx.Provide(func() resource.Detector {
		return resource.StringDetector(semconv.SchemaURL, semconv.ServiceNameKey, func() (string, error) {
			return "ClTest", nil
		})
	}),
)
