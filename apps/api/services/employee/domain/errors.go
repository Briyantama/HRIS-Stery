package domain

import "errors"

// Sentinel errors the HTTP layer maps to status codes. Wrap them with %w.
var (
	// ErrNotFound means the resource does not exist for the tenant (cross-tenant ids look the same).
	ErrNotFound = errors.New("not found")
	// ErrConflict means the request clashes with existing state (duplicate email, already terminated).
	ErrConflict = errors.New("conflict")
	// ErrInvalid means the input failed validation.
	ErrInvalid = errors.New("invalid input")
)
