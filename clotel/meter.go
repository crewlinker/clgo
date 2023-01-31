package clotel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// NewMeterProvider initiales otel provider for metrics throughout the application
func NewMeterProvider(cfg Config, det resource.Detector, mr metric.Reader) (*metric.MeterProvider, error) {

	// detect the resource with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DetectorDetectTimeout)
	defer cancel()
	res, err := det.Detect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(mr),
		metric.WithResource(res))

	// set globally in case libraries don't allow injecting
	global.SetMeterProvider(mp)
	return mp, nil
}

// NewMetricExporter inits a metric exporter
func NewMetricExporter(cfg Config) (metric.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.MetricExporterConnectTimeout)
	defer cancel()

	return otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(cfg.ExporterEndpoint))
}
