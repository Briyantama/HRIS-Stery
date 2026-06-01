package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// DepartmentID is a typed UUID wrapper to prevent accidental string comparisons.
type DepartmentID string

// NewDepartmentID creates a new DepartmentID from a string, validating UUID format.
func NewDepartmentID(id string) (DepartmentID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("invalid department ID: %w", err)
	}
	return DepartmentID(id), nil
}

// GenerateDepartmentID creates a new random DepartmentID.
func GenerateDepartmentID() DepartmentID {
	return DepartmentID(uuid.New().String())
}

// MustNewDepartmentID creates a new DepartmentID, panicking if invalid.
func MustNewDepartmentID(id string) DepartmentID {
	did, err := NewDepartmentID(id)
	if err != nil {
		panic(err)
	}
	return did
}

// String returns the string representation of the DepartmentID.
func (d DepartmentID) String() string {
	return string(d)
}

// IsZero returns true if the DepartmentID is empty.
func (d DepartmentID) IsZero() bool {
	return d == ""
}

// Equals compares two DepartmentIDs for equality.
func (d DepartmentID) Equals(other DepartmentID) bool {
	return d == other
}
