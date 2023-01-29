package clhttp

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

// NewClient inits a http with optional tracing if available in the environment
func NewClient(tp trace.TracerProvider, pr propagation.TextMapPropagator) *http.Client {
	c := &http.Client{Transport: http.DefaultTransport}
	if tp != nil {
		c.Transport = otelhttp.NewTransport(http.DefaultTransport,
			otelhttp.WithPropagators(pr),
			otelhttp.WithTracerProvider(tp))
		// set default client and transport for any other service that uses it to perform logic
		http.DefaultClient = c
		http.DefaultTransport = http.DefaultClient.Transport
	}
	return c
}

// moduleName for naming conventions
const moduleName = "clhttp"

// Prod module to expose http services as dependencies
var Prod = fx.Module(moduleName,
	fx.Provide(fx.Annotate(NewClient, fx.ParamTags(`optional:"true"`, `optional:"true"`))),
)
