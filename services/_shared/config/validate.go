package config

import (
	"errors"
	"fmt"
	"strings"
)

// Validatable is implemented by config structs that need to be validated.
type Validatable interface {
	Validate() error
}

// ValidateAll validates all provided config structs and returns a combined error.
func ValidateAll(validators ...Validatable) error {
	var errs []error
	for _, v := range validators {
		if err := v.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// required returns an error if the given value is empty.
func required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

// Validate checks that AppConfig has required fields.
func (c AppConfig) Validate() error {
	return nil // app config has no required fields
}

// Validate checks that DatabaseConfig has required fields.
func (c DatabaseConfig) Validate() error {
	return required("DATABASE_URL", c.URL)
}

// Validate checks that ServerConfig has required fields.
func (c ServerConfig) Validate() error {
	return required("GRPC_PORT", c.GRPCPort)
}

// Validate checks that NATSConfig has required fields.
func (c NATSConfig) Validate() error {
	return required("NATS_URL", c.URL)
}

// Validate checks that ObservabilityConfig has required fields.
func (c ObservabilityConfig) Validate() error {
	return nil // observability config has no required fields
}

// Validate checks that JWTConfig has required fields.
// Either the PEM-encoded or base64-encoded keys must be provided.
func (c JWTConfig) Validate() error {
	var errs []error

	if c.PrivateKey == "" && c.PrivateKeyBase64 == "" {
		errs = append(errs, fmt.Errorf("AUTH_PRIVATE_KEY or AUTH_PRIVATE_KEY_BASE64 is required"))
	}

	if c.PublicKey == "" && c.PublicKeyBase64 == "" {
		errs = append(errs, fmt.Errorf("AUTH_PUBLIC_KEY or AUTH_PUBLIC_KEY_BASE64 is required"))
	}

	return errors.Join(errs...)
}

// Validate checks that RedisConfig has required fields.
func (c RedisConfig) Validate() error {
	// Redis config is optional; only validate if explicitly enabled
	// (a service may not use Redis at all)
	return nil
}

// Validate checks that SMTPConfig has required fields.
// SMTP is required only for notification-service.
func (c SMTPConfig) Validate() error {
	// SMTP validation is deferred to notification-service's specific config
	// because not all services use SMTP
	return nil
}
