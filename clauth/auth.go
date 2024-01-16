// Package auth provides authentication (Auth) and authorization (Authz).
package clauth

import (
	"context"
	"encoding/base64"
	"io/fs"
	"strings"
	"time"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures the package.
type Config struct {
	// id for the system that is unning OPA.
	AuthzOPASystemID string `env:"AUTHZ_OPA_SYSTEM_ID" envDefault:"auth"`
	// PrivateSigningKeys will hold private keys for signing JWTs
	AuthnPubPrivKeySetB64JSON string `env:"AUTHN_PUB_PRIV_KEY_SET_B64_JSON"`
	// AuthnDefaultSignKeyID defines the default key id used for signing
	AuthnDefaultSignKeyID string `env:"AUTHN_DEFAULT_SIGN_KEY_ID" envDefault:"key1"`
}

// moduleName for consistent component naming.
const moduleName = "clauth"

// Provide the auth components as an fx dependency.
func Provide() fx.Option {
	return fx.Options(
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// provide the webos webhooks client
		fx.Provide(fx.Annotate(NewAuthz,
			fx.OnStart(func(ctx context.Context, a *Authz) error { return a.Start(ctx) }),
			fx.OnStop(func(ctx context.Context, a *Authz) error { return a.Stop(ctx) }),
		)),
		// provide authentication service
		fx.Provide(NewAuthn),
		// provide a wall clock
		fx.Supply(fx.Annotate(jwt.ClockFunc(time.Now), fx.As(new(jwt.Clock)))),
	)
}

// BundleProvide provides a bundle server.
func BundleProvide(bfs fs.FS) fx.Option {
	return fx.Options(
		fx.Supply(BundleFS{FS: bfs}),
		fx.Provide(fx.Annotate(NewFSBundles,
			fx.OnStart(func(ctx context.Context, a *FSBundles) error { return a.Start(ctx) }),
			fx.OnStop(func(ctx context.Context, a *FSBundles) error { return a.Stop(ctx) }),
		)),
		fx.Provide(func(bs *FSBundles) BundleServer { return bs }),
	)
}

// TestProvide provides authn authz dependencies that are easy to use in
// tests.
func TestProvide(policies map[string]string) fx.Option {
	return fx.Options(
		Provide(),

		// supply the policies
		fx.Supply(MockBundle(policies)),

		// provide a bundle server that is easy to use in tests.
		fx.Provide(fx.Annotate(NewMockBundles,
			fx.OnStart(func(ctx context.Context, a *MockBundles) error { return a.Start(ctx) }),
			fx.OnStop(func(ctx context.Context, a *MockBundles) error { return a.Stop(ctx) }),
		)),

		// provide as bundle server
		fx.Provide(func(bs *MockBundles) BundleServer { return bs }),

		// set the configuration to have a signing key just for testing
		fx.Decorate(func(cfg Config) Config {
			//nolint:lll
			cfg.AuthnPubPrivKeySetB64JSON = base64.URLEncoding.EncodeToString([]byte(`{
				"keys": [
					{
						"p": "zP5Ts7sL893m0RkhhtrtYnGu8fuTUnfo1vvIcJ_UN3CDsWcycsOZJ_aRQSoJ_c6REl9xLPl_enV4B_VPmsF-wAhFiqZyRWeriP5PS70sAKEwlDJshfU9hjs_dTLJlOL_SU-k0BQbtkUJyX-dd3KWZ_AZWuqlfE0kWI-yiWHGic8",
						"kty": "RSA",
						"q": "sVQGOlqDuKtvivOWM43JoFx996yOBdVTgY-X_Nh022tK8cCWNXkExENJoqIowtmpRvYTRZhh0cySR-2tsfC5LT-tI7PFIHykyi5HTkuggXBShtBlLhjCn8LlBbaRolhEksfqji79Fd9tjFQ9uAJp8TQbOdfI6JPyjiGEIPnjvlc",
						"d": "fAWemd9jC-MLQuT-HlJHKEXQUIi6VASADZe5C8bYDHbraG_VXt7eKM9x-ov9gHh-6xuWl9sbkeuLA5kTFDsYh8sZnT88VeXs2ugy-B0vKNm_qBfqKAvKEEpl5cWbZiFsIEJ8PHS06kC7iQKxFHc0b1tRyZqFPFpbG9Cht99uUJXNaO3SuUmPqRHBDTjTesDbkNFiFxaRVveJhzV99002M_E8rXLc0l0I0qCxtJyeTNZF5oWXoRpy7QJ3Zuz2kSOCQoZQfIb-IgAu0AwTfv8NVvoRc0hMwjILc2F63mH1h4Ypsfx9m5dxjm6OtklMn-CNSg6aE2ItaULhTCK095CCFQ",
						"e": "AQAB",
						"use": "sig",
						"kid": "key1",
						"qi": "sW_PVQ27ElK6vbilzvTU2wDyl1fUUmB_TmKg5eQGBbE3UGNxT3EGsNgpc388ROR8M4-rUiEBJ_-vVck_POOjU3tZNIzmiZP3Tw745RvDkzNG2k-R7T7emxcvH49zbBgYXDb1XyWOIOVBFqwE4HZw92CVZCmeEWw5fISOkgzGy1w",
						"dp": "a4Zb0UKjml8i2zsbYuki6yhGY5daRz-uWlXnZWvwnMPf0AYZaClBBL1Io62xX_giEEkPzE9yloFXXJVIFBy6p2-vSnLULaObTlhWr5uioRHrsVBhrEJe6zHYr1jcc8Q9s-6avKpPfuPnplHR_v2T9yDxq8a41uJ_1hRJydYHlfE",
						"alg": "RS256",
						"dq": "VgDDeH-3zNPQqFqVaXGF7XGOYpXc17Vr57Vl6Gpu2pBB69gUweBs0Gc2CluNW1tHfzQPirxqDN-jvqDmkhuHJAvzBBLHM4dgQPKLAM0rDjwUum_N8rptgiB7BPdT0KHwuCOffdAKTRZswheFS35YNXSpE7e1KB_BDu_wbjHkI8c",
						"n": "jf8gT2tctBUx6k91sef47WCL7CpbUbMV20yVFkRmG-A1zt0cTvbeHukwsvSJpGOWPKEHasZ4UaSHAasneVtuVy6F26lntGj3_B9sBXT4YkCAhli2yt04Ieb5R2OwUestk3NbUxGPWwO7wp0s8o_cUoWqMr6K3Ecbgo8qcKmwhujNP3qpTzTKr-wA_2Tj-jSU6PycfAplGE0B6hwCyCKQlcuEtyRWIWgajHvCuMR5Rsh7iG_aamz1p1AvV-bniGOi-mQ_KAxOQpUs59Nte0Ces0g-JalwccOC2VN4zfHX0js2FImelVasMQ-K1lagHZPHaL9nB6w8gnTc0mg4Hk13WQ"
					}
				]
			}`))

			return cfg
		}),
	)
}
