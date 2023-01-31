# clgo [![Test](https://github.com/crewlinker/clgo/actions/workflows/test.yaml/badge.svg)](https://github.com/crewlinker/clgo/actions/workflows/test.yaml)

Opinionated but Re-usable Go libraries for the Crewlinker platform

## usage

- Setup the development environment `mage -v dev`
- Run the full test suite: `mage -v test`
- To release a new version: `mage -v release v0.1.1`

## backlog

- [ ] MUST include a mechanism to provide isolated schemas to tests, using a "versioned" migration strategy
      in a migraiton directory
- [ ] MUST include tracing, and re-add the test for contextual postgres logging (from the old 'back' repo)
- [ ] SHOULD upgrade otel packages when otelsql package is supported
- [ ] SHOULD add the Atlasgo github integration for checking migrations
- [x] SHOULD Allow configuration of the postgres application name to diagnose connections
- [x] SHOULD allow iam authentication to a database
- [ ] COULD develop metric middleware for aws client so we can measure (average) latency (per service?)?

## Correct Otel dependencies

OTEL is still a bit of a moving target. So if the dependencies break after running `go mod tidy`. This is the
set of constraints that should make it right.

```
	go.opentelemetry.io/contrib/detectors/aws/ecs v1.12.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.37.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.37.0
	go.opentelemetry.io/contrib/propagators/aws v1.12.0
	go.opentelemetry.io/otel v1.11.2
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.34.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.11.2
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.11.2
	go.opentelemetry.io/otel/metric v0.34.0
	go.opentelemetry.io/otel/sdk v1.11.2
	go.opentelemetry.io/otel/sdk/metric v0.34.0
	go.opentelemetry.io/otel/trace v1.11.2
```
