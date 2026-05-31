package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// UserID is a typed UUID for user identity.
type UserID struct {
	value string
}

// NewUserID creates a new UserID from a UUID string.
func NewUserID(id string) (UserID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return UserID{}, fmt.Errorf("invalid user ID: %w", err)
	}
	return UserID{value: id}, nil
}

// NewUserIDOrNil parses a user ID, returning zero struct if invalid.
func NewUserIDOrNil(id string) UserID {
	uid, _ := NewUserID(id)
	return uid
}

// MustNewUserID panics on invalid ID.
func MustNewUserID(id string) UserID {
	uid, err := NewUserID(id)
	if err != nil {
		panic(err)
	}
	return uid
}

// GenerateUserID creates a new random UserID.
func GenerateUserID() UserID {
	return UserID{value: uuid.New().String()}
}

// String returns the UUID string representation.
func (u UserID) String() string {
	return u.value
}

// IsZero reports whether the ID is uninitialized.
func (u UserID) IsZero() bool {
	return u.value == ""
}

// Equals compares two UserIDs for equality.
func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}
