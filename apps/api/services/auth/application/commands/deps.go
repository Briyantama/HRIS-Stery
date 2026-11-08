package commands

import "github.com/Briyantama/HRIS-Stery/services/auth/application"

type (
	TokenService   = application.TokenService
	EventPublisher = application.EventPublisher
	TokenClaims    = application.TokenClaims
)
