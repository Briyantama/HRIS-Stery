package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/leave/v1"
	"github.com/hris-stery/hris-stery/services/_shared/observability"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/application/queries"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LeaveServiceServer implements the gRPC Leave service.
type LeaveServiceServer struct {
	pb.UnimplementedLeaveServiceServer
	logger                     *zap.Logger
	createLeaveRequestHandler  *commands.CreateLeaveRequestHandler
	approveLeaveRequestHandler *commands.ApproveLeaveRequestHandler
	rejectLeaveRequestHandler  *commands.RejectLeaveRequestHandler
	cancelLeaveRequestHandler  *commands.CancelLeaveRequestHandler
	getLeaveRequestHandler     *queries.GetLeaveRequestHandler
	listLeaveRequestsHandler   *queries.ListLeaveRequestsHandler
	getLeaveBalanceHandler     *queries.GetLeaveBalanceHandler
	listLeaveTypesHandler      *queries.ListLeaveTypesHandler
}

// NewLeaveServiceServer creates a new gRPC Leave service server.
func NewLeaveServiceServer(
	logger *zap.Logger,
	createLeaveRequestHandler *commands.CreateLeaveRequestHandler,
	approveLeaveRequestHandler *commands.ApproveLeaveRequestHandler,
	rejectLeaveRequestHandler *commands.RejectLeaveRequestHandler,
	cancelLeaveRequestHandler *commands.CancelLeaveRequestHandler,
	getLeaveRequestHandler *queries.GetLeaveRequestHandler,
	listLeaveRequestsHandler *queries.ListLeaveRequestsHandler,
	getLeaveBalanceHandler *queries.GetLeaveBalanceHandler,
	listLeaveTypesHandler *queries.ListLeaveTypesHandler,
) *LeaveServiceServer {
	return &LeaveServiceServer{
		logger:                     logger,
		createLeaveRequestHandler:  createLeaveRequestHandler,
		approveLeaveRequestHandler: approveLeaveRequestHandler,
		rejectLeaveRequestHandler:  rejectLeaveRequestHandler,
		cancelLeaveRequestHandler:  cancelLeaveRequestHandler,
		getLeaveRequestHandler:     getLeaveRequestHandler,
		listLeaveRequestsHandler:   listLeaveRequestsHandler,
		getLeaveBalanceHandler:     getLeaveBalanceHandler,
		listLeaveTypesHandler:      listLeaveTypesHandler,
	}
}

// ApplyLeave creates a new leave request.
func (s *LeaveServiceServer) ApplyLeave(ctx context.Context, req *pb.ApplyLeaveRequest) (*pb.ApplyLeaveResponse, error) {
	if req == nil || req.EmployeeId == "" || req.LeaveTypeId == "" || req.StartDate == "" || req.EndDate == "" {
		return nil, status.Error(codes.InvalidArgument, "employee_id, leave_type_id, start_date, and end_date are required")
	}

	tenantID := observability.TenantIDFromContext(ctx)
	if tenantID == "" && req.TenantId != "" {
		tenantID = req.TenantId
	}
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if req.TenantId != "" && req.TenantId != tenantID {
		return nil, status.Error(codes.PermissionDenied, "tenant_id mismatch")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start_date format: %v", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid end_date format: %v", err)
	}

	actorID := observability.UserIDFromContext(ctx)
	if actorID == "" {
		actorID = "system"
	}

	result, err := s.createLeaveRequestHandler.Handle(ctx, commands.CreateLeaveRequestCommand{
		TenantID:    tenantID,
		EmployeeID:  req.EmployeeId,
		LeaveTypeID: req.LeaveTypeId,
		StartDate:   startDate,
		EndDate:     endDate,
		Reason:      req.Reason,
		ActorID:     actorID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return nil, status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
		}
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "resource not found: %v", err)
		}
		if strings.Contains(err.Error(), "insufficient balance") {
			return nil, status.Errorf(codes.FailedPrecondition, "insufficient leave balance")
		}
		return nil, status.Errorf(codes.Internal, "create leave request failed: %v", err)
	}

	// Fetch the created leave request for the response
	leaveReq, err := s.getLeaveRequestHandler.Handle(ctx, queries.GetLeaveRequestQuery{
		TenantID:       tenantID,
		LeaveRequestID: result.LeaveRequestID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch created leave request: %v", err)
	}

	return &pb.ApplyLeaveResponse{
		LeaveRequest: mapLeaveRequestDTOToProto(leaveReq),
	}, nil
}

