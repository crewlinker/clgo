// Package clwebserver implements serving of HTTP
package clwebserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/crewlinker/clgo/clconfig"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures the http server.
type Config struct {
	// BindAddrPort configures where the web server will listen for incoming tcp traffic
	BindAddrPort string `env:"BIND_ADDR_PORT" envDefault:"127.0.0.1:8282"`
	// HTTP read timeout, See: https://blog.cloudflare.com/exposing-go-on-the-internet/
	ReadTimeout time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
	// HTTP write timeout, See: https://blog.cloudflare.com/exposing-go-on-the-internet/
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"12s"`
	// HTTP idle timeout, See: https://blog.cloudflare.com/exposing-go-on-the-internet/
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" envDefault:"120s"`
}

// NewListener provides a tcp connection listener for the webserver.
func NewListener(cfg Config) (*net.TCPListener, error) {
	ap, err := netip.ParseAddrPort(cfg.BindAddrPort)
	if err != nil {
		return nil, fmt.Errorf("failed to parse addr/port: %w", err)
	}

	ln, err := net.ListenTCP("tcp", net.TCPAddrFromAddrPort(ap))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return ln, nil
}

// New inits the http server.
func New(cfg Config, logs *zap.Logger, h http.Handler, _ *net.TCPListener) *http.Server {
	return &http.Server{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      h,
		ErrorLog:     zap.NewStdLog(logs),
	}
}

// moduleName standardizes the module name.
const moduleName = "clwebserver"

// Provide dependencies.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the config
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// provide the listener
		fx.Provide(fx.Annotate(NewListener)),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// provide the server dependency
		fx.Provide(fx.Annotate(New,
			fx.OnStart(func(ctx context.Context, logs *zap.Logger, ln *net.TCPListener, s *http.Server) error {
				go s.Serve(ln) //nolint:errcheck

				logs.Info("http server started", zap.Stringer("addr", ln.Addr()))

				return nil
			}),
			fx.OnStop(func(ctx context.Context, logs *zap.Logger, s *http.Server) error {
				dl, hasdl := ctx.Deadline()
				logs.Info("shutting down http server", zap.Bool("has_dl", hasdl), zap.Duration("dl", time.Until(dl)))

				if err := s.Shutdown(ctx); err != nil {
					return fmt.Errorf("failed to shut down: %w", err)
				}

				return nil
			}))),
	)
}
