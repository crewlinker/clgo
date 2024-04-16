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

// Handler is a generic lambda handler interface.
type Handler[I, O any] interface {
	Handle(ctx context.Context, input I) (O, error)
}

// InvokeHandler provides the fx.Invoke option to start the lambda right after the 'start' lifecycle event. As
// opposed to [Invoke] it requires a single lambda.Handler dependency to be provided by the application.
func InvokeHandler() fx.Option {
	return fx.Invoke(fx.Annotate(func(fxlc fx.Lifecycle, logs *zap.Logger, hdlr lambda.Handler) {
		logs = logs.Named("lambda")

		if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
			return // only add the lambda stat when we're actually executing inside the lambda, for testing
		}

		fxlc.Append(fx.Hook{OnStart: func(context.Context) error {
			go lambda.StartWithOptions(hdlr,
				lambda.WithContext(baseContext(logs)), //nolint:contextcheck
				lambda.WithEnableSIGTERM(func() {
					logs.Info("received SIGTERM, shutting down")
				}))

			return nil
		}})
	}))
}

// Invoke provides the fx Invoke option to start the lambda right after the 'start' lifecycle event. It does
// not actually start the lambda unless the AWS_LAMBDA_RUNTIME_API is present. Which will be present on a
// real deployment but not during testing.
func Invoke[I, O any]() fx.Option {
	return fx.Invoke(fx.Annotate(func(fxlc fx.Lifecycle, logs *zap.Logger, hdlr Handler[I, O]) {
		logs = logs.Named("lambda")

		if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
			return // only add the lambda stat when we're actually executing inside the lambda, for testing
		}

		fxlc.Append(fx.Hook{OnStart: func(context.Context) error {
			go lambda.StartWithOptions(hdlr.Handle,
				lambda.WithContext(baseContext(logs)), //nolint:contextcheck
				lambda.WithEnableSIGTERM(func() {
					logs.Info("received SIGTERM, shutting down")
				}))

			return nil
		}})
	}))
}

// basecontext builds the root context for all lambda invocations.
func baseContext(logs *zap.Logger) context.Context {
	return clzap.WithLogger(context.Background(), logs)
}

// Lambda provides shared fx options (mostly modules) that may be used in any lambda handler so we can
// initialize them all in the same way (and even generate the main.go for all lambdas).
func Lambda[I, O any](o ...fx.Option) fx.Option {
	return fx.Options(append(o, []fx.Option{
		clzap.Fx(),      // log fx lines to zap
		clzap.Provide(), // provide logging to all handlers
		Invoke[I, O](),  // invoke the lambda
	}...)...)
}
