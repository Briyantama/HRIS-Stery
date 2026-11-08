package domain

import (
	"github.com/google/uuid"
)

// LeaveTypeID is a typed UUID wrapper for leave type identification.
type LeaveTypeID string

// NewLeaveTypeID creates a new LeaveTypeID from a string.
func NewLeaveTypeID(id string) (LeaveTypeID, error) {
	if id == "" {
		return "", ErrInvalidLeaveTypeID
	}
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidLeaveTypeID
	}
	return LeaveTypeID(id), nil
}

// MustNewLeaveTypeID creates a new LeaveTypeID or panics.
func MustNewLeaveTypeID(id string) LeaveTypeID {
	ltid, err := NewLeaveTypeID(id)
	if err != nil {
		panic(err)
	}
	return ltid
}

// GenerateLeaveTypeID generates a new random LeaveTypeID.
func GenerateLeaveTypeID() LeaveTypeID {
	return LeaveTypeID(uuid.New().String())
}

// String returns the string representation of LeaveTypeID.
func (l LeaveTypeID) String() string {
	return string(l)
}

// IsZero returns true if the LeaveTypeID is empty.
func (l LeaveTypeID) IsZero() bool {
	return l == ""
}

// Equals compares two LeaveTypeIDs.
func (l LeaveTypeID) Equals(other LeaveTypeID) bool {
	return l == other
}
