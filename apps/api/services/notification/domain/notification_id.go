package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type NotificationID struct {
	value uuid.UUID
}

func NewNotificationID(id string) (NotificationID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return NotificationID{}, fmt.Errorf("invalid notification id: %w", err)
	}
	return NotificationID{value: parsed}, nil
}

func GenerateNotificationID() NotificationID {
	return NotificationID{value: uuid.New()}
}

func (id NotificationID) String() string {
	return id.value.String()
}

func (id NotificationID) IsZero() bool {
	return id.value == uuid.Nil
}

func (id NotificationID) Equals(other NotificationID) bool {
	return id.value == other.value
}
