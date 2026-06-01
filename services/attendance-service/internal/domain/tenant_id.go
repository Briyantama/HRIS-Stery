package domain

import (
	"github.com/google/uuid"
)

// TenantID is a typed UUID wrapper for tenant identification.
// Used for RLS context and multi-tenancy boundaries.
type TenantID string

// NewTenantID creates a new TenantID from a string.
func NewTenantID(id string) (TenantID, error) {
	if id == "" {
		return "", ErrInvalidTenantID
	}
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidTenantID
	}
	return TenantID(id), nil
}

// MustNewTenantID creates a new TenantID or panics.
func MustNewTenantID(id string) TenantID {
	tid, err := NewTenantID(id)
	if err != nil {
		panic(err)
	}
	return tid
}

// String returns the string representation of TenantID.
func (t TenantID) String() string {
	return string(t)
}

// IsZero returns true if the TenantID is empty.
func (t TenantID) IsZero() bool {
	return t == ""
}

// Equals compares two TenantIDs.
func (t TenantID) Equals(other TenantID) bool {
	return t == other
}
