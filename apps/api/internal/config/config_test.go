package config_test

import (
	"strings"
	"testing"

	"github.com/Briyantama/HRIS-Stery/internal/config"
)

func validEnv() map[string]string {
	return map[string]string{
		"DATABASE_URL":           "postgres://u:dbsecret@db/x",
		"REDIS_ADDR":             "localhost:6379",
		"REDIS_PASSWORD":         "redissecret",
		"AUTH_PRIVATE_KEY":       "-----BEGIN-----\\nprivsecret\\n-----END-----",
		"AUTH_PUBLIC_KEY_BASE64": "cHVi",
		"S3_ENDPOINT":            "http://localhost:9000/",
		"S3_REGION":              "us-east-1",
		"S3_BUCKET":              "hris-documents",
		"S3_ACCESS_KEY_ID":       "key",
		"S3_SECRET_ACCESS_KEY":   "s3secret",
		"SMTP_PASSWORD":          "smtpsecret",
	}
}

func load(env map[string]string) (config.Config, error) {
	return config.Load(func(k string) string { return env[k] })
}

func TestLoad_defaultsAndKeys(t *testing.T) {
	cfg, err := load(validEnv())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8000" || cfg.SMTP.Port != 587 || !cfg.SMTP.TLSEnabled || !cfg.S3.PathStyle {
		t.Errorf("defaults wrong: %+v", cfg)
	}
	if cfg.S3.Endpoint != "http://localhost:9000" || cfg.S3.PublicEndpoint != cfg.S3.Endpoint {
		t.Errorf("endpoint = %q / %q", cfg.S3.Endpoint, cfg.S3.PublicEndpoint)
	}
	if !strings.Contains(string(cfg.AuthPrivateKeyPEM), "\nprivsecret\n") || string(cfg.AuthPublicKeyPEM) != "pub" {
		t.Errorf("keys not decoded: %q %q", cfg.AuthPrivateKeyPEM, cfg.AuthPublicKeyPEM)
	}
}

func TestLoad_reportsEveryProblemByName(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]string)
		want   string
	}{
		{"no database", func(e map[string]string) { delete(e, "DATABASE_URL") }, "DATABASE_URL is required"},
		{"no redis", func(e map[string]string) { delete(e, "REDIS_ADDR") }, "REDIS_ADDR is required"},
		{"no private key", func(e map[string]string) { delete(e, "AUTH_PRIVATE_KEY") }, "AUTH_PRIVATE_KEY or AUTH_PRIVATE_KEY_BASE64 is required"},
		{"no public key", func(e map[string]string) { delete(e, "AUTH_PUBLIC_KEY_BASE64") }, "AUTH_PUBLIC_KEY or AUTH_PUBLIC_KEY_BASE64 is required"},
		{"no s3 endpoint", func(e map[string]string) { delete(e, "S3_ENDPOINT") }, "S3_ENDPOINT is required"},
		{"bad port", func(e map[string]string) { e["PORT"] = "99999" }, "PORT must be a port number"},
		{"bad base64", func(e map[string]string) { e["AUTH_PUBLIC_KEY_BASE64"] = "!!" }, "AUTH_PUBLIC_KEY_BASE64 must be valid base64"},
		{"bad proxies", func(e map[string]string) { e["TRUSTED_PROXY_CIDRS"] = "nope" }, "TRUSTED_PROXY_CIDRS"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := validEnv()
			tc.mutate(env)
			_, err := load(env)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
			for _, secret := range []string{"dbsecret", "redissecret", "s3secret", "smtpsecret", "privsecret"} {
				if strings.Contains(err.Error(), secret) {
					t.Errorf("error leaks %q", secret)
				}
			}
		})
	}
}

func TestConfigString_redactsSecrets(t *testing.T) {
	cfg, err := load(validEnv())
	if err != nil {
		t.Fatal(err)
	}
	got := cfg.String()
	for _, secret := range []string{"dbsecret", "redissecret", "s3secret", "smtpsecret", "privsecret"} {
		if strings.Contains(got, secret) {
			t.Errorf("String() leaks %q: %s", secret, got)
		}
	}
}

func TestLoad_trustedProxies(t *testing.T) {
	env := validEnv()
	env["TRUSTED_PROXY_CIDRS"] = "10.0.0.0/8, 127.0.0.1"
	cfg, err := load(env)
	if err != nil || len(cfg.TrustedProxies) != 2 {
		t.Fatalf("proxies = %v, err = %v", cfg.TrustedProxies, err)
	}
}
