// Package clotel provides re-usable components for OpenTelemetry integration
package clotel

import (
	"context"
	"fmt"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// moduleName for naming conventions.
const moduleName = "clotel"

// Base module with di setup Base between test and prod environment.
func Base() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// we can use the xray id generator in all cases
		fx.Provide(fx.Annotate(xray.NewIDGenerator, fx.As(new(sdktrace.IDGenerator)))),
		// we also provide an xray propagator for anywhere it code we need this
		fx.Provide(func() propagation.TextMapPropagator {
			xp := xray.Propagator{}

			return xp
		}),
		// provide the tracer provider
		fx.Provide(fx.Annotate(NewTracerProvider,
			fx.OnStop(func(ctx context.Context, tp *sdktrace.TracerProvider) error {
				if err := tp.Shutdown(ctx); err != nil {
					return fmt.Errorf("failed to shutdown: %w", err)
				}

				return nil
			}),
		)),
		// also provide as more generic interface
		fx.Provide(func(tp *sdktrace.TracerProvider) trace.TracerProvider { return tp }),
		// provide the metrer provider
		fx.Provide(fx.Annotate(NewMeterProvider)),
		// also provide as more generic interface
		fx.Provide(func(mp *sdkmetric.MeterProvider) metric.MeterProvider { return mp }),
	)
}

// Service provides otel dependencies for container services.
func Service() fx.Option {
	return fx.Options(Base(),
		// service will export traces over grpc
		fx.Provide(fx.Annotate(newGrpcExporter,
			fx.OnStart(func(ctx context.Context, e *otlptrace.Exporter) error {
				if err := e.Start(ctx); err != nil {
					return fmt.Errorf("failed to start: %w", err)
				}

				return nil
			}),
			fx.OnStop(func(ctx context.Context, e *otlptrace.Exporter) error {
				if err := e.Shutdown(ctx); err != nil {
					return fmt.Errorf("failed to shutdown: %w", err)
				}

				return nil
			}),
		)),
		// provide the grpc exporter as a generic span exporter as well
		fx.Provide(func(e *otlptrace.Exporter) sdktrace.SpanExporter { return e }),
		// detect expects ecs resource
		fx.Provide(ecs.NewResourceDetector),
		// decorate to fix an issue that prevents log correlation
		fx.Decorate(WithExtraEcsAttributes),

		// provide dependencies for metric export
		fx.Provide(sdkmetric.NewPeriodicReader),
		fx.Provide(NewMetricExporter),
	)
}

// Test configures the DI for a test environment.
func Test() fx.Option {
	return fx.Options(Base(),
		fx.Provide(sdkmetric.NewManualReader),
		fx.Provide(fx.Annotate(tracetest.NewInMemoryExporter)),
		fx.Provide(func(e *tracetest.InMemoryExporter) sdktrace.SpanExporter { return e }),
		fx.Provide(func() resource.Detector {
			return resource.StringDetector(semconv.SchemaURL, semconv.ServiceNameKey, func() (string, error) {
				return "ClTest", nil
			})
		}),
	)
}
