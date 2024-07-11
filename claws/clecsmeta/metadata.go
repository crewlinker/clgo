// Package clecsmeta provides metadata information about the ECS instance the code is running on.
package clecsmeta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "embed"
)

// Config configures this package.
type Config struct{}

// Metadata instance.
type Metadata struct {
	client *http.Client
	uri    string
	task   TaskMetadataV4
}

// ErrNoMetadataURI if the code provided no metadata uri.
var ErrNoMetadataURI = errors.New("no metadata uri provided")

const ecsContainerMetadataURIV4Env = "ECS_CONTAINER_METADATA_URI_V4"

// New inits the empty metadata.
func New(cfg Config, logs *zap.Logger, client *http.Client, ecsMetadataURI string) (*Metadata, error) {
	if ecsMetadataURI == "" {
		return nil, ErrNoMetadataURI
	}

	return &Metadata{uri: ecsMetadataURI, client: client}, nil
}

func (md Metadata) TaskV4() TaskMetadataV4 {
	return md.task
}

func (md *Metadata) Start(ctx context.Context) error {
	loc := lo.Must(url.JoinPath(md.uri, "task"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loc, nil)
	if err != nil {
		return fmt.Errorf("failed to init request: %w", err)
	}

	resp, err := md.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&md.task); err != nil {
		return fmt.Errorf("failed to decode response (%s): %w", resp.Status, err)
	}

	return nil
}

// moduleName for naming conventions.
const moduleName = "clecsmeta"

// Provide configures the DI for providng rpc.
func Provide() fx.Option {
	metadataURIFromEnv := os.Getenv(ecsContainerMetadataURIV4Env)

	return fx.Module(moduleName,
		fx.Supply(fx.Annotate(metadataURIFromEnv, fx.ResultTags(`name:"ecs_metadata_uri"`))),

		shared(),
	)
}

//go:embed testdata/metadatav4_response_task.json
var exampleResponse1 []byte

// TestProvide provides the metadata for testing environment.
func TestProvide() fx.Option {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(exampleResponse1)
	}))

	return fx.Options(
		fx.Supply(fx.Annotate(srv.URL, fx.ResultTags(`name:"ecs_metadata_uri"`))),
		shared(),
	)
}

// shared di options.
func shared() fx.Option {
	return fx.Options(
		// provide the middleware
		fx.Provide(fx.Annotate(New,
			fx.ParamTags(``, ``, ``, `name:"ecs_metadata_uri"`),
			fx.OnStart(func(c context.Context, m *Metadata) error { return m.Start(c) }))),
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
	)
}
