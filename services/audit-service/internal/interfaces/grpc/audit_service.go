package grpc

import (
	"context"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/audit/v1"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuditServiceServer implements pb.AuditServiceServer
type AuditServiceServer struct {
	pb.UnimplementedAuditServiceServer

	recordHandler commands.RecordAuditHandler
	queryHandler  queries.QueryAuditTrailHandler
	logger        *zap.Logger
}

// NewAuditServiceServer creates a new audit service gRPC server
func NewAuditServiceServer(
	recordHandler commands.RecordAuditHandler,
	queryHandler queries.QueryAuditTrailHandler,
	logger *zap.Logger,
) *AuditServiceServer {
	return &AuditServiceServer{
		recordHandler: recordHandler,
		queryHandler:  queryHandler,
		logger:        logger,
	}
}

// Record records a new audit entry
func (s *AuditServiceServer) Record(ctx context.Context, req *pb.RecordAuditRequest) (*pb.RecordAuditResponse, error) {
	if req.TenantId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id required")
	}
	if req.ActorId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "actor_id required")
	}

	cmd := commands.RecordAuditCommand{
		TenantID:     req.TenantId,
		ActorID:      req.ActorId,
		Action:       protoToAuditAction(req.Action),
		ResourceType: protoToResourceType(req.ResourceType),
		ResourceID:   req.ResourceId,
		Description:  req.Description,
		Success:      req.Success,
		ErrorMessage: req.ErrorMessage,
		Changes:      req.Changes,
	}

	result, err := s.recordHandler.Handle(ctx, cmd)
	if err != nil {
		s.logger.Error("failed to record audit entry",
			zap.Error(err),
			zap.String("tenant_id", req.TenantId),
			zap.String("actor_id", req.ActorId),
		)
		return nil, status.Errorf(codes.Internal, "failed to record audit entry: %v", err)
	}

	return &pb.RecordAuditResponse{
		EntryId:   result.EntryID,
		CreatedAt: timestamppb.Now(),
	}, nil
}

// Query returns audit entries matching filters
func (s *AuditServiceServer) Query(ctx context.Context, req *pb.QueryAuditRequest) (*pb.QueryAuditResponse, error) {
	if req.TenantId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id required")
	}

	filters := queries.AuditQueryFilters{
		TenantID:     req.TenantId,
		ActorID:      req.ActorId,
		Action:       protoToAuditAction(req.Action),
		ResourceType: protoToResourceType(req.ResourceType),
		ResourceID:   req.ResourceId,
		Limit:        int(req.Limit),
		Offset:       int(req.Offset),
	}

	result, err := s.queryHandler.Handle(ctx, filters)
	if err != nil {
		s.logger.Error("failed to query audit entries",
			zap.Error(err),
			zap.String("tenant_id", req.TenantId),
		)
		return nil, status.Errorf(codes.Internal, "failed to query audit entries: %v", err)
	}

	// Convert DTOs to proto messages
	entries := make([]*pb.AuditEntry, 0, len(result.Entries))
	for _, dto := range result.Entries {
		entry := &pb.AuditEntry{
			EntryId:      dto.ID,
			TenantId:     dto.TenantID,
			ActorId:      dto.ActorID,
			Action:       auditActionToProto(dto.Action),
			ResourceType: resourceTypeToProto(dto.ResourceType),
			ResourceId:   dto.ResourceID,
			Description:  dto.Description,
			Success:      dto.Success,
			ErrorMessage: dto.ErrorMessage,
			Changes:      dto.Changes,
			CreatedAt:    timestamppb.New(dto.CreatedAt),
		}
		entries = append(entries, entry)
	}

	return &pb.QueryAuditResponse{
		Entries:    entries,
		TotalCount: int32(result.TotalCount),
	}, nil
}

