package domain

import (
	"fmt"

	"uuid"
)

// TenantID is a typed UUID for tenant identity.
// Using a value object prevents accidental raw string comparisons and ensures type safety.
type TenantID struct {
	value string
}

// NewTenantID creates a new TenantID from a UUID string.
func NewTenantID(id string) (TenantID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return TenantID{}, fmt.Errorf("invalid tenant ID: %w", err)
	}
	return TenantID{value: id}, nil
}

// NewTenantIDOrNil parses a tenant ID, returning nil-safe struct if invalid.
// Only for trusted internal sources; external callers must use NewTenantID.
func NewTenantIDOrNil(id string) TenantID {
	tid, _ := NewTenantID(id)
	return tid
}

// MustNewTenantID panics on invalid ID. Use only in tests or guaranteed-valid contexts.
func MustNewTenantID(id string) TenantID {
	tid, err := NewTenantID(id)
	if err != nil {
		panic(err)
	}
	return tid
}

// String returns the UUID string representation.
func (t TenantID) String() string {
	return t.value
}

// IsZero reports whether the ID is uninitialized.
func (t TenantID) IsZero() bool {
	return t.value == ""
}

// Equals compares two TenantIDs for equality.
func (t TenantID) Equals(other TenantID) bool {
	return t.value == other.value
}

// MarshalText makes the ID a plain string in JSON (JWT claims).
func (t TenantID) MarshalText() ([]byte, error) { return []byte(t.value), nil }

// UnmarshalText parses a UUID string; the empty string is the zero ID.
func (t *TenantID) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*t = TenantID{}
		return nil
	}
	id, err := NewTenantID(string(b))
	*t = id
	return err
}
