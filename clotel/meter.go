package clotel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// NewMeterProvider initializes otel provider for metrics throughout the application.
func NewMeterProvider(cfg Config, det resource.Detector, mtr metric.Reader) (*metric.MeterProvider, error) {
	// detect the resource with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DetectorDetectTimeout)
	defer cancel()

	res, err := det.Detect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource: %w", err)
	}

	mtp := metric.NewMeterProvider(
		metric.WithReader(mtr),
		metric.WithResource(res))

	// set globally in case libraries don't allow injecting
	global.SetMeterProvider(mtp)

	return mtp, nil
}

// NewMetricExporter inits a metric exporter.
func NewMetricExporter(cfg Config) (metric.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.MetricExporterConnectTimeout)
	defer cancel()

	exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(cfg.ExporterEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to init exporter: %w", err)
	}

	return exp, nil
}
