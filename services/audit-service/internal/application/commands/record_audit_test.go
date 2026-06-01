package commands

import (
	"context"
	"testing"

	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
)

// MockAuditRepository for testing
type MockAuditRepository struct {
	recordedEntries []*domain.AuditEntry
	recordErr       error
}

func (m *MockAuditRepository) Record(ctx context.Context, entry *domain.AuditEntry) error {
	if m.recordErr != nil {
		return m.recordErr
	}
	m.recordedEntries = append(m.recordedEntries, entry)
	return nil
}

func (m *MockAuditRepository) GetByID(ctx context.Context, tenantID domain.TenantID, entryID domain.AuditEntryID) (*domain.AuditEntry, error) {
	return nil, nil
}

func (m *MockAuditRepository) Query(ctx context.Context, tenantID domain.TenantID, filters domain.QueryFilters) ([]*domain.AuditEntry, int, error) {
	return nil, 0, nil
}

func TestRecordAuditHandler_Handle(t *testing.T) {
	mockRepo := &MockAuditRepository{}
	handler := NewRecordAuditHandler(mockRepo)

	cmd := RecordAuditCommand{
		TenantID:     "550e8400-e29b-41d4-a716-446655440000",
		ActorID:      "actor-123",
		Action:       domain.ActionLogin,
		ResourceType: domain.ResourceUser,
		ResourceID:   "user-456",
		Description:  "User logged in",
		Success:      true,
		Changes: map[string]string{
			"status": "active",
		},
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.EntryID == "" {
		t.Error("entry ID should not be empty")
	}

	if len(mockRepo.recordedEntries) != 1 {
		t.Errorf("expected 1 recorded entry, got %d", len(mockRepo.recordedEntries))
	}

	recorded := mockRepo.recordedEntries[0]
	if recorded.ActorID() != "actor-123" {
		t.Error("actor ID mismatch")
	}
	if recorded.Action() != domain.ActionLogin {
		t.Error("action mismatch")
	}
}

func TestRecordAuditHandler_InvalidTenantID(t *testing.T) {
	mockRepo := &MockAuditRepository{}
	handler := NewRecordAuditHandler(mockRepo)

	cmd := RecordAuditCommand{
		TenantID:     "", // invalid
		ActorID:      "actor-123",
		Action:       domain.ActionLogin,
		ResourceType: domain.ResourceUser,
	}

	_, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for invalid tenant ID")
	}
}
