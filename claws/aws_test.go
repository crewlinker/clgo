package claws_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/crewlinker/clgo/claws"
	"github.com/crewlinker/clgo/clotel"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestAwsclient(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "claws")
}

var _ = Describe("config without tracing", Serial, func() {
	var cfg aws.Config
	BeforeEach(func(ctx context.Context) {
		os.Setenv("AWS_REGION", "foo-bar-1")
		DeferCleanup(os.Unsetenv, "AWS_REGION")

		app := fx.New(
			fx.Populate(&cfg),
			fx.Decorate(claws.DynamoEndpointDecorator("http://foo:1")),
			clzap.Test(), claws.Prod())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should construct the config", func() {
		Expect(cfg.Region).To(Equal("foo-bar-1"))

		ep, err := cfg.EndpointResolverWithOptions.ResolveEndpoint(dynamodb.ServiceID, "eu-west-1")
		Expect(err).ToNot(HaveOccurred())
		Expect(ep.URL).To(Equal("http://foo:1"))
	})
})

var _ = Describe("config with tracing", Serial, func() {
	var cfg aws.Config
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&cfg), clzap.Test(), claws.Prod(), clotel.Test())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should have tracing options on client", func() {
		Expect(cfg.APIOptions).To(HaveLen(4))
	})
})
