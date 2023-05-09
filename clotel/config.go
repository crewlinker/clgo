package clotel

import "time"

// Config configures the code in this package.
type Config struct {
	// DetectorDetectTimeout bound the time it may take to init a trace provider
	DetectorDetectTimeout time.Duration `env:"DETECTOR_DETECT_TIMEOUT" envDefault:"1s"`
	// ExporterTimeout overwrites the timeout for exporting spans. This can be useful in tests to speed
	// them up
	ExporterTimeout time.Duration `env:"EXPORTER_TIMEOUT" envDefault:"10s"`
	// ExporterEndpoint configures where otel span exporter will send data to
	ExporterEndpoint string `env:"EXPORTER_ENDPOINT" envDefault:"localhost:4317"`
	// MetricExporterConnectTimeout configures how long we'll wait for het metric exporter to connect to the collector
	MetricExporterConnectTimeout time.Duration `env:"METRIC_EXPORTER_CONNECT_TIMEOUT" envDefault:"1s"`
}
