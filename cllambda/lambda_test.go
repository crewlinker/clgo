package cllambda_test

import (
	"context"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/crewlinker/clgo/cllambda"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func TestCLLambda(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/cllambda")
}

var _ = Describe("full app dependencies", func() {
	It("should wire up all dependencies as in actual deployment", func(ctx context.Context) {
		var h *Handler
		Expect(fx.New(
			fx.Supply(env.Options{Environment: map[string]string{"CLZAP_LEVEL": "panic"}}),
			cllambda.Lambda[Input, Output](Prod),
			fx.Populate(&h),
		).Start(ctx)).To(Succeed())
		Expect(h).ToNot(BeNil())
	})
})

type (
	// Input for testing
	Input = struct{}
	// Output for testing
	Output = struct{}
)

// Handler for testing
type Handler struct{}

// New for testing
func New(logs *zap.Logger) *Handler { return &Handler{} }

// Handle implementation
func (Handler) Handle(ctx context.Context, in Input) (out Output, err error) {
	return
}

var Prod = fx.Module("lambda_test",
	fx.Provide(fx.Annotate(New)),
	fx.Provide(fx.Annotate(func(h *Handler) cllambda.Handler[Input, Output] { return h },
		fx.As(new(cllambda.Handler[Input, Output])))),
)
