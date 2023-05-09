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
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "cllambda")
}

var _ = Describe("full app dependencies", func() {
	It("should wire up all dependencies as in actual deployment", func(ctx context.Context) {
		var hdlr *Handler
		Expect(fx.New(
			fx.Supply(env.Options{Environment: map[string]string{"CLZAP_LEVEL": "panic"}}),
			cllambda.Lambda[Input, Output](Prod()),
			fx.Populate(&hdlr),
		).Start(ctx)).To(Succeed())
		Expect(hdlr).ToNot(BeNil())
	})
})

type (
	// Input for testing.
	Input = struct{}
	// Output for testing.
	Output = struct{}
)

// Handler for testing.
type Handler struct{}

// New for testing.
func New(*zap.Logger) *Handler { return &Handler{} }

// Handle implementation.
func (Handler) Handle(context.Context, Input) (Output, error) {
	return Output{}, nil
}

func Prod() fx.Option {
	return fx.Module("lambda_test",
		fx.Provide(fx.Annotate(New)),
		fx.Provide(fx.Annotate(func(h *Handler) cllambda.Handler[Input, Output] { return h },
			fx.As(new(cllambda.Handler[Input, Output])))),
	)
}
