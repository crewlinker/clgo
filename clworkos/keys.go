package clworkos

import (
	"encoding/base64"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// Keys hold our own private keys, and the WorkOS public keys.
type Keys struct {
	workos struct {
		public jwk.Set
	}
	signing struct {
		private jwk.Set
		public  jwk.Set
	}
	encrypt struct {
		private jwk.Set
		public  jwk.Set
	}
}

// NewKeys creates a new Keys instance.
func NewKeys(cfg Config) (*Keys, error) {
	keys := &Keys{}

	// decode and parse the signing keys
	{
		dec, err := base64.StdEncoding.DecodeString(cfg.PubPrivSigningKeySetB64JSON)
		if err != nil {
			return nil, fmt.Errorf("failed to decode signing keys as base64: %w", err)
		}

		keys.signing.private, err = jwk.Parse(dec)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signing encoded key set: %w", err)
		}

		keys.signing.public, err = jwk.PublicSetOf(keys.signing.private)
		if err != nil {
			return nil, fmt.Errorf("failed to get signing public key set: %w", err)
		}
	}

	// decode and parse the encrypt keys
	{
		dec, err := base64.StdEncoding.DecodeString(cfg.PubPrivEncryptKeySetB64JSON)
		if err != nil {
			return nil, fmt.Errorf("failed to decode encrypt keys as base64: %w", err)
		}

		keys.encrypt.private, err = jwk.Parse(dec)
		if err != nil {
			return nil, fmt.Errorf("failed to parse encoded encrypt key set: %w", err)
		}

		keys.encrypt.public, err = jwk.PublicSetOf(keys.encrypt.private)
		if err != nil {
			return nil, fmt.Errorf("failed to get encrypt public key set: %w", err)
		}
	}

	return keys, nil
}
