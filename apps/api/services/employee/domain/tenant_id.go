package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// TenantID is a typed UUID wrapper to prevent accidental string comparisons.
type TenantID string

// NewTenantID creates a new TenantID from a string, validating UUID format.
func NewTenantID(id string) (TenantID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("invalid tenant ID: %w", err)
	}
	return TenantID(id), nil
}

// MustNewTenantID creates a new TenantID, panicking if invalid.
func MustNewTenantID(id string) TenantID {
	tid, err := NewTenantID(id)
	if err != nil {
		panic(err)
	}
	return tid
}

// String returns the string representation of the TenantID.
func (t TenantID) String() string {
	return string(t)
}

// IsZero returns true if the TenantID is empty.
func (t TenantID) IsZero() bool {
	return t == ""
}

// Equals compares two TenantIDs for equality.
func (t TenantID) Equals(other TenantID) bool {
	return t == other
}