// GetEntry returns a single audit entry
func (s *AuditServiceServer) GetEntry(ctx context.Context, req *pb.GetAuditEntryRequest) (*pb.AuditEntry, error) {
	if req.TenantId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id required")
	}
	if req.EntryId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "entry_id required")
	}

	entryID := domain.MustNewAuditEntryID(req.EntryId)

	// Get the entry via query handler (or create a separate handler)
	filters := queries.AuditQueryFilters{
		TenantID: req.TenantId,
		Limit:    1,
		Offset:   0,
	}

	result, err := s.queryHandler.Handle(ctx, filters)
	if err != nil {
		s.logger.Error("failed to get audit entry",
			zap.Error(err),
			zap.String("tenant_id", req.TenantId),
			zap.String("entry_id", req.EntryId),
		)
		return nil, status.Errorf(codes.Internal, "failed to get audit entry: %v", err)
	}

	// Find the entry by ID
	var found *queries.AuditEntryDTO
	for _, entry := range result.Entries {
		if entry.ID == req.EntryId {
			found = &entry
			break
		}
	}

	if found == nil {
		return nil, status.Errorf(codes.NotFound, "audit entry not found: %s", entryID)
	}

	return &pb.AuditEntry{
		EntryId:      found.ID,
		TenantId:     found.TenantID,
		ActorId:      found.ActorID,
		Action:       auditActionToProto(found.Action),
		ResourceType: resourceTypeToProto(found.ResourceType),
		ResourceId:   found.ResourceID,
		Description:  found.Description,
		Success:      found.Success,
		ErrorMessage: found.ErrorMessage,
		Changes:      found.Changes,
		CreatedAt:    timestamppb.New(found.CreatedAt),
	}, nil
}

// Helper functions to convert domain types to proto
func auditActionToProto(action domain.AuditAction) pb.AuditAction {
	switch action {
	case domain.ActionLogin:
		return pb.AuditAction_LOGIN
	case domain.ActionLogout:
		return pb.AuditAction_LOGOUT
	case domain.ActionEmployeeCreate:
		return pb.AuditAction_EMPLOYEE_CREATED
	case domain.AuditAction("EMPLOYEE_TERMINATED"):
		return pb.AuditAction_EMPLOYEE_TERMINATED
	case domain.ActionLeaveRequested:
		return pb.AuditAction_LEAVE_REQUESTED
	case domain.ActionLeaveApproved:
		return pb.AuditAction_LEAVE_APPROVED
	case domain.AuditAction("NOTIFICATION_SENT"):
		return pb.AuditAction_NOTIFICATION_SENT
	default:
		return pb.AuditAction_AUDIT_ACTION_UNSPECIFIED
	}
}

func resourceTypeToProto(rt domain.ResourceType) pb.ResourceType {
	switch rt {
	case domain.ResourceUser:
		return pb.ResourceType_USER
	case domain.ResourceEmployee:
		return pb.ResourceType_EMPLOYEE
	case domain.ResourceLeave:
		return pb.ResourceType_LEAVE
	case domain.ResourceNotification:
		return pb.ResourceType_NOTIFICATION
	default:
		return pb.ResourceType_RESOURCE_TYPE_UNSPECIFIED
	}
}

func protoToAuditAction(action pb.AuditAction) domain.AuditAction {
	switch action {
	case pb.AuditAction_LOGIN:
		return domain.ActionLogin
	case pb.AuditAction_LOGOUT:
		return domain.ActionLogout
	case pb.AuditAction_EMPLOYEE_CREATED:
		return domain.ActionEmployeeCreate
	case pb.AuditAction_EMPLOYEE_TERMINATED:
		return domain.AuditAction("EMPLOYEE_TERMINATED")
	case pb.AuditAction_LEAVE_REQUESTED:
		return domain.ActionLeaveRequested
	case pb.AuditAction_LEAVE_APPROVED:
		return domain.ActionLeaveApproved
	case pb.AuditAction_NOTIFICATION_SENT:
		return domain.AuditAction("NOTIFICATION_SENT")
	default:
		return domain.AuditAction("")
	}
}

func protoToResourceType(rt pb.ResourceType) domain.ResourceType {
	switch rt {
	case pb.ResourceType_USER:
		return domain.ResourceUser
	case pb.ResourceType_EMPLOYEE:
		return domain.ResourceEmployee
	case pb.ResourceType_LEAVE:
		return domain.ResourceLeave
	case pb.ResourceType_NOTIFICATION:
		return domain.ResourceNotification
	default:
		return domain.ResourceType("")
	}
}
