package clauth

import (
	"bytes"
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/logging"
	"github.com/open-policy-agent/opa/sdk"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"

	_ "embed"
)

//go:embed opa.yml
var cfg []byte

// Authz provides authn and authz functionality. It includes a simple web server that
// serves our policy bundle on a random port on localhost.
type Authz struct {
	cfg  Config
	bsrv BundleServer
	logs *zap.Logger
	opa  *sdk.OPA
	opaw *zapio.Writer
}

// NewAuthz inits the auth service.
func NewAuthz(cfg Config, logs *zap.Logger, bsrv BundleServer) (a *Authz, err error) {
	return &Authz{bsrv: bsrv, cfg: cfg, logs: logs.Named("authz"), opaw: &zapio.Writer{Log: logs.Named("opa")}}, nil
}

// Start the auth service.
func (a *Authz) Start(ctx context.Context) (err error) {
	// pass ologs to zap
	ologs := logging.New()
	ologs.SetOutput(a.opaw)

	// resp, err := http.Get(a.bsrv.URL() + "/bundles/bundle.tar.gz")

	// // fmt.Println(resp.Status)

	// _ = resp

	// panic("b" + err.Error())

	a.opa, err = sdk.New(ctx, sdk.Options{
		ID:            a.cfg.AuthzOPASystemID,
		Config:        bytes.NewReader(bytes.ReplaceAll(cfg, []byte(`$SERVICE_URL$`), []byte(a.bsrv.URL()))),
		Logger:        ologs,
		ConsoleLogger: ologs,
	})
	if err != nil {
		return fmt.Errorf("failed to init opa: %w", err)
	}

	return nil
}

// Stop the auth service.
func (a *Authz) Stop(ctx context.Context) (err error) {
	if err := a.opaw.Close(); err != nil {
		return fmt.Errorf("failed to close zap writer: %w", err)
	}

	return
}

// IsAuthorized the user for a given setup.
func (a *Authz) IsAuthorized(ctx context.Context, inp any) (bool, error) {
	res, err := a.opa.Decision(ctx, sdk.DecisionOptions{
		Path:  "/rpc/allow",
		Input: inp,
	})
	if err != nil {
		return false, fmt.Errorf("failed to decide: %w", err)
	}

	allow, ok := res.Result.(bool)
	if !ok {
		return false, fmt.Errorf("decision did not return bool, but: %T", res.Result)
	}

	return allow, nil
}
