package queries

import (
	"context"
	"testing"

	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
)

// MockAuditRepository for query testing
type MockAuditRepositoryForQuery struct {
	entries  []*domain.AuditEntry
	queryErr error
}

func (m *MockAuditRepositoryForQuery) Record(ctx context.Context, entry *domain.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockAuditRepositoryForQuery) GetByID(ctx context.Context, tenantID domain.TenantID, entryID domain.AuditEntryID) (*domain.AuditEntry, error) {
	return nil, nil
}

func (m *MockAuditRepositoryForQuery) Query(ctx context.Context, tenantID domain.TenantID, filters domain.QueryFilters) ([]*domain.AuditEntry, int, error) {
	if m.queryErr != nil {
		return nil, 0, m.queryErr
	}

	// Simple filter logic for testing
	var results []*domain.AuditEntry
	for _, e := range m.entries {
		if e.TenantID() != tenantID {
			continue
		}
		if filters.ActorID != "" && e.ActorID() != filters.ActorID {
			continue
		}
		if filters.Action != "" && e.Action() != filters.Action {
			continue
		}
		if filters.ResourceType != "" && e.ResourceType() != filters.ResourceType {
			continue
		}
		results = append(results, e)
	}

	// Apply pagination
	limit := filters.Limit
	if limit <= 0 {
		limit = 100
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}

	if offset >= len(results) {
		return []*domain.AuditEntry{}, len(results), nil
	}

	return results[offset:end], len(results), nil
}

func TestQueryAuditTrailHandler_AllEntries(t *testing.T) {
	mockRepo := &MockAuditRepositoryForQuery{}
	handler := NewQueryAuditTrailHandler(mockRepo)

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	// Add some entries
	for i := 0; i < 3; i++ {
		entry, _ := domain.NewAuditEntry(
			tenantID,
			"actor-123",
			domain.ActionLogin,
			domain.ResourceUser,
			"",
			"Test entry",
			true,
			"",
			nil,
		)
		mockRepo.Record(context.Background(), entry)
	}

	filters := AuditQueryFilters{
		TenantID: tenantID.String(),
		Limit:    100,
		Offset:   0,
	}

	result, err := handler.Handle(context.Background(), filters)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result.Entries))
	}

	if result.TotalCount != 3 {
		t.Errorf("expected total count 3, got %d", result.TotalCount)
	}
}

func TestQueryAuditTrailHandler_FilterByActor(t *testing.T) {
	mockRepo := &MockAuditRepositoryForQuery{}
	handler := NewQueryAuditTrailHandler(mockRepo)

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	// Add entries with different actors
	entry1, _ := domain.NewAuditEntry(tenantID, "actor-1", domain.ActionLogin, domain.ResourceUser, "", "Entry 1", true, "", nil)
	entry2, _ := domain.NewAuditEntry(tenantID, "actor-2", domain.ActionLogin, domain.ResourceUser, "", "Entry 2", true, "", nil)
	mockRepo.Record(context.Background(), entry1)
	mockRepo.Record(context.Background(), entry2)

	filters := AuditQueryFilters{
		TenantID: tenantID.String(),
		ActorID:  "actor-1",
		Limit:    100,
	}

	result, err := handler.Handle(context.Background(), filters)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result.Entries))
	}

	if result.Entries[0].ActorID != "actor-1" {
		t.Errorf("expected actor-1, got %s", result.Entries[0].ActorID)
	}
}

func TestQueryAuditTrailHandler_FilterByAction(t *testing.T) {
	mockRepo := &MockAuditRepositoryForQuery{}
	handler := NewQueryAuditTrailHandler(mockRepo)

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	// Add entries with different actions
	entry1, _ := domain.NewAuditEntry(tenantID, "actor", domain.ActionLogin, domain.ResourceUser, "", "Entry 1", true, "", nil)
	entry2, _ := domain.NewAuditEntry(tenantID, "actor", domain.ActionEmployeeCreate, domain.ResourceEmployee, "", "Entry 2", true, "", nil)
	mockRepo.Record(context.Background(), entry1)
	mockRepo.Record(context.Background(), entry2)

	filters := AuditQueryFilters{
		TenantID: tenantID.String(),
		Action:   domain.ActionEmployeeCreate,
		Limit:    100,
	}

	result, err := handler.Handle(context.Background(), filters)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result.Entries))
	}

	if result.Entries[0].Action != domain.ActionEmployeeCreate {
		t.Errorf("expected EMPLOYEE_CREATED, got %s", result.Entries[0].Action)
	}
}

