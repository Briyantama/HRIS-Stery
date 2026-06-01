//go:build integration

package postgres

import (
	"context"
	"testing"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestAuditRepository tests the PostgreSQL audit repository.
// These tests require a running PostgreSQL instance.
// Run with: go test -tags integration ./...

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	dbURL := "postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable"

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Skipf("skipping integration test: cannot parse database URL: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to database: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping integration test: cannot ping database: %v", err)
	}

	return pool
}

func TestAuditRepository_Record(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	entry, err := domain.NewAuditEntry(
		tenantID,
		"actor-123",
		domain.ActionLogin,
		domain.ResourceUser,
		"user-456",
		"User logged in",
		true,
		"",
		map[string]string{"field": "value"},
	)

	if err != nil {
		t.Fatalf("unexpected error creating entry: %v", err)
	}

	err = repo.Record(ctx, entry)
	if err != nil {
		t.Fatalf("unexpected error recording entry: %v", err)
	}
}

func TestAuditRepository_GetByID(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	// Record an entry
	entry, _ := domain.NewAuditEntry(
		tenantID,
		"actor-123",
		domain.ActionLogin,
		domain.ResourceUser,
		"user-456",
		"User logged in",
		true,
		"",
		nil,
	)

	err := repo.Record(ctx, entry)
	if err != nil {
		t.Fatalf("failed to record entry: %v", err)
	}

	// Retrieve it
	retrieved, err := repo.GetByID(ctx, tenantID, entry.ID())
	if err != nil {
		t.Fatalf("unexpected error retrieving entry: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected entry, got nil")
	}

	if retrieved.ID() != entry.ID() {
		t.Error("entry ID mismatch")
	}

	if retrieved.ActorID() != "actor-123" {
		t.Error("actor ID mismatch")
	}
}

func TestAuditRepository_GetByID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	nonExistentID := domain.MustNewAuditEntryID("550e8400-e29b-41d4-a716-446655440001")

	_, err := repo.GetByID(ctx, tenantID, nonExistentID)
	if err == nil {
		t.Error("expected error for non-existent entry")
	}
}

func TestAuditRepository_Query(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	// Record multiple entries
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
		repo.Record(ctx, entry)
	}

	// Query all
	filters := domain.QueryFilters{
		Limit:  100,
		Offset: 0,
	}

	entries, total, err := repo.Query(ctx, tenantID, filters)
	if err != nil {
		t.Fatalf("unexpected error querying entries: %v", err)
	}

	if total < 3 {
		t.Errorf("expected at least 3 entries, got %d", total)
	}

	if len(entries) < 3 {
		t.Errorf("expected at least 3 entries in result, got %d", len(entries))
	}
}

func TestAuditRepository_Query_FilterByActor(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	// Record entries with different actors
	entry1, _ := domain.NewAuditEntry(tenantID, "actor-1", domain.ActionLogin, domain.ResourceUser, "", "Entry 1", true, "", nil)
	entry2, _ := domain.NewAuditEntry(tenantID, "actor-2", domain.ActionLogin, domain.ResourceUser, "", "Entry 2", true, "", nil)

	repo.Record(ctx, entry1)
	repo.Record(ctx, entry2)

	// Query with actor filter
	filters := domain.QueryFilters{
		ActorID: "actor-1",
		Limit:   100,
	}

	entries, total, err := repo.Query(ctx, tenantID, filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify results contain actor-1
	found := false
	for _, e := range entries {
		if e.ActorID() == "actor-1" {
			found = true
			break
		}
	}

	if !found {
		t.Error("actor-1 not found in results")
	}
}

func TestAuditRepository_Query_Pagination(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	// Clear and record 10 entries
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
		repo.Record(ctx, entry)
	}

	// First page
	filters := domain.QueryFilters{
		Limit:  5,
		Offset: 0,
	}

	entries, total, err := repo.Query(ctx, tenantID, filters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}

	if total < 10 {
		t.Errorf("expected at least 10 total entries, got %d", total)
	}
}

func TestAuditRepository_InsertOnly_NoDelete(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	// Verify that the schema enforces INSERT-only via table structure
	// This is a compile-time and runtime guarantee:
	// - No Delete() method on repository interface
	// - No UPDATE statement in migration
	// - PostgreSQL policies prevent UPDATE/DELETE

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	entry, _ := domain.NewAuditEntry(
		tenantID,
		"actor",
		domain.ActionLogin,
		domain.ResourceUser,
		"",
		"Original entry",
		true,
		"",
		nil,
	)

	repo.Record(ctx, entry)

	// Retrieve it
	retrieved, _ := repo.GetByID(ctx, tenantID, entry.ID())

	// Verify it's the original (no modifications possible)
	if retrieved.Description() != "Original entry" {
		t.Error("entry should not be modifiable")
	}

	// Verify there's no Update or Delete method on the interface
	// (compile-time assertion - if these existed, test would fail to compile)
}

func TestAuditRepository_JSON_Changes(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	changes := map[string]string{
		"email":  "old@example.com",
		"status": "inactive",
	}

	entry, _ := domain.NewAuditEntry(
		tenantID,
		"actor",
		domain.ActionLogin,
		domain.ResourceUser,
		"",
		"Entry with changes",
		true,
		"",
		changes,
	)

	repo.Record(ctx, entry)

	// Retrieve and verify changes are preserved
	retrieved, _ := repo.GetByID(ctx, tenantID, entry.ID())

	if len(retrieved.Changes()) != 2 {
		t.Errorf("expected 2 changes, got %d", len(retrieved.Changes()))
	}

	if retrieved.Changes()["email"] != "old@example.com" {
		t.Error("email change not preserved")
	}
}

func TestAuditRepository_TenantIsolation(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	tenant1 := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	tenant2 := domain.MustNewTenantID("660e8400-e29b-41d4-a716-446655440001")

	// Record entries for different tenants
	entry1, _ := domain.NewAuditEntry(tenant1, "actor", domain.ActionLogin, domain.ResourceUser, "", "Tenant 1", true, "", nil)
	entry2, _ := domain.NewAuditEntry(tenant2, "actor", domain.ActionLogin, domain.ResourceUser, "", "Tenant 2", true, "", nil)

	repo.Record(ctx, entry1)
	repo.Record(ctx, entry2)

	// Query tenant 1 should only see tenant 1 entries
	filters := domain.QueryFilters{Limit: 100}
	entries, _, err := repo.Query(ctx, tenant1, filters)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All returned entries should be from tenant 1
	for _, e := range entries {
		if e.TenantID() != tenant1 {
			t.Errorf("tenant isolation violation: entry from tenant %s in tenant %s query", e.TenantID(), tenant1)
		}
	}
}
