package domain

import (
	"github.com/google/uuid"
)

// LeaveBalanceID is a typed UUID wrapper for leave balance identification.
type LeaveBalanceID string

// NewLeaveBalanceID creates a new LeaveBalanceID from a string.
func NewLeaveBalanceID(id string) (LeaveBalanceID, error) {
	if id == "" {
		return "", ErrInvalidLeaveBalanceID
	}
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidLeaveBalanceID
	}
	return LeaveBalanceID(id), nil
}

// MustNewLeaveBalanceID creates a new LeaveBalanceID or panics.
func MustNewLeaveBalanceID(id string) LeaveBalanceID {
	lbid, err := NewLeaveBalanceID(id)
	if err != nil {
		panic(err)
	}
	return lbid
}

// GenerateLeaveBalanceID generates a new random LeaveBalanceID.
func GenerateLeaveBalanceID() LeaveBalanceID {
	return LeaveBalanceID(uuid.New().String())
}

// String returns the string representation of LeaveBalanceID.
func (l LeaveBalanceID) String() string {
	return string(l)
}

// IsZero returns true if the LeaveBalanceID is empty.
func (l LeaveBalanceID) IsZero() bool {
	return l == ""
}

// Equals compares two LeaveBalanceIDs.
func (l LeaveBalanceID) Equals(other LeaveBalanceID) bool {
	return l == other
}
