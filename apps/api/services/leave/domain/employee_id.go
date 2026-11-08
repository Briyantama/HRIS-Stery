package domain

import (
	"github.com/google/uuid"
)

// EmployeeID is a typed UUID wrapper for employee identification.
type EmployeeID string

// NewEmployeeID creates a new EmployeeID from a string.
func NewEmployeeID(id string) (EmployeeID, error) {
	if id == "" {
		return "", ErrInvalidEmployeeID
	}
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidEmployeeID
	}
	return EmployeeID(id), nil
}

// MustNewEmployeeID creates a new EmployeeID or panics.
func MustNewEmployeeID(id string) EmployeeID {
	eid, err := NewEmployeeID(id)
	if err != nil {
		panic(err)
	}
	return eid
}

// GenerateEmployeeID generates a new random EmployeeID.
func GenerateEmployeeID() EmployeeID {
	return EmployeeID(uuid.New().String())
}

// String returns the string representation of EmployeeID.
func (e EmployeeID) String() string {
	return string(e)
}

// IsZero returns true if the EmployeeID is empty.
func (e EmployeeID) IsZero() bool {
	return e == ""
}

// Equals compares two EmployeeIDs.
func (e EmployeeID) Equals(other EmployeeID) bool {
	return e == other
}
