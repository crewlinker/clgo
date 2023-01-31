package clotel

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/zapr"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

// NewTracerProvider inits a tracer provider
func NewTracerProvider(
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

	logs.Info("detected resource", zap.Stringer("attributes", res))

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

// newGrcpExporter rturns the grpc exporter
func newGrpcExporter(cfg Config) *otlptrace.Exporter {
	return otlptracegrpc.NewUnstarted(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(cfg.ExporterTimeout),
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
	)
}

// we need a custom detector because of the log correlation issue described here:
// https://github.com/aws-observability/aws-otel-collector/issues/1766
type extraEcsDetector struct{ resource.Detector }

func (det extraEcsDetector) Detect(ctx context.Context) (res *resource.Resource, err error) {
	res, err = det.Detector.Detect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect: %w", err)
	}

	if res == nil {
		return res, nil
	}

	var kvs []attribute.KeyValue
	if arns, ok := res.Set().Value(semconv.AWSLogGroupARNsKey); ok {
		kvs = append(kvs, semconv.AWSLogGroupARNsKey.StringSlice([]string{arns.AsString()}))
	}
	if names, ok := res.Set().Value(semconv.AWSLogGroupNamesKey); ok {
		kvs = append(kvs, semconv.AWSLogGroupNamesKey.StringSlice([]string{names.AsString()}))
	}

	// instead set the attributes as string slices for the otel exporter to enable log2trace correlation
	return resource.Merge(res, resource.NewSchemaless(kvs...))
}

// WithExtraEcsAttributes decorates the detector with extra ecs attributess to fix log tracing
func WithExtraEcsAttributes(d resource.Detector) resource.Detector {
	return &extraEcsDetector{Detector: d}
}
