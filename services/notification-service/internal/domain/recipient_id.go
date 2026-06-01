package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type RecipientID struct {
	value uuid.UUID
}

func NewRecipientID(id string) (RecipientID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return RecipientID{}, fmt.Errorf("invalid recipient id: %w", err)
	}
	return RecipientID{value: parsed}, nil
}

func GenerateRecipientID() RecipientID {
	return RecipientID{value: uuid.New()}
}

func (id RecipientID) String() string {
	return id.value.String()
}

func (id RecipientID) IsZero() bool {
	return id.value == uuid.Nil
}

func (id RecipientID) Equals(other RecipientID) bool {
	return id.value == other.value
}
