package clcedard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/crewlinker/clgo/clzap"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// BatchInput describes a set of inputs to all be checked at the same time.
type BatchInput struct {
	Policies string         `json:"policies"`
	Entities []any          `json:"entities"`
	Schema   map[string]any `json:"schema"`
	Items    []InputItem    `json:"items"`
}

// BatchOutput describes the output of batching a set of authorization checks.
type BatchOutput struct {
	Items []Output `json:"items"`
}

// InputItem represents the input for a single authorization
// check.
type InputItem struct {
	Principal string         `json:"principal"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource"`
	Context   map[string]any `json:"context"`
}

// Input for authorzation.
type Input struct {
	InputItem
	Policies string         `json:"policies"`
	Schema   map[string]any `json:"schema"`
	Entities []any          `json:"entities"`
}

// Output from authorization.
type Output struct {
	Decision      string   `json:"decision"`
	PolicyIDs     []string `json:"policy_ids"`
	ErrorMessages []string `json:"error_messages"`
}

// BatchIsAuthorized returns a list of booleans indicating whether each input is authorized.
// Any errors are gathered and returned as a single error.
func (c Client) BatchIsAuthorized(ctx context.Context, in *BatchInput) (ress []bool, err error) {
	out, err := c.BatchAuthorize(ctx, in)
	if err != nil {
		return nil, err
	}

	for _, item := range out.Items {
		res, e := outputRes(item)
		err = errors.Join(e)

		ress = append(ress, res)
	}

	return
}

// outputRes decides the bool result and optional error.
func outputRes(out Output) (bool, error) {
	res := false
	if out.Decision == "Allow" {
		res = true
	}

	if len(out.ErrorMessages) > 0 {
		return res, fmt.Errorf("authorization failed: %s", strings.Join(out.ErrorMessages, ", ")) //nolint:goerr113
	}

	return res, nil
}

// IsAuthorized returns true the authorization returned an Allow decision
// without errors. Otherwise, it returns false.
func (c *Client) IsAuthorized(ctx context.Context, in *Input) (bool, error) {
	out, err := c.Authorize(ctx, in)
	if err != nil {
		return false, err
	}

	return outputRes(*out)
}

// Authorize asks the cedard service authorizes the given input.
func (c *Client) BatchAuthorize(ctx context.Context, in *BatchInput) (out *BatchOutput, err error) {
	bof := backoff.NewExponentialBackOff()
	bof.MaxElapsedTime = c.cfg.BackoffMaxElapsedTime

	//nolint: wrapcheck
	return out, backoff.Retry(func() error {
		out = new(BatchOutput)

		err := c.doRequest(ctx, "/authorize_batch", in, out)
		if err != nil {
			clzap.Log(ctx, c.logs).Info("authorize batch failed", zap.Error(err))
		}

		return err
	}, backoff.WithContext(bof, ctx))
}

// Authorize asks the cedard service authorizes the given input.
func (c *Client) Authorize(ctx context.Context, in *Input) (out *Output, err error) {
	bof := backoff.NewExponentialBackOff()
	bof.MaxElapsedTime = c.cfg.BackoffMaxElapsedTime

	//nolint: wrapcheck
	return out, backoff.Retry(func() error {
		out = new(Output)

		err := c.doRequest(ctx, "/authorize", in, out)
		if err != nil {
			clzap.Log(ctx, c.logs).Info("authorize failed", zap.Error(err))
		}

		return err
	}, backoff.WithContext(bof, ctx))
}

// doRequest peforms the HTTP request with json (de)serialization.
func (c *Client) doRequest(ctx context.Context, path string, in, out any) error {
	input, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(input))
	if err != nil {
		return fmt.Errorf("failed to init HTTP request: %w", err)
	}

	signed, err := c.signedJWT(ctx)
	if err != nil {
		return fmt.Errorf("failed to sign jwt: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", signed)

	resp, err := c.htcl.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		//nolint: wrapcheck
		return backoff.Permanent(fmt.Errorf("client error: %q, HTTP body: %q", resp.Status, //nolint:goerr113
			string(lo.Must1(io.ReadAll(resp.Body)))))
	default:
		return fmt.Errorf("non-client error: %q, HTTP body: %q", resp.Status, //nolint:goerr113
			string(lo.Must1(io.ReadAll(resp.Body))))
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// signedJWT creates a signed JWT for the request.
func (c *Client) signedJWT(context.Context) (string, error) {
	tok, err := jwt.NewBuilder().Expiration(time.Now().Add(time.Minute)).Build()
	if err != nil {
		return "", fmt.Errorf("failed to build JWT: %w", err)
	}

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.HS256, []byte(c.cfg.JWTSigningSecret)))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return string(signed), nil
}
