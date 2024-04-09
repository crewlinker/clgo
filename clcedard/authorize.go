package clcedard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// Input for authorzation.
type Input struct {
	Principal string         `json:"principal"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource"`
	Policies  string         `json:"policies"`
	Context   map[string]any `json:"context"`
	Schema    map[string]any `json:"schema"`
	Entities  []any          `json:"entities"`
}

// Output from authorization.
type Output struct {
	Decision      string   `json:"decision"`
	PolicyIds     []string `json:"policy_ids"`
	ErrorMessages []string `json:"error_messages"`
}

// IsAuthorized returns true the authorization returned an Allow decision
// without errors. Otherwise, it returns false.
func (c *Client) IsAuthorized(ctx context.Context, in *Input) (bool, error) {
	out, err := c.Authorize(ctx, in)
	if err != nil {
		return false, err
	}

	res := false
	if out.Decision == "Allow" {
		res = true
	}

	if len(out.ErrorMessages) > 0 {
		return res, fmt.Errorf("authorization failed: %s", strings.Join(out.ErrorMessages, ", "))
	}

	return res, nil
}

// Authorize asks the cedard service authorizes the given input.
func (c *Client) Authorize(ctx context.Context, in *Input) (out *Output, err error) {
	bof := backoff.NewExponentialBackOff()
	bof.MaxElapsedTime = c.cfg.BackoffMaxElapsedTime

	//nolint: wrapcheck
	return out, backoff.Retry(func() error {
		out, err = c.authorize(ctx, in)
		c.logs.Info("authorize failed, retrying", zap.Error(err))

		return err
	}, backoff.WithContext(bof, ctx))
}

// authorize performs the actual authorization request.
func (c *Client) authorize(ctx context.Context, in *Input) (*Output, error) {
	input, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/authorize", bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to init HTTP request: %w", err)
	}

	signed, err := c.signedJWT(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign jwt: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", signed)

	resp, err := c.htcl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		//nolint: wrapcheck
		return nil, backoff.Permanent(fmt.Errorf("client error: %q, HTTP body: %q", resp.Status,
			string(lo.Must1(io.ReadAll(resp.Body)))))
	default:
		return nil, fmt.Errorf("non-client error: %q, HTTP body: %q", resp.Status,
			string(lo.Must1(io.ReadAll(resp.Body))))
	}

	var out Output
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &out, nil
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
