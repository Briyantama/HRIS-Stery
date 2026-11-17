// Package config loads the API's settings from environment variables and validates them at startup.
package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultPort         = "8000"
	defaultSMTPPort     = 587
	defaultTemplatesDir = "./services/notification/templates"
	defaultAccessHours  = 1
	defaultRefreshDays  = 7
	defaultLogLevel     = "info"
	maxPort             = 65535
)

// S3 locates the S3-compatible bucket that holds documents.
type S3 struct {
	Endpoint string
	// PublicEndpoint is the origin browsers reach the bucket at; it defaults to Endpoint.
	PublicEndpoint  string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	// PathStyle addresses the bucket as {endpoint}/{bucket}, which self-hosted S3 servers (RustFS, MinIO) need.
	PathStyle bool
}

// SMTP is the outgoing mail server. An empty Host keeps the notification service in mock mode.
type SMTP struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromEmail  string
	FromName   string
	TLSEnabled bool
}

// Config holds every setting the API needs. Secret fields are never printed.
type Config struct {
	Port     string
	LogLevel string

	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	// RedisTLS connects to Redis over TLS (REDIS_TLS=true).
	RedisTLS bool

	// AuthPrivateKeyPEM and AuthPublicKeyPEM are the RS256 key pair that signs and verifies access tokens.
	AuthPrivateKeyPEM []byte
	AuthPublicKeyPEM  []byte
	// AccessTokenHours and RefreshTokenDays are the token lifetimes (JWT_EXPIRATION_HOURS, REFRESH_TOKEN_EXPIRATION_DAYS).
	AccessTokenHours int
	RefreshTokenDays int

	S3 S3

	SMTP SMTP
	// EmailTemplatesDir holds the notification templates.
	EmailTemplatesDir string

	// TrustedProxies are the reverse proxies whose X-Forwarded-For is believed. Empty means the
	// connection address is always used.
	TrustedProxies []netip.Prefix
}

// Load reads and validates the configuration, reporting every problem at once.
// Error messages name variables but never include their values.
func Load(getenv func(string) string) (Config, error) {
	var problems []string

	required := func(name string) string {
		value := getenv(name)
		if value == "" {
			problems = append(problems, name+" is required")
		}
		return value
	}
	port := func(name, fallback string) string {
		value := getenv(name)
		if value == "" {
			return fallback
		}
		if n, err := strconv.Atoi(value); err != nil || n < 1 || n > maxPort {
			problems = append(problems, name+" must be a port number between 1 and 65535")
		}
		return value
	}
	boolean := func(name string, fallback bool) bool {
		raw := getenv(name)
		if raw == "" {
			return fallback
		}
		v, err := strconv.ParseBool(raw)
		if err != nil {
			problems = append(problems, name+" must be true or false")
			return fallback
		}
		return v
	}
	positive := func(name string, fallback int) int {
		raw := getenv(name)
		if raw == "" {
			return fallback
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			problems = append(problems, name+" must be a whole number, at least 1")
			return fallback
		}
		return n
	}
	baseURL := func(name string) string {
		raw := getenv(name)
		if raw == "" {
			return ""
		}
		v, ok := normalizeBaseURL(raw)
		if !ok {
			problems = append(problems, name+" must be an absolute http or https URL without credentials, query or fragment")
		}
		return v
	}

	cfg := Config{
		Port:              port("PORT", defaultPort),
		LogLevel:          orDefault(getenv("LOG_LEVEL"), defaultLogLevel),
		DatabaseURL:       required("DATABASE_URL"),
		RedisAddr:         required("REDIS_ADDR"),
		RedisPassword:     getenv("REDIS_PASSWORD"),
		RedisTLS:          boolean("REDIS_TLS", false),
		AccessTokenHours:  positive("JWT_EXPIRATION_HOURS", defaultAccessHours),
		RefreshTokenDays:  positive("REFRESH_TOKEN_EXPIRATION_DAYS", defaultRefreshDays),
		EmailTemplatesDir: orDefault(getenv("EMAIL_TEMPLATES_DIR"), defaultTemplatesDir),
		S3: S3{
			Region:          required("S3_REGION"),
			Bucket:          required("S3_BUCKET"),
			AccessKeyID:     required("S3_ACCESS_KEY_ID"),
			SecretAccessKey: required("S3_SECRET_ACCESS_KEY"),
			PathStyle:       boolean("S3_PATH_STYLE", true),
		},
		SMTP: SMTP{
			Host:       getenv("SMTP_HOST"),
			Port:       defaultSMTPPort,
			Username:   getenv("SMTP_USERNAME"),
			Password:   getenv("SMTP_PASSWORD"),
			FromEmail:  getenv("SMTP_FROM_EMAIL"),
			FromName:   getenv("SMTP_FROM_NAME"),
			TLSEnabled: boolean("SMTP_TLS_ENABLED", true),
		},
	}
	if raw := port("SMTP_PORT", ""); raw != "" {
		cfg.SMTP.Port, _ = strconv.Atoi(raw)
	}

	if cfg.S3.Endpoint = baseURL("S3_ENDPOINT"); getenv("S3_ENDPOINT") == "" {
		problems = append(problems, "S3_ENDPOINT is required")
	}
	cfg.S3.PublicEndpoint = orDefault(baseURL("S3_PUBLIC_ENDPOINT"), cfg.S3.Endpoint)

	var err error
	if cfg.AuthPrivateKeyPEM, err = loadPEM(getenv, "AUTH_PRIVATE_KEY", "AUTH_PRIVATE_KEY_BASE64"); err != nil {
		problems = append(problems, err.Error())
	}
	if cfg.AuthPublicKeyPEM, err = loadPEM(getenv, "AUTH_PUBLIC_KEY", "AUTH_PUBLIC_KEY_BASE64"); err != nil {
		problems = append(problems, err.Error())
	}

	if raw := getenv("TRUSTED_PROXY_CIDRS"); raw != "" {
		proxies, ok := parseProxies(raw)
		if !ok {
			problems = append(problems, "TRUSTED_PROXY_CIDRS must be a comma-separated list of CIDR ranges or addresses")
		}
		cfg.TrustedProxies = proxies
	}

	if len(problems) > 0 {
		return Config{}, errors.New("invalid configuration: " + strings.Join(problems, "; "))
	}
	return cfg, nil
}

