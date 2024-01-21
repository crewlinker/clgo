package clcdkcr

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

// Output as is expected to be returned from a
// custom resource handler implementation.
type Output struct {
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

// Result of handling.
type Result struct {
	// Visited returns which handlers where visited.
	Visited []string
	// Handled is set to the resource that handled it.
	Handled string
	// Err holds any error while handling the event
	Err error
	// Output as returned from the lambda handler
	Output
}

// Handler for CRUD operations.
type Handler[I, O any] interface {
	Type() string
	Create(ctx context.Context, ev cfn.Event, in I) (string, O, bool, error)
	Update(ctx context.Context, ev cfn.Event, in I, inOld I) (string, O, bool, error)
	Delete(ctx context.Context, ev cfn.Event, in I) (O, bool, error)
}

// handle handles custom resource events for 1 resource type.
func handle[
	I, O any,
](
	ctx context.Context,
	logs *zap.Logger,
	val *validator.Validate,
	evn cfn.Event,
	res *Result,
	hdl Handler[I, O],
) error {
	logs.Info("handling resource event", zap.Any("event", evn))

	var inProps I
	if err := decodeValidateProps(val, evn.ResourceProperties, &inProps); err != nil {
		return fmt.Errorf("failed to decode/validate (new) resource properties: %w", err)
	}

	logs.Info("decoded (new) resource properties", zap.Any("props", inProps))

	var (
		outProps O
		err      error
		prid     string
		noEcho   bool
	)

	switch evn.RequestType {
	case cfn.RequestCreate:
		prid, outProps, noEcho, err = hdl.Create(ctx, evn, inProps)
	case cfn.RequestUpdate:
		var inOldProps I
		if err := decodeValidateProps(val, evn.OldResourceProperties, &inOldProps); err != nil {
			return fmt.Errorf("failed to decode/validate old input properties: %w", err)
		}

		logs.Info("decoded old resource properties", zap.Any("props", inOldProps))

		prid, outProps, noEcho, err = hdl.Update(ctx, evn, inProps, inOldProps)
	case cfn.RequestDelete:
		// for delete the returned physical resource id must always be the same or cloudformation
		// will error. We don't event allow delete to change that.
		prid = evn.PhysicalResourceID

		outProps, noEcho, err = hdl.Delete(ctx, evn, inProps)
	default:
		err = UnsupportedRequestTypeError{evn.RequestType}
	}

	if err != nil {
		return fmt.Errorf("failed to handle %s: %w", evn.RequestType, err)
	}

	logs.Info("encoding output properties", zap.Any("props", outProps))

	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.TextUnmarshallerHookFunc(),
		Metadata:   nil,
		Result:     &res.Output.Data,
	})
	if err != nil {
		return fmt.Errorf("failed to setup output props decoder: %w", err)
	}

	res.Output.PhysicalResourceID = prid
	res.Output.NoEcho = noEcho

	if err := dec.Decode(outProps); err != nil {
		return fmt.Errorf("failed to decode output props: %w", err)
	}

	logs.Info("output ready", zap.Any("output", res.Output))

	return nil
}

// UnsupportedResourceTypeError is returned when the resource is not supported.
type UnsupportedResourceTypeError struct{ rt string }

func (e UnsupportedResourceTypeError) Error() string {
	return fmt.Sprintf("unsupported resource type: %s", e.rt)
}

// UnsupportedRequestTypeError is returned when the request is not supported.
type UnsupportedRequestTypeError struct{ rt cfn.RequestType }

func (e UnsupportedRequestTypeError) Error() string {
	return fmt.Sprintf("unsupported request type: %s", e.rt)
}

// visitHandle handles custom resource events for 1 resource type.
func visitHandle[
	I, O any,
](
	ctx context.Context,
	logs *zap.Logger,
	val *validator.Validate,
	ev cfn.Event,
	res *Result,
	h Handler[I, O],
) {
	res.Visited = append(res.Visited, h.Type())

	switch {
	case res.Handled != "" || res.Err != nil:
		return // already handled
	case h.Type() != ev.ResourceType:
		return // not supported
	}

	res.Err = handle(ctx, logs.Named(h.Type()), val, ev, res, h)
	res.Handled = h.Type()
}

// checkError will check if res indicates it was.
func checkError(ev cfn.Event, res Result) (Result, error) {
	if res.Err != nil {
		return res, res.Err
	}

	if res.Handled == "" {
		return res, UnsupportedResourceTypeError{ev.ResourceType}
	}

	return res, nil
}

