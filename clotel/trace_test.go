package clotel_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/crewlinker/clgo/clotel"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

func TestClotel(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clotel")
}

var _ = Describe("otel tracing", func() {
	var trp *sdktrace.TracerProvider
	var tpi trace.TracerProvider
	var tobs *tracetest.InMemoryExporter
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&trp, &tpi, &tobs), clotel.TestProvide(), clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should provide tracing", func(ctx context.Context) {
		_, span := trp.Tracer("test").Start(ctx, "my-span")
		span.End()

		Expect(trp.ForceFlush(ctx)).To(Succeed())
		spans := tobs.GetSpans().Snapshots()

		Expect(spans).To(HaveLen(1))
		Expect(spans[0].Name()).To(Equal("my-span"))
	})
})

type testDetector struct{}

func (testDetector) Detect(ctx context.Context) (*resource.Resource, error) {
	res, err := resource.Detect(ctx,
		resource.StringDetector(semconv.SchemaURL, semconv.AWSLogGroupARNsKey, func() (string, error) {
			return "xyz", nil
		}), resource.StringDetector(semconv.SchemaURL, semconv.AWSLogGroupNamesKey, func() (string, error) {
			return "zyx", nil
		}))
	if err != nil {
		return nil, fmt.Errorf("failed to detect: %w", err)
	}

	return res, nil
}

var _ = Describe("extra ecs detector", func() {
	var det resource.Detector
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&det),
			fx.Supply(fx.Annotate(testDetector{}, fx.As(new(resource.Detector)))),
			fx.Decorate(clotel.WithExtraEcsAttributes),
			clzap.TestProvide(),
		)
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should overwrite cloudwatch log arns/names", func(ctx context.Context) {
		res, err := det.Detect(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(fmt.Sprint(res)).To(Equal(`aws.log.group.arns=[xyz],aws.log.group.names=[zyx]`))
	})
})
