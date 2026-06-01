package grpc

import (
	"context"
	"testing"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/audit/v1"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockHandlers for gRPC tests
type MockRecordHandler struct {
	called  bool
	lastCmd commands.RecordAuditCommand
}

func (m *MockRecordHandler) Handle(ctx context.Context, cmd commands.RecordAuditCommand) (*commands.RecordAuditResult, error) {
	m.called = true
	m.lastCmd = cmd
	return &commands.RecordAuditResult{EntryID: "550e8400-e29b-41d4-a716-446655440000"}, nil
}

type MockQueryHandler struct {
	called bool
}

func (m *MockQueryHandler) Handle(ctx context.Context, filters queries.AuditQueryFilters) (*queries.QueryAuditTrailResult, error) {
	m.called = true
	return &queries.QueryAuditTrailResult{
		Entries: []queries.AuditEntryDTO{
			{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				TenantID:     filters.TenantID,
				ActorID:      "actor-123",
				Action:       domain.ActionLogin,
				ResourceType: domain.ResourceUser,
				ResourceID:   "user-456",
				Description:  "Test entry",
				Success:      true,
				Changes:      map[string]string{},
			},
		},
		TotalCount: 1,
	}, nil
}

func TestAuditServiceServer_Record(t *testing.T) {
	logger, _ := zap.NewProduction()
	recordHandler := &MockRecordHandler{}
	queryHandler := &MockQueryHandler{}

	server := NewAuditServiceServer(recordHandler, queryHandler, logger)

	req := &pb.RecordAuditRequest{
		TenantId:     "550e8400-e29b-41d4-a716-446655440000",
		ActorId:      "actor-123",
		Action:       pb.AuditAction_LOGIN,
		ResourceType: pb.ResourceType_USER,
		ResourceId:   "user-456",
		Description:  "User logged in",
		Success:      true,
		Changes:      map[string]string{},
	}

	resp, err := server.Record(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !recordHandler.called {
		t.Error("record handler was not called")
	}

	if resp.EntryId == "" {
		t.Error("response entry_id should not be empty")
	}

	if resp.CreatedAt == nil {
		t.Error("response created_at should not be nil")
	}
}

func TestAuditServiceServer_Record_MissingTenantID(t *testing.T) {
	logger, _ := zap.NewProduction()
	recordHandler := &MockRecordHandler{}
	queryHandler := &MockQueryHandler{}

	server := NewAuditServiceServer(recordHandler, queryHandler, logger)

	req := &pb.RecordAuditRequest{
		TenantId: "", // missing
		ActorId:  "actor-123",
	}

	_, err := server.Record(context.Background(), req)

	if err == nil {
		t.Error("expected error for missing tenant_id")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestAuditServiceServer_Record_MissingActorID(t *testing.T) {
	logger, _ := zap.NewProduction()
	recordHandler := &MockRecordHandler{}
	queryHandler := &MockQueryHandler{}

	server := NewAuditServiceServer(recordHandler, queryHandler, logger)

	req := &pb.RecordAuditRequest{
		TenantId: "550e8400-e29b-41d4-a716-446655440000",
		ActorId:  "", // missing
	}

	_, err := server.Record(context.Background(), req)

	if err == nil {
		t.Error("expected error for missing actor_id")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestAuditServiceServer_Query(t *testing.T) {
	logger, _ := zap.NewProduction()
	recordHandler := &MockRecordHandler{}
	queryHandler := &MockQueryHandler{}

	server := NewAuditServiceServer(recordHandler, queryHandler, logger)

	req := &pb.QueryAuditRequest{
		TenantId: "550e8400-e29b-41d4-a716-446655440000",
		Limit:    100,
		Offset:   0,
	}

	resp, err := server.Query(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !queryHandler.called {
		t.Error("query handler was not called")
	}

	if len(resp.Entries) == 0 {
		t.Error("response should contain entries")
	}

	if resp.TotalCount != 1 {
		t.Errorf("expected total count 1, got %d", resp.TotalCount)
	}
}

func TestAuditServiceServer_Query_MissingTenantID(t *testing.T) {
	logger, _ := zap.NewProduction()
	recordHandler := &MockRecordHandler{}
	queryHandler := &MockQueryHandler{}

	server := NewAuditServiceServer(recordHandler, queryHandler, logger)

	req := &pb.QueryAuditRequest{
		TenantId: "", // missing
	}

	_, err := server.Query(context.Background(), req)

	if err == nil {
		t.Error("expected error for missing tenant_id")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestAuditServiceServer_GetEntry(t *testing.T) {
	logger, _ := zap.NewProduction()
	recordHandler := &MockRecordHandler{}
	queryHandler := &MockQueryHandler{}

	server := NewAuditServiceServer(recordHandler, queryHandler, logger)

	req := &pb.GetAuditEntryRequest{
		TenantId: "550e8400-e29b-41d4-a716-446655440000",
		EntryId:  "550e8400-e29b-41d4-a716-446655440000",
	}

	resp, err := server.GetEntry(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.EntryId == "" {
		t.Error("response entry_id should not be empty")
	}

	if resp.ActorId != "actor-123" {
		t.Errorf("expected actor-123, got %s", resp.ActorId)
	}
}

func TestAuditServiceServer_GetEntry_MissingTenantID(t *testing.T) {
	logger, _ := zap.NewProduction()
	recordHandler := &MockRecordHandler{}
	queryHandler := &MockQueryHandler{}

	server := NewAuditServiceServer(recordHandler, queryHandler, logger)

	req := &pb.GetAuditEntryRequest{
		TenantId: "", // missing
		EntryId:  "550e8400-e29b-41d4-a716-446655440000",
	}

	_, err := server.GetEntry(context.Background(), req)

	if err == nil {
		t.Error("expected error for missing tenant_id")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestAuditServiceServer_GetEntry_MissingEntryID(t *testing.T) {
	logger, _ := zap.NewProduction()
	recordHandler := &MockRecordHandler{}
	queryHandler := &MockQueryHandler{}

	server := NewAuditServiceServer(recordHandler, queryHandler, logger)

	req := &pb.GetAuditEntryRequest{
		TenantId: "550e8400-e29b-41d4-a716-446655440000",
		EntryId:  "", // missing
	}

	_, err := server.GetEntry(context.Background(), req)

	if err == nil {
		t.Error("expected error for missing entry_id")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestAuditServiceServer_TypeConversions(t *testing.T) {
	// Test proto ↔ domain type conversions
	testCases := []struct {
		protoAction    pb.AuditAction
		domainAction   domain.AuditAction
		protoResource  pb.ResourceType
		domainResource domain.ResourceType
	}{
		{pb.AuditAction_LOGIN, domain.ActionLogin, pb.ResourceType_USER, domain.ResourceUser},
		{pb.AuditAction_LOGOUT, domain.ActionLogout, pb.ResourceType_EMPLOYEE, domain.ResourceEmployee},
		{pb.AuditAction_EMPLOYEE_CREATED, domain.ActionEmployeeCreate, pb.ResourceType_LEAVE, domain.ResourceLeave},
		{pb.AuditAction_LEAVE_REQUESTED, domain.ActionLeaveRequested, pb.ResourceType_NOTIFICATION, domain.ResourceNotification},
		{pb.AuditAction_LEAVE_APPROVED, domain.ActionLeaveApproved, pb.ResourceType_USER, domain.ResourceUser},
	}

	for _, tc := range testCases {
		// Test proto to domain
		domainActionResult := protoToAuditAction(tc.protoAction)
		if domainActionResult != tc.domainAction {
			t.Errorf("action conversion failed: %v → %v (expected %v)", tc.protoAction, domainActionResult, tc.domainAction)
		}

		// Test domain to proto
		protoActionResult := auditActionToProto(tc.domainAction)
		if protoActionResult != tc.protoAction {
			t.Errorf("action conversion failed: %v → %v (expected %v)", tc.domainAction, protoActionResult, tc.protoAction)
		}

		// Test resource type conversions
		domainResourceResult := protoToResourceType(tc.protoResource)
		if domainResourceResult != tc.domainResource {
			t.Errorf("resource type conversion failed: %v → %v (expected %v)", tc.protoResource, domainResourceResult, tc.domainResource)
		}

		protoResourceResult := resourceTypeToProto(tc.domainResource)
		if protoResourceResult != tc.protoResource {
			t.Errorf("resource type conversion failed: %v → %v (expected %v)", tc.domainResource, protoResourceResult, tc.protoResource)
		}
	}
}
