package config

// AppConfig contains application-level metadata.
type AppConfig struct {
	Env     string `mapstructure:"env"`
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// DatabaseConfig contains PostgreSQL connection settings.
type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

// ServerConfig contains gRPC and HTTP server settings.
type ServerConfig struct {
	GRPCPort    string `mapstructure:"grpc_port"`
	MetricsPort string `mapstructure:"metrics_port"`
}

// NATSConfig contains NATS connection settings.
type NATSConfig struct {
	URL string `mapstructure:"url"`
}

// ObservabilityConfig contains logging and tracing settings.
type ObservabilityConfig struct {
	LogLevel     string `mapstructure:"log_level"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint"`
	Enabled      bool   `mapstructure:"enabled"`
}

// JWTConfig contains JWT signing and validation settings (auth-service).
type JWTConfig struct {
	PrivateKey        string `mapstructure:"private_key"`
	PublicKey         string `mapstructure:"public_key"`
	PrivateKeyBase64  string `mapstructure:"private_key_b64"`
	PublicKeyBase64   string `mapstructure:"public_key_b64"`
	ExpirationHours   int    `mapstructure:"expiration_hours"`
	RefreshExpiryDays int    `mapstructure:"refresh_expiry_days"`
}

// RedisConfig contains Redis connection settings (auth-service token store).
type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

// SMTPConfig contains email/SMTP settings (notification-service).
type SMTPConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	FromEmail  string `mapstructure:"from_email"`
	FromName   string `mapstructure:"from_name"`
	TLSEnabled bool   `mapstructure:"tls_enabled"`
}
