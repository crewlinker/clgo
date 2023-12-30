// Package postgresresource implements a custom resource for Postgres
package postgresresource

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/crewlinker/clgo/claws"
	"github.com/crewlinker/clgo/clbuildinfo"
	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/cllambda"
	"github.com/crewlinker/clgo/clzap"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Define the handling input output as described in the documentation of the "mini-framework":
// https://docs.aws.amazon.com/cdk/api/v2/python/aws_cdk.custom_resources/README.html#handling-lifecycle-events-onevent
type (
	// Input into the handler.
	Input cfn.Event
	// Output into the handler.
	Output struct {
		// The allocated/assigned physical ID of the resource. If omitted for Create events, the event's RequestId
		// will be used. For Update, the current physical ID will be used. If a different value is returned,
		// CloudFormation will follow with a subsequent Delete for the previous ID (resource replacement).
		// For Delete, it will always return the current physical resource ID, and if the user returns a different one,
		// an error will occur.
		PhysicalResourceID string `json:"PhysicalResourceId"`
		// Resource attributes, which can later be retrieved through Fn::GetAtt on the custom resource object.
		Data map[string]any `json:"Data"`
		// Whether to mask the output of the custom resource when retrieved by using the Fn::GetAtt function.
		NoEcho bool `json:"NoEcho"`
	}
)

// Config configures the handler from env.
type Config struct{}

// SecretsManager provides an interface for reading AWS secrets.
type SecretsManager interface {
	GetSecretValue(
		ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options),
	) (*secretsmanager.GetSecretValueOutput, error)
}

// Handler handles custom resource requests for Fastly.
type Handler struct {
	cfg  Config
	logs *zap.Logger
	val  *validator.Validate
	smc  SecretsManager
}

// New inits the handler.
func New(
	cfg Config,
	logs *zap.Logger,

	smc SecretsManager,
) (*Handler, error) {
	val := validator.New()
	if err := val.RegisterValidation("resource_ident", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(fl.Field().String())
	}); err != nil {
		return nil, fmt.Errorf("failed to register validation: %w", err)
	}

	return &Handler{
		cfg:  cfg,
		logs: logs,
		val:  val,
		smc:  smc,
	}, nil
}

// Handle lambda input.
func (h Handler) Handle(ctx context.Context, in Input) (out Output, err error) {
	defer func() { h.logs.Info("handled", zap.Any("output", out)) }()
	h.logs.Info("handle", zap.Any("input", in))

	// handle the actual resources
	switch in.ResourceType {
	case "Custom::CrewlinkerPostgresTenant":
		var props TenantProperties
		if err = h.decodeValidateProps(in.ResourceProperties, &props); err != nil {
			return errorf("failed to validate properties: %w", in, err)
		}

		h.logs.Info("with properties", zap.Any("properties", props))

		switch in.RequestType {
		case cfn.RequestCreate:
			return h.handleTenantCreate(ctx, in, props)
		case cfn.RequestUpdate:
			var oldProps TenantProperties
			if err = h.decodeValidateProps(in.OldResourceProperties, &oldProps); err != nil {
				return errorf("failed to validate old properties: %w", in, err)
			}

			return h.handleTenantUpdate(ctx, in, props, oldProps)
		case cfn.RequestDelete:
			return h.handleTenantDelete(ctx, in, props)
		default:
			return errorf("unsupported request type", in)
		}
	default:
		return errorf("unsupported resource", in)
	}
}

// errorf returns a formatted error while referencing the resource type and request type.
func errorf(m string, in Input, v ...any) (Output, error) {
	return Output{PhysicalResourceID: in.PhysicalResourceID},
		fmt.Errorf("failed: '%s/%s': %w", in.ResourceType, in.RequestType, fmt.Errorf(m, v...)) //nolint:goerr113
}

// untilty function that decodes properties into a struct and validates it.
func (h Handler) decodeValidateProps(propm map[string]any, v any) (err error) {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.TextUnmarshallerHookFunc(),
		Metadata:   nil,
		Result:     v,
	})
	if err != nil {
		return fmt.Errorf("failed to init decoder: %w", err)
	}

	if err = dec.Decode(propm); err != nil {
		return fmt.Errorf("failed to decode properties: %w", err)
	}

	if err = h.val.Struct(v); err != nil {
		return fmt.Errorf("failed to validate properties: %w", err)
	}

	return
}

// moduleName for naming conventions.
const moduleName = "postgresresource"

// shared dependency setup.
func shared() fx.Option {
	return fx.Module("lambda/postgresresource",
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		fx.Provide(fx.Annotate(New)),
		fx.Provide(fx.Annotate(secretsmanager.NewFromConfig, fx.As(new(SecretsManager)))),
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		fx.Provide(fx.Annotate(func(h *Handler) cllambda.Handler[Input, Output] { return h },
			fx.As(new(cllambda.Handler[Input, Output])))),
		claws.Provide(),
	)
}

// TestProvide dependency setup.
func TestProvide() fx.Option {
	return fx.Options(
		clzap.TestProvide(),
		shared(),
	)
}

// Provide dependency setup.
func Provide(version string) fx.Option {
	return fx.Options(
		clbuildinfo.Provide(version),
		cllambda.Lambda[Input, Output](shared()),
	)
}
