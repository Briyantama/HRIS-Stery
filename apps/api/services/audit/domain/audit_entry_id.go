package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type AuditEntryID struct {
	value uuid.UUID
}

func NewAuditEntryID() AuditEntryID {
	return AuditEntryID{value: uuid.New()}
}

func NewAuditEntryIDFromString(s string) (AuditEntryID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return AuditEntryID{}, fmt.Errorf("invalid audit entry ID: %w", err)
	}
	return AuditEntryID{value: parsed}, nil
}

func (id AuditEntryID) String() string {
	return id.value.String()
}

func (id AuditEntryID) IsZero() bool {
	return id.value == uuid.Nil
}

func MustNewAuditEntryID(s string) AuditEntryID {
	id, err := NewAuditEntryIDFromString(s)
	if err != nil {
		panic(err)
	}
	return id
}
