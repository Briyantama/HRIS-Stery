package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type DocumentID struct {
	value uuid.UUID
}

func NewDocumentID() DocumentID {
	return DocumentID{value: uuid.New()}
}

func NewDocumentIDFromString(s string) (DocumentID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return DocumentID{}, fmt.Errorf("invalid document id: %w", err)
	}
	return DocumentID{value: parsed}, nil
}

func (id DocumentID) String() string {
	return id.value.String()
}

func (id DocumentID) IsZero() bool {
	return id.value == uuid.Nil
}

func MustNewDocumentID(s string) DocumentID {
	id, err := NewDocumentIDFromString(s)
	if err != nil {
		panic(err)
	}
	return id
}