// GetLeaveRequest retrieves a leave request.
func (s *LeaveServiceServer) GetLeaveRequest(ctx context.Context, req *pb.GetLeaveRequestRequest) (*pb.GetLeaveRequestResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	tenantID := observability.TenantIDFromContext(ctx)
	if tenantID == "" && req.TenantId != "" {
		tenantID = req.TenantId
	}
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if req.TenantId != "" && req.TenantId != tenantID {
		return nil, status.Error(codes.PermissionDenied, "tenant_id mismatch")
	}

	result, err := s.getLeaveRequestHandler.Handle(ctx, queries.GetLeaveRequestQuery{
		TenantID:       tenantID,
		LeaveRequestID: req.Id,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "leave request not found")
		}
		return nil, status.Errorf(codes.Internal, "fetch leave request: %v", err)
	}

	return &pb.GetLeaveRequestResponse{
		LeaveRequest: mapLeaveRequestDTOToProto(result),
	}, nil
}

// ListLeaveRequests lists leave requests.
func (s *LeaveServiceServer) ListLeaveRequests(ctx context.Context, req *pb.ListLeaveRequestsRequest) (*pb.ListLeaveRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	tenantID := observability.TenantIDFromContext(ctx)
	if tenantID == "" && req.TenantId != "" {
		tenantID = req.TenantId
	}
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if req.TenantId != "" && req.TenantId != tenantID {
		return nil, status.Error(codes.PermissionDenied, "tenant_id mismatch")
	}

	var year *int
	if req.Year > 0 {
		y := int(req.Year)
		year = &y
	}

	var employeeID *string
	if req.EmployeeId != "" {
		employeeID = &req.EmployeeId
	}

	var approverID *string
	if req.ApproverId != "" {
		approverID = &req.ApproverId
	}

	var statusFilter *string
	if req.Status != pb.LeaveStatus_LEAVE_STATUS_UNSPECIFIED {
		statusStr := req.Status.String()
		statusFilter = &statusStr
	}

	result, err := s.listLeaveRequestsHandler.Handle(ctx, queries.ListLeaveRequestsQuery{
		TenantID:     tenantID,
		EmployeeID:   employeeID,
		ApprovedByID: approverID,
		Status:       statusFilter,
		Year:         year,
		Offset:       0,
		Limit:        int(req.PageSize),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list leave requests: %v", err)
	}

	leaves := make([]*pb.LeaveRequest, len(result.Records))
	for i, lr := range result.Records {
		leaves[i] = mapLeaveRequestDTOToProto(lr)
	}

	return &pb.ListLeaveRequestsResponse{
		LeaveRequests: leaves,
		TotalCount:    int32(result.Total),
		NextPageToken: "",
	}, nil
}

// ApproveLeave approves a leave request.
func (s *LeaveServiceServer) ApproveLeave(ctx context.Context, req *pb.ApproveLeaveRequest) (*pb.ApproveLeaveResponse, error) {
	if req == nil || req.Id == "" || req.ApproverId == "" {
		return nil, status.Error(codes.InvalidArgument, "id and approver_id are required")
	}

	tenantID := observability.TenantIDFromContext(ctx)
	if tenantID == "" && req.TenantId != "" {
		tenantID = req.TenantId
	}
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if req.TenantId != "" && req.TenantId != tenantID {
		return nil, status.Error(codes.PermissionDenied, "tenant_id mismatch")
	}

	// Fine-grained RBAC: only managers and hr_admins can approve leave
	if err := requireApprovalPermission(ctx); err != nil {
		return nil, err
	}

	actorID := observability.UserIDFromContext(ctx)
	if actorID == "" {
		actorID = req.ApproverId
	}

	result, err := s.approveLeaveRequestHandler.Handle(ctx, commands.ApproveLeaveRequestCommand{
		TenantID:       tenantID,
		LeaveRequestID: req.Id,
		ApprovedByID:   req.ApproverId,
		ActorID:        actorID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "leave request not found")
		}
		if strings.Contains(err.Error(), "invalid") {
			return nil, status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
		}
		if strings.Contains(err.Error(), "already") {
			return nil, status.Errorf(codes.FailedPrecondition, "leave request already processed")
		}
		return nil, status.Errorf(codes.Internal, "approve leave request: %v", err)
	}

	// Fetch the updated leave request for the response
	leaveReq, err := s.getLeaveRequestHandler.Handle(ctx, queries.GetLeaveRequestQuery{
		TenantID:       tenantID,
		LeaveRequestID: result.LeaveRequestID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch updated leave request: %v", err)
	}

	return &pb.ApproveLeaveResponse{
		LeaveRequest: mapLeaveRequestDTOToProto(leaveReq),
	}, nil
}

// RejectLeave rejects a leave request.
func (s *LeaveServiceServer) RejectLeave(ctx context.Context, req *pb.RejectLeaveRequest) (*pb.RejectLeaveResponse, error) {
	if req == nil || req.Id == "" || req.ApproverId == "" {
		return nil, status.Error(codes.InvalidArgument, "id and approver_id are required")
	}

	tenantID := observability.TenantIDFromContext(ctx)
	if tenantID == "" && req.TenantId != "" {
		tenantID = req.TenantId
	}
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if req.TenantId != "" && req.TenantId != tenantID {
		return nil, status.Error(codes.PermissionDenied, "tenant_id mismatch")
	}

	// Fine-grained RBAC: only managers and hr_admins can reject leave
	if err := requireApprovalPermission(ctx); err != nil {
		return nil, err
	}

	actorID := observability.UserIDFromContext(ctx)
	if actorID == "" {
		actorID = req.ApproverId
	}

	result, err := s.rejectLeaveRequestHandler.Handle(ctx, commands.RejectLeaveRequestCommand{
		TenantID:       tenantID,
		LeaveRequestID: req.Id,
		Reason:         req.RejectionReason,
		ActorID:        actorID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "leave request not found")
		}
		if strings.Contains(err.Error(), "invalid") {
			return nil, status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
		}
		if strings.Contains(err.Error(), "already") {
			return nil, status.Errorf(codes.FailedPrecondition, "leave request already processed")
		}
		return nil, status.Errorf(codes.Internal, "reject leave request: %v", err)
	}

	// Fetch the updated leave request for the response
	leaveReq, err := s.getLeaveRequestHandler.Handle(ctx, queries.GetLeaveRequestQuery{
		TenantID:       tenantID,
		LeaveRequestID: result.LeaveRequestID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch updated leave request: %v", err)
	}

	return &pb.RejectLeaveResponse{
		LeaveRequest: mapLeaveRequestDTOToProto(leaveReq),
	}, nil
}

// CancelLeave cancels a leave request.
func (s *LeaveServiceServer) CancelLeave(ctx context.Context, req *pb.CancelLeaveRequest) (*pb.CancelLeaveResponse, error) {
	if req == nil || req.Id == "" || req.EmployeeId == "" {
		return nil, status.Error(codes.InvalidArgument, "id and employee_id are required")
	}

	tenantID := observability.TenantIDFromContext(ctx)
	if tenantID == "" && req.TenantId != "" {
		tenantID = req.TenantId
	}
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if req.TenantId != "" && req.TenantId != tenantID {
		return nil, status.Error(codes.PermissionDenied, "tenant_id mismatch")
	}

	// Fetch the leave request early to verify ownership for cancellation
	leaveReqDTO, err := s.getLeaveRequestHandler.Handle(ctx, queries.GetLeaveRequestQuery{
		TenantID:       tenantID,
		LeaveRequestID: req.Id,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "leave request not found")
		}
		return nil, status.Errorf(codes.Internal, "fetch leave request: %v", err)
	}

	// Fine-grained RBAC: employees can only cancel their own leave, admins can cancel any
	if err := verifyCancelPermission(ctx, leaveReqDTO.EmployeeID); err != nil {
		return nil, err
	}

	actorID := observability.UserIDFromContext(ctx)
	if actorID == "" {
		actorID = req.EmployeeId
	}

	result, err := s.cancelLeaveRequestHandler.Handle(ctx, commands.CancelLeaveRequestCommand{
		TenantID:       tenantID,
		LeaveRequestID: req.Id,
		Reason:         fmt.Sprintf("Cancelled by employee %s", req.EmployeeId),
		ActorID:        actorID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "leave request not found")
		}
		if strings.Contains(err.Error(), "invalid") {
			return nil, status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
		}
		if strings.Contains(err.Error(), "cannot") || strings.Contains(err.Error(), "already") {
			return nil, status.Errorf(codes.FailedPrecondition, "cannot cancel leave request: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "cancel leave request: %v", err)
	}

	// Fetch the updated leave request for the response
	leaveReq, err := s.getLeaveRequestHandler.Handle(ctx, queries.GetLeaveRequestQuery{
		TenantID:       tenantID,
		LeaveRequestID: result.LeaveRequestID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch updated leave request: %v", err)
	}

	return &pb.CancelLeaveResponse{
		LeaveRequest: mapLeaveRequestDTOToProto(leaveReq),
	}, nil
}

// GetLeaveBalance retrieves leave balance for an employee.
func (s *LeaveServiceServer) GetLeaveBalance(ctx context.Context, req *pb.GetLeaveBalanceRequest) (*pb.GetLeaveBalanceResponse, error) {
	if req == nil || req.EmployeeId == "" {
		return nil, status.Error(codes.InvalidArgument, "employee_id is required")
	}

	tenantID := observability.TenantIDFromContext(ctx)
	if tenantID == "" && req.TenantId != "" {
		tenantID = req.TenantId
	}
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if req.TenantId != "" && req.TenantId != tenantID {
		return nil, status.Error(codes.PermissionDenied, "tenant_id mismatch")
	}

	year := int(req.Year)
	if year == 0 {
		year = time.Now().Year()
	}

	result, err := s.getLeaveBalanceHandler.Handle(ctx, queries.GetLeaveBalanceQuery{
		TenantID:   tenantID,
		EmployeeID: req.EmployeeId,
		Year:       year,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "leave balance not found")
		}
		return nil, status.Errorf(codes.Internal, "fetch leave balance: %v", err)
	}

	return &pb.GetLeaveBalanceResponse{
		Balance: mapLeaveBalanceDTOToProto(result),
	}, nil
}

// ListLeaveTypes lists all leave types for a tenant.
func (s *LeaveServiceServer) ListLeaveTypes(ctx context.Context, req *pb.ListLeaveTypesRequest) (*pb.ListLeaveTypesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	tenantID := observability.TenantIDFromContext(ctx)
	if tenantID == "" && req.TenantId != "" {
		tenantID = req.TenantId
	}
	if tenantID == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if req.TenantId != "" && req.TenantId != tenantID {
		return nil, status.Error(codes.PermissionDenied, "tenant_id mismatch")
	}

	result, err := s.listLeaveTypesHandler.Handle(ctx, queries.ListLeaveTypesQuery{
		TenantID: tenantID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch leave types: %v", err)
	}

	types := make([]*pb.LeaveType, len(result.LeaveTypes))
	for i, lt := range result.LeaveTypes {
		types[i] = &pb.LeaveType{
			Id:               lt.ID,
			TenantId:         tenantID,
			Code:             pb.LeaveTypeCode(pb.LeaveTypeCode_value[lt.Code]),
			Name:             lt.Name,
			MaxDaysPerYear:   int32(lt.MaxDaysPerYear),
			RequiresDocument: lt.RequiresDocument,
			IsPaid:           lt.IsPaid,
		}
	}

	return &pb.ListLeaveTypesResponse{
		LeaveTypes: types,
	}, nil
}

// mapLeaveRequestDTOToProto converts a LeaveRequestDTO to a proto LeaveRequest.
func mapLeaveRequestDTOToProto(dto *queries.LeaveRequestDTO) *pb.LeaveRequest {
	pbStatus := pb.LeaveStatus_LEAVE_STATUS_UNSPECIFIED
	switch dto.Status {
	case "PENDING":
		pbStatus = pb.LeaveStatus_LEAVE_STATUS_PENDING
	case "APPROVED":
		pbStatus = pb.LeaveStatus_LEAVE_STATUS_APPROVED
	case "REJECTED":
		pbStatus = pb.LeaveStatus_LEAVE_STATUS_REJECTED
	case "CANCELLED":
		pbStatus = pb.LeaveStatus_LEAVE_STATUS_CANCELLED
	}

	approverID := ""
	if dto.ApprovedByID != nil {
		approverID = *dto.ApprovedByID
	}

	lr := &pb.LeaveRequest{
		Id:              dto.ID,
		TenantId:        dto.TenantID,
		EmployeeId:      dto.EmployeeID,
		LeaveTypeId:     dto.LeaveTypeID,
		Status:          pbStatus,
		StartDate:       dto.StartDate,
		EndDate:         dto.EndDate,
		DaysCount:       int32(dto.DaysCount),
		Reason:          dto.Reason,
		RejectionReason: dto.RejectionReason,
		ApprovedById:    approverID,
		CreatedAt:       parseTimestamp(dto.CreatedAt),
		UpdatedAt:       parseTimestamp(dto.UpdatedAt),
	}

	if dto.ApprovedAt != nil {
		lr.ApprovedAt = parseTimestamp(*dto.ApprovedAt)
	}

	return lr
}

// mapLeaveBalanceDTOToProto converts a LeaveBalanceResult to a proto LeaveBalance.
func mapLeaveBalanceDTOToProto(dto *queries.LeaveBalanceResult) *pb.LeaveBalance {
	balances := make([]*pb.LeaveTypeBalance, len(dto.ByType))
	for i, b := range dto.ByType {
		balances[i] = &pb.LeaveTypeBalance{
			LeaveTypeId:   b.LeaveTypeID,
			EntitledDays:  int32(b.EntitledDays),
			UsedDays:      int32(b.UsedDays),
			PendingDays:   int32(b.PendingDays),
			RemainingDays: int32(b.RemainingDays),
		}
	}

	return &pb.LeaveBalance{
		TenantId:   dto.TenantID,
		EmployeeId: dto.EmployeeID,
		Year:       int32(dto.Year),
		Balances:   balances,
	}
}

// parseTimestamp parses an RFC3339 timestamp string to protobuf timestamp.
func parseTimestamp(ts string) *timestamppb.Timestamp {
	if ts == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t, _ = time.Parse("2006-01-02T15:04:05Z", ts)
	}
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