func TestQueryAuditTrailHandler_Pagination(t *testing.T) {
	mockRepo := &MockAuditRepositoryForQuery{}
	handler := NewQueryAuditTrailHandler(mockRepo)

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	// Add 10 entries
	for i := 0; i < 10; i++ {
		entry, _ := domain.NewAuditEntry(
			tenantID,
			"actor",
			domain.ActionLogin,
			domain.ResourceUser,
			"",
			"Entry",
			true,
			"",
			nil,
		)
		mockRepo.Record(context.Background(), entry)
	}

	// First page
	filters := AuditQueryFilters{
		TenantID: tenantID.String(),
		Limit:    5,
		Offset:   0,
	}

	result, err := handler.Handle(context.Background(), filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entries) != 5 {
		t.Errorf("expected 5 entries on first page, got %d", len(result.Entries))
	}
	if result.TotalCount != 10 {
		t.Errorf("expected total count 10, got %d", result.TotalCount)
	}

	// Second page
	filters.Offset = 5
	result, err = handler.Handle(context.Background(), filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entries) != 5 {
		t.Errorf("expected 5 entries on second page, got %d", len(result.Entries))
	}
}

func TestQueryAuditTrailHandler_DTOMapping(t *testing.T) {
	mockRepo := &MockAuditRepositoryForQuery{}
	handler := NewQueryAuditTrailHandler(mockRepo)

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	changes := map[string]string{"field": "value"}

	entry, _ := domain.NewAuditEntry(
		tenantID,
		"actor-123",
		domain.ActionLogin,
		domain.ResourceUser,
		"resource-456",
		"User logged in",
		false,
		"invalid credentials",
		changes,
	)
	mockRepo.Record(context.Background(), entry)

	filters := AuditQueryFilters{
		TenantID: tenantID.String(),
		Limit:    100,
	}

	result, err := handler.Handle(context.Background(), filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dto := result.Entries[0]

	if dto.ID != entry.ID().String() {
		t.Error("ID mismatch")
	}
	if dto.TenantID != tenantID.String() {
		t.Error("tenant ID mismatch")
	}
	if dto.ActorID != "actor-123" {
		t.Error("actor ID mismatch")
	}
	if dto.Action != domain.ActionLogin {
		t.Error("action mismatch")
	}
	if dto.ResourceType != domain.ResourceUser {
		t.Error("resource type mismatch")
	}
	if dto.ResourceID != "resource-456" {
		t.Error("resource ID mismatch")
	}
	if dto.Description != "User logged in" {
		t.Error("description mismatch")
	}
	if dto.Success != false {
		t.Error("success should be false")
	}
	if dto.ErrorMessage != "invalid credentials" {
		t.Error("error message mismatch")
	}
	if len(dto.Changes) != 1 || dto.Changes["field"] != "value" {
		t.Error("changes mismatch")
	}
	if dto.CreatedAt.IsZero() {
		t.Error("created_at should not be zero")
	}
}

// Note: Tenant ID validation happens at gRPC boundary, not in handler.
// The handler delegates validation to domain layer (MustNewTenantID).
// Empty tenant ID would panic in MustNewTenantID (by design for programming errors).

func TestQueryAuditTrailHandler_NoResults(t *testing.T) {
	mockRepo := &MockAuditRepositoryForQuery{}
	handler := NewQueryAuditTrailHandler(mockRepo)

	filters := AuditQueryFilters{
		TenantID: "550e8400-e29b-41d4-a716-446655440000",
		Limit:    100,
	}

	result, err := handler.Handle(context.Background(), filters)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}

	if result.TotalCount != 0 {
		t.Errorf("expected total count 0, got %d", result.TotalCount)
	}
}
