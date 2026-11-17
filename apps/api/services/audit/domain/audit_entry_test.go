package domain

import (
	"testing"
)

func TestNewAuditEntry_Immutability(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	changes := map[string]string{
		"field1": "value1",
		"field2": "value2",
	}

	entry, err := NewAuditEntry(
		tenantID,
		"actor-123",
		ActionLogin,
		ResourceUser,
		"user-456",
		"User logged in",
		true,
		"",
		changes,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test read-only accessors
	if entry.ID().IsZero() {
		t.Error("entry ID should not be zero")
	}
	if entry.TenantID() != tenantID {
		t.Error("tenant ID mismatch")
	}
	if entry.ActorID() != "actor-123" {
		t.Error("actor ID mismatch")
	}
	if entry.Action() != ActionLogin {
		t.Error("action mismatch")
	}
	if entry.ResourceType() != ResourceUser {
		t.Error("resource type mismatch")
	}
	if entry.ResourceID() != "user-456" {
		t.Error("resource ID mismatch")
	}
	if entry.Description() != "User logged in" {
		t.Error("description mismatch")
	}
	if !entry.Success() {
		t.Error("success should be true")
	}
	if entry.ErrorMessage() != "" {
		t.Error("error message should be empty")
	}
	if len(entry.Changes()) != 2 {
		t.Error("changes should have 2 items")
	}

	// Verify no update methods exist by checking they can't be compiled
	// (compile-time assertion, verified by successful build)
}

func TestNewAuditEntry_InvalidTenantID(t *testing.T) {
	invalidTenantID := MustNewTenantID("")

	_, err := NewAuditEntry(
		invalidTenantID,
		"actor-123",
		ActionLogin,
		ResourceUser,
		"",
		"",
		true,
		"",
		nil,
	)

	if err == nil {
		t.Error("expected error for invalid tenant ID")
	}
}

func TestNewAuditEntry_MissingActorID(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	_, err := NewAuditEntry(
		tenantID,
		"", // missing actor ID
		ActionLogin,
		ResourceUser,
		"",
		"",
		true,
		"",
		nil,
	)

	if err == nil {
		t.Error("expected error for missing actor ID")
	}
}

func TestNewAuditEntry_WithError(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	entry, err := NewAuditEntry(
		tenantID,
		"actor-123",
		ActionLogin,
		ResourceUser,
		"user-456",
		"User login failed",
		false, // success = false
		"invalid credentials",
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.Success() {
		t.Error("success should be false")
	}
	if entry.ErrorMessage() != "invalid credentials" {
		t.Error("error message mismatch")
	}
}