// loadPEM returns the key from the base64 variable if set, else from the plain variable (where a
// literal \n stands for a newline). The error names the variables, never their contents.
func loadPEM(getenv func(string) string, plainName, base64Name string) ([]byte, error) {
	if b64 := getenv(base64Name); b64 != "" {
		pem, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("%s must be valid base64", base64Name)
		}
		return pem, nil
	}
	if plain := getenv(plainName); plain != "" {
		return []byte(strings.ReplaceAll(plain, `\n`, "\n")), nil
	}
	return nil, fmt.Errorf("%s or %s is required", plainName, base64Name)
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// parseProxies reads a comma-separated list of CIDR ranges; a bare address is a single-host range.
func parseProxies(raw string) ([]netip.Prefix, bool) {
	var out []netip.Prefix
	for part := range strings.SplitSeq(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if prefix, err := netip.ParsePrefix(part); err == nil {
			out = append(out, prefix.Masked())
			continue
		}
		addr, err := netip.ParseAddr(part)
		if err != nil {
			return nil, false
		}
		out = append(out, netip.PrefixFrom(addr, addr.BitLen()))
	}
	return out, true
}

// normalizeBaseURL accepts an absolute http(s) URL with no credentials, query or fragment and
// returns it without a trailing slash.
func normalizeBaseURL(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		return "", false
	}
	if u.RawQuery != "" || u.Fragment != "" || strings.ContainsAny(raw, "?#") {
		return "", false
	}
	return strings.TrimRight(raw, "/"), true
}

// String summarizes the configuration with every secret redacted.
func (c Config) String() string {
	return fmt.Sprintf(
		"Config{port=:%s logLevel=%s redis=%s s3=%s/%s smtpHost=%s smtpPort=%d templates=%s trustedProxies=%d authKeys=[redacted] database=[redacted] redisPassword=[redacted] smtpPassword=[redacted] s3Secret=[redacted]}",
		c.Port, c.LogLevel, c.RedisAddr, c.S3.Endpoint, c.S3.Bucket, c.SMTP.Host, c.SMTP.Port, c.EmailTemplatesDir, len(c.TrustedProxies),
	)
}
