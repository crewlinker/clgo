// Package cllambda providees reusable fx code for building AWS Lambda infra
package cllambda

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/crewlinker/clgo/clzap"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Handler is a generic lambda handler interface
type Handler[I, O any] interface {
	Handle(context.Context, I) (O, error)
}

// Invoke provides the fx Invoke option to start the lambda right after the 'start' lifecycle event. It does
// not actually start the lambda unless the AWS_LAMBDA_RUNTIME_API is present. Which will be present on a
// real deployment but not during testing.
func Invoke[I, O any]() fx.Option {
	return fx.Invoke(fx.Annotate(func(lc fx.Lifecycle, logs *zap.Logger, h Handler[I, O]) {
		logs = logs.Named("cllambda.invoke")
		if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
			return // only add the lambda stat when we're actually executing inside the lambda, for testing
		}

		lc.Append(fx.Hook{OnStart: func(ctx context.Context) error {
			go lambda.StartWithOptions(h.Handle, lambda.WithEnableSIGTERM(func() {
				logs.Info("function container shutting down")
			}))
			return nil
		}})
	}))
}

// Lambda provides shared fx options (mostly modules) that may be used in any lambda handler so we can
// initialize them all in the same way (and even generate the main.go for all lambdas)
func Lambda[I, O any](o ...fx.Option) fx.Option {
	return fx.Options(append(o, []fx.Option{
		clzap.Fx(),     // log fx lines to zap
		clzap.Prod,     // provide logging to all handlers
		Invoke[I, O](), // invoke the lambda
	}...)...)
}
