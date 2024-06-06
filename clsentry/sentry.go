// Package clsentry provides Sentry error reporting.
package clsentry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/crewlinker/clgo/clbuildinfo"
	"github.com/crewlinker/clgo/clconfig"
	sentry "github.com/getsentry/sentry-go"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

// Config configures.
type Config struct {
	// Sentry DSN to send data to.
	DSN string `env:"DSN"`
	// TracesSampleRate is the rate at which traces are captures for Sentry.
	TracesSampleRate float64 `env:"TRACES_SAMPLE_RATE" envDefault:"1.0"`
	// DefaultFlushTimeout is the default timeout for flushing to Sentry.
	DefaultFlushTimeout time.Duration `env:"DEFAULT_FLUSH_TIMEOUT" envDefault:"2s"`
	// AttachStacktrace is whether to attach stacktrace to pure capture message calls.
	AttachStacktrace bool `env:"ATTACH_STACKTRACE" envDefault:"true"`
	// If set, will add this environment to the Sentry scope.
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

// StringFromEventID converts a sentry event ID to a string pointer.
func StringFromEventID(id *sentry.EventID) *string {
	if id == nil {
		return nil
	}

	s := string(*id)

	return &s
}

// BeforeSendFunc allows events to be modified before they are sent to Sentry.
type BeforeSendFunc func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event

// newOptions creates a new sentry client options.
func newOptions(cfg Config, binfo clbuildinfo.Info, beforeSend BeforeSendFunc) sentry.ClientOptions {
	return sentry.ClientOptions{
		Dsn:              cfg.DSN,
		TracesSampleRate: cfg.TracesSampleRate,
		AttachStacktrace: cfg.AttachStacktrace,

		Release:     binfo.Version(),
		Environment: cfg.Environment,

		BeforeSend: beforeSend,
	}
}

// newHub inits a new sentry hub.
func newHub(client *sentry.Client, scope *sentry.Scope) *sentry.Hub {
	return sentry.NewHub(client, scope)
}

// ErrFlushFailed is returned when the flush fails during fx shutdown.
var ErrFlushFailed = errors.New("failed to flush sentry hub")

// moduleName for naming conventions.
const moduleName = "clsentry"

// Provide configures the DI for providng rpc.
func Provide() fx.Option {
	var cfg Config

	return fx.Module(moduleName,
		fx.Provide(sentry.NewClient, sentry.NewScope),
		fx.Populate(&cfg),
		fx.Provide(fx.Annotate(newOptions, fx.ParamTags(``, ``, `optional:"true"`))),

		// provide the sentry hub and flush on shutdown.
		fx.Provide(fx.Annotate(newHub, fx.OnStop(func(ctx context.Context, hub *sentry.Hub) error {
			dl, ok := ctx.Deadline()
			if !ok {
				dl = time.Now().Add(cfg.DefaultFlushTimeout)
			}

			if !hub.Flush(time.Until(dl)) {
				return fmt.Errorf("%w", ErrFlushFailed)
			}

			return nil
		}))),

		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
	)
}

type testServer struct{ *httptest.Server }

// ObservedEvents keeps observe logs.
type ObservedEvents struct {
	evs []sentry.Event
	mu  sync.RWMutex
}

// Events returns a copy of all the ovserved events. It is safe to call from multiple goroutines.
func (o *ObservedEvents) Events() (evs []sentry.Event) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	evs = make([]sentry.Event, len(o.evs))
	copy(evs, o.evs)

	return evs
}

// TestProvide provides dependencies in a way that makes testing Sentry integration easier.
func TestProvide() fx.Option {
	return fx.Options(
		fx.Provide(func() *ObservedEvents {
			return &ObservedEvents{}
		}),

		fx.Provide(fx.Annotate(func(obs *ObservedEvents) *testServer {
			return &testServer{httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				parts, event := bytes.SplitN(lo.Must1(io.ReadAll(r.Body)), []byte("\n"), 3), sentry.Event{} //nolint:gomnd

				lo.Must0(json.Unmarshal(parts[2], &event))

				obs.mu.Lock()
				obs.evs = append(obs.evs, event)
				obs.mu.Unlock()
			}))}
		}, fx.OnStop(func(ts *testServer) { ts.Close() }))),

		fx.Decorate(func(c Config, ts *testServer) Config {
			loc, _ := url.Parse(ts.URL)
			loc.User = url.UserPassword("someuser", "")
			c.DSN = loc.String() + "/someproject"

			return c
		}),

		Provide(),
	)
}
