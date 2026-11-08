package domain

import (
	"github.com/google/uuid"
)

// LeaveRequestID is a typed UUID wrapper for leave request identification.
type LeaveRequestID string

// NewLeaveRequestID creates a new LeaveRequestID from a string.
func NewLeaveRequestID(id string) (LeaveRequestID, error) {
	if id == "" {
		return "", ErrInvalidLeaveRequestID
	}
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidLeaveRequestID
	}
	return LeaveRequestID(id), nil
}

// MustNewLeaveRequestID creates a new LeaveRequestID or panics.
func MustNewLeaveRequestID(id string) LeaveRequestID {
	lrid, err := NewLeaveRequestID(id)
	if err != nil {
		panic(err)
	}
	return lrid
}

// GenerateLeaveRequestID generates a new random LeaveRequestID.
func GenerateLeaveRequestID() LeaveRequestID {
	return LeaveRequestID(uuid.New().String())
}

// String returns the string representation of LeaveRequestID.
func (l LeaveRequestID) String() string {
	return string(l)
}

// IsZero returns true if the LeaveRequestID is empty.
func (l LeaveRequestID) IsZero() bool {
	return l == ""
}

// Equals compares two LeaveRequestIDs.
func (l LeaveRequestID) Equals(other LeaveRequestID) bool {
	return l == other
}
