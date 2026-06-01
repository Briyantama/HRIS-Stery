package domain

import "github.com/google/uuid"

type TenantID struct {
	value uuid.UUID
}

func MustNewTenantID(s string) TenantID {
	id, _ := uuid.Parse(s)
	return TenantID{value: id}
}

func (t TenantID) String() string { return t.value.String() }
func (t TenantID) IsZero() bool   { return t.value == uuid.Nil }
