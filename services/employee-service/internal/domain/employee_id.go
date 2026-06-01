package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// EmployeeID is a typed UUID wrapper to prevent accidental string comparisons.
type EmployeeID string

// NewEmployeeID creates a new EmployeeID from a string, validating UUID format.
func NewEmployeeID(id string) (EmployeeID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("invalid employee ID: %w", err)
	}
	return EmployeeID(id), nil
}

// GenerateEmployeeID creates a new random EmployeeID.
func GenerateEmployeeID() EmployeeID {
	return EmployeeID(uuid.New().String())
}

// MustNewEmployeeID creates a new EmployeeID, panicking if invalid.
func MustNewEmployeeID(id string) EmployeeID {
	eid, err := NewEmployeeID(id)
	if err != nil {
		panic(err)
	}
	return eid
}

// String returns the string representation of the EmployeeID.
func (e EmployeeID) String() string {
	return string(e)
}

// IsZero returns true if the EmployeeID is empty.
func (e EmployeeID) IsZero() bool {
	return e == ""
}

// Equals compares two EmployeeIDs for equality.
func (e EmployeeID) Equals(other EmployeeID) bool {
	return e == other
}
