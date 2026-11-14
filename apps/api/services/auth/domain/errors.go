package domain

import "errors"

// Sentinel errors the HTTP layer maps to status codes. Wrap them with %w.
var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrInvalidInput    = errors.New("invalid input")
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
)
