package clauthz

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/open-policy-agent/opa/sdk/test"
)

// readHeaderTimeout for our internal bundle server.
const readHeaderTimeout = time.Second * 3

// BundleServer interface defines the server that bundles will be fetched from. In case of
// a remove bundle server the Start and Stop can do nothing.
type BundleServer interface {
	URL() string
}

// MockBundles provides a bundle server that is easy to use for test.
type MockBundles struct{ *test.Server }

// MockBundle is a type that can be supplied to easily define policies in tests.
type MockBundle map[string]string

// NewMockBundles inits a bundle server.
func NewMockBundles(mb MockBundle) (bs *MockBundles, err error) {
	bs = &MockBundles{}

	bs.Server, err = test.NewServer(test.MockBundle("/bundles/bundle.tar.gz", mb))
	if err != nil {
		return nil, fmt.Errorf("failed to init mock bundle server: %w", err)
	}

	return bs, nil
}

func (bs MockBundles) Start(context.Context) error {
	return nil
}

func (bs MockBundles) Stop(context.Context) error {
	bs.Server.Stop()

	return nil
}

// BundleFS declares a type to carry the fs.FS that holds the OPA bundle as pre-build tar.gz.
type BundleFS struct{ fs.FS }

// FSBundles implements a bundle server that reads a tar.gz from the
// filesystem. Possibly through embedding it in the binary.
type FSBundles struct {
	svc *http.Server
	ln  net.Listener
}

// NewFSBundles inits the bundle server.
func NewFSBundles(bfs BundleFS) (*FSBundles, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &FSBundles{ln: ln, svc: &http.Server{
		ReadHeaderTimeout: readHeaderTimeout,
		Handler:           http.FileServer(http.FS(bfs)),
	}}, nil
}

// URL returns the url at which the bundles are served.
func (bs FSBundles) URL() string {
	return "http://" + bs.ln.Addr().String()
}

// Star the bundle server.
func (bs FSBundles) Start(context.Context) error {
	go bs.svc.Serve(bs.ln) //nolint:errcheck

	return nil
}

// Stop the bundle server.
func (bs FSBundles) Stop(ctx context.Context) error {
	if err := bs.svc.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown bundle server: %w", err)
	}

	return nil
}
