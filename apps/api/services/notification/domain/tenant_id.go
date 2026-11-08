package domain

import (
	"github.com/google/uuid"
)

type TenantID string

func NewTenantID(id string) (TenantID, error) {
	if id == "" {
		return "", ErrInvalidTenantID
	}
	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidTenantID
	}
	return TenantID(id), nil
}

func MustNewTenantID(id string) TenantID {
	tid, err := NewTenantID(id)
	if err != nil {
		panic(err)
	}
	return tid
}

func (t TenantID) String() string {
	return string(t)
}

func (t TenantID) IsZero() bool {
	return t == ""
}

func (t TenantID) Equals(other TenantID) bool {
	return t == other
}
