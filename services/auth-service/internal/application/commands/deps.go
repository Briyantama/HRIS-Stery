package commands

import "github.com/hris-stery/hris-stery/services/auth-service/internal/application"

type (
	TokenService   = application.TokenService
	EventPublisher = application.EventPublisher
	TokenClaims    = application.TokenClaims
)
