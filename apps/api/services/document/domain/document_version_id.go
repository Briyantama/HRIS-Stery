package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type DocumentVersionID struct {
	value uuid.UUID
}

func NewDocumentVersionID() DocumentVersionID {
	return DocumentVersionID{value: uuid.New()}
}

func NewDocumentVersionIDFromString(s string) (DocumentVersionID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return DocumentVersionID{}, fmt.Errorf("invalid document version id: %w", err)
	}
	return DocumentVersionID{value: parsed}, nil
}

func (id DocumentVersionID) String() string {
	return id.value.String()
}

func (id DocumentVersionID) IsZero() bool {
	return id.value == uuid.Nil
}

func MustNewDocumentVersionID(s string) DocumentVersionID {
	id, err := NewDocumentVersionIDFromString(s)
	if err != nil {
		panic(err)
	}
	return id
}
