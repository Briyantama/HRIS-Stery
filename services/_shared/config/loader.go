package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Option is a functional option for configuring the Viper instance.
type Option func(*viper.Viper)

// WithDefault sets a default value for a config key.
func WithDefault(key string, value any) Option {
	return func(v *viper.Viper) {
		v.SetDefault(key, value)
	}
}

// WithConfigFile specifies a custom config file path.
func WithConfigFile(path string) Option {
	return func(v *viper.Viper) {
		v.SetConfigFile(path)
	}
}

// Load reads configuration from .env file and environment variables.
// The .env file is optional; if absent, configuration comes from env vars only.
// Environment variables always override file values (12-factor App).
// dst must be a pointer to a struct with mapstructure tags.
func Load(dst any, opts ...Option) error {
	v := viper.New()

	// Apply functional options
	for _, opt := range opts {
		opt(v)
	}

	// Set config file defaults if not overridden
	if v.ConfigFileUsed() == "" {
		v.SetConfigFile(".env")
	}
	v.SetConfigType("dotenv")

	// Enable automatic env var binding with . → _ replacement
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set global defaults
	setDefaults(v)

	// Bind explicit env var names (for non-standard mappings)
	bindEnvs(v)

	// Apply additional options (this allows overriding defaults)
	for _, opt := range opts {
		opt(v)
	}

	// Read config file (if it exists)
	if err := v.ReadInConfig(); err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			// .env file not found — OK, use env vars only
		} else {
			// Real parse error
			return fmt.Errorf("parse config file: %w", err)
		}
	}

	// Unmarshal into destination struct
	decoderConfig := &mapstructure.DecoderConfig{
		TagName: "mapstructure",
		Result:  dst,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return fmt.Errorf("create decoder: %w", err)
	}

	if err := decoder.Decode(v.AllSettings()); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	return nil
}

// MustLoad calls Load and panics if an error occurs.
// Intended for use in main() where startup failure is fatal.
func MustLoad(dst any, opts ...Option) {
	if err := Load(dst, opts...); err != nil {
		panic(fmt.Sprintf("load config: %v", err))
	}
}

// setDefaults sets Viper defaults for well-known config keys.
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "development")
	v.SetDefault("observability.log_level", "info")
	v.SetDefault("observability.enabled", false)
	v.SetDefault("server.metrics_port", "8081")
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", "6379")
	v.SetDefault("jwt.expiration_hours", 1)
	v.SetDefault("jwt.refresh_expiry_days", 7)
	v.SetDefault("smtp.port", 587)
	v.SetDefault("smtp.tls_enabled", false)
}

// bindEnvs binds non-standard env var names to Viper keys.
// This ensures backward compatibility with existing .env.example files.
func bindEnvs(v *viper.Viper) {
	// server.grpc_port → GRPC_PORT (not SERVER_GRPC_PORT)
	_ = v.BindEnv("server.grpc_port", "GRPC_PORT")

	// observability.* → non-standard names
	_ = v.BindEnv("observability.log_level", "LOG_LEVEL")
	_ = v.BindEnv("observability.otlp_endpoint", "OTEL_EXPORTER_OTLP_ENDPOINT")
	_ = v.BindEnv("observability.enabled", "OTEL_ENABLED")

	// jwt.* → AUTH_* names
	_ = v.BindEnv("jwt.private_key", "AUTH_PRIVATE_KEY")
	_ = v.BindEnv("jwt.public_key", "AUTH_PUBLIC_KEY")
	_ = v.BindEnv("jwt.private_key_b64", "AUTH_PRIVATE_KEY_BASE64")
	_ = v.BindEnv("jwt.public_key_b64", "AUTH_PUBLIC_KEY_BASE64")

	// jwt.* → JWT_* names for TTL config
	_ = v.BindEnv("jwt.expiration_hours", "JWT_EXPIRATION_HOURS")
	_ = v.BindEnv("jwt.refresh_expiry_days", "REFRESH_TOKEN_EXPIRATION_DAYS")

	// rate limiting (future)
	_ = v.BindEnv("rate_limit.requests", "RATE_LIMIT_REQUESTS")
	_ = v.BindEnv("rate_limit.window_seconds", "RATE_LIMIT_WINDOW_SECONDS")
}
