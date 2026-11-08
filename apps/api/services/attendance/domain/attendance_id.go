package domain

import (
	"github.com/google/uuid"
)

// AttendanceID is a typed UUID wrapper for attendance record IDs.
type AttendanceID string

// NewAttendanceID creates a new AttendanceID from a string.
func NewAttendanceID(id string) (AttendanceID, error) {
	if id == "" {
		return "", ErrInvalidAttendanceID
	}
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidAttendanceID
	}
	return AttendanceID(id), nil
}

// MustNewAttendanceID creates a new AttendanceID or panics.
func MustNewAttendanceID(id string) AttendanceID {
	aid, err := NewAttendanceID(id)
	if err != nil {
		panic(err)
	}
	return aid
}

// GenerateAttendanceID generates a new random AttendanceID.
func GenerateAttendanceID() AttendanceID {
	return AttendanceID(uuid.New().String())
}

// String returns the string representation of AttendanceID.
func (a AttendanceID) String() string {
	return string(a)
}

// IsZero returns true if the AttendanceID is empty.
func (a AttendanceID) IsZero() bool {
	return a == ""
}

// Equals compares two AttendanceIDs.
func (a AttendanceID) Equals(other AttendanceID) bool {
	return a == other
}
