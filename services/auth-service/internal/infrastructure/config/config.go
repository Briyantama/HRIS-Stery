package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// LoadPEM loads RSA key material from env vars or file paths.
// Supports both plain PEM and base64-encoded PEM.
// Priority: {*}_FILE env var > {*}_BASE64 env var > {*} env var
func LoadPEM(privateEnv, publicEnv, privateBase64Env, publicBase64Env string) (privatePEM, publicPEM []byte, err error) {
	privatePEM, err = loadKeyMaterial(privateEnv, privateBase64Env)
	if err != nil {
		return nil, nil, fmt.Errorf("load private key: %w", err)
	}
	publicPEM, err = loadKeyMaterial(publicEnv, publicBase64Env)
	if err != nil {
		return nil, nil, fmt.Errorf("load public key: %w", err)
	}
	return privatePEM, publicPEM, nil
}

func loadKeyMaterial(inlineEnv, base64Env string) ([]byte, error) {
	// Try base64 env var first
	if b64 := os.Getenv(base64Env); b64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("decode base64: %w", err)
		}
		return decoded, nil
	}

	// Fall back to plain PEM env var
	inline := os.Getenv(inlineEnv)
	if inline == "" {
		return nil, fmt.Errorf("%s (or %s) is required", inlineEnv, base64Env)
	}
	return []byte(strings.ReplaceAll(inline, `\n`, "\n")), nil
}
