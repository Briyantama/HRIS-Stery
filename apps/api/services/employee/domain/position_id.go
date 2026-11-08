package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// PositionID is a typed UUID wrapper to prevent accidental string comparisons.
type PositionID string

// NewPositionID creates a new PositionID from a string, validating UUID format.
func NewPositionID(id string) (PositionID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("invalid position ID: %w", err)
	}
	return PositionID(id), nil
}

// GeneratePositionID creates a new random PositionID.
func GeneratePositionID() PositionID {
	return PositionID(uuid.New().String())
}

// MustNewPositionID creates a new PositionID, panicking if invalid.
func MustNewPositionID(id string) PositionID {
	pid, err := NewPositionID(id)
	if err != nil {
		panic(err)
	}
	return pid
}

// String returns the string representation of the PositionID.
func (p PositionID) String() string {
	return string(p)
}

// IsZero returns true if the PositionID is empty.
func (p PositionID) IsZero() bool {
	return p == ""
}

// Equals compares two PositionIDs for equality.
func (p PositionID) Equals(other PositionID) bool {
	return p == other
}
