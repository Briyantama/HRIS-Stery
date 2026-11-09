package domain

import "uuid"

type TenantID struct {
	value uuid.UUID
}

// NewTenantID parses a tenant UUID; the nil UUID is rejected.
func NewTenantID(s string) (TenantID, error) {
	id, err := uuid.Parse(s)
	if err != nil || id == uuid.Nil() {
		return TenantID{}, ErrInvalidTenantID
	}
	return TenantID{value: id}, nil
}

func (t TenantID) String() string { return t.value.String() }
func (t TenantID) IsZero() bool   { return t.value == uuid.Nil() }