// untilty function that decodes properties into a struct and validates it.
func decodeValidateProps(val *validator.Validate, propm map[string]any, v any) (err error) {
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

	if err = val.Struct(v); err != nil {
		return fmt.Errorf("failed to validate properties: %w", err)
	}

	return
}

// Handle1 handles custom resource events for 1 resource type.
func Handle1[
	I1, O1 any,
](
	ctx context.Context,
	logs *zap.Logger,
	val *validator.Validate,
	ev cfn.Event,
	h1 Handler[I1, O1],
) (res Result, err error) {
	visitHandle(ctx, logs, val, ev, &res, h1)

	return checkError(ev, res)
}

// Handle2 handles custom resource events for 2 resource types.
func Handle2[
	I1, O1 any,
	I2, O2 any,
](
	ctx context.Context,
	logs *zap.Logger,
	val *validator.Validate,
	ev cfn.Event,
	h1 Handler[I1, O1],
	h2 Handler[I2, O2],
) (res Result, err error) {
	visitHandle(ctx, logs, val, ev, &res, h1)
	visitHandle(ctx, logs, val, ev, &res, h2)

	return checkError(ev, res)
}

// Handle3 handles custom resource events for 3 resource types.
func Handle3[
	I1, O1 any,
	I2, O2 any,
	I3, O3 any,
](
	ctx context.Context,
	logs *zap.Logger,
	val *validator.Validate,
	ev cfn.Event,
	h1 Handler[I1, O1],
	h2 Handler[I2, O2],
	h3 Handler[I3, O3],
) (res Result, err error) {
	visitHandle(ctx, logs, val, ev, &res, h1)
	visitHandle(ctx, logs, val, ev, &res, h2)
	visitHandle(ctx, logs, val, ev, &res, h3)

	return checkError(ev, res)
}

// Handle4 handles custom resource events for 4 resource types.
func Handle4[
	I1, O1 any,
	I2, O2 any,
	I3, O3 any,
	I4, O4 any,
](
	ctx context.Context,
	logs *zap.Logger,
	val *validator.Validate,
	ev cfn.Event,
	h1 Handler[I1, O1],
	h2 Handler[I2, O2],
	h3 Handler[I3, O3],
	h4 Handler[I4, O4],
) (res Result, err error) {
	visitHandle(ctx, logs, val, ev, &res, h1)
	visitHandle(ctx, logs, val, ev, &res, h2)
	visitHandle(ctx, logs, val, ev, &res, h3)
	visitHandle(ctx, logs, val, ev, &res, h4)

	return checkError(ev, res)
}

// Handle5 handles custom resource events for 5 resource types.
func Handle5[
	I1, O1 any,
	I2, O2 any,
	I3, O3 any,
	I4, O4 any,
	I5, O5 any,
](
	ctx context.Context,
	logs *zap.Logger,
	val *validator.Validate,
	ev cfn.Event,
	h1 Handler[I1, O1],
	h2 Handler[I2, O2],
	h3 Handler[I3, O3],
	h4 Handler[I4, O4],
	h5 Handler[I5, O5],
) (res Result, err error) {
	visitHandle(ctx, logs, val, ev, &res, h1)
	visitHandle(ctx, logs, val, ev, &res, h2)
	visitHandle(ctx, logs, val, ev, &res, h3)
	visitHandle(ctx, logs, val, ev, &res, h4)
	visitHandle(ctx, logs, val, ev, &res, h5)

	return checkError(ev, res)
}

// Handle6 handles custom resource events for 6 resource types.
func Handle6[
	I1, O1 any,
	I2, O2 any,
	I3, O3 any,
	I4, O4 any,
	I5, O5 any,
	I6, O6 any,
](
	ctx context.Context,
	logs *zap.Logger,
	val *validator.Validate,
	ev cfn.Event,
	h1 Handler[I1, O1],
	h2 Handler[I2, O2],
	h3 Handler[I3, O3],
	h4 Handler[I4, O4],
	h5 Handler[I5, O5],
	h6 Handler[I6, O6],
) (res Result, err error) {
	visitHandle(ctx, logs, val, ev, &res, h1)
	visitHandle(ctx, logs, val, ev, &res, h2)
	visitHandle(ctx, logs, val, ev, &res, h3)
	visitHandle(ctx, logs, val, ev, &res, h4)
	visitHandle(ctx, logs, val, ev, &res, h5)
	visitHandle(ctx, logs, val, ev, &res, h6)

	return checkError(ev, res)
}
