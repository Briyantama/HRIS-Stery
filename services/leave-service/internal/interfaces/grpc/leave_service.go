package grpc

import (
	"context"
	"fmt"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/leave/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LeaveServiceServer implements the gRPC Leave service.
type LeaveServiceServer struct {
	pb.UnimplementedLeaveServiceServer
	logger *zap.Logger
}

// NewLeaveServiceServer creates a new gRPC Leave service server.
func NewLeaveServiceServer(logger *zap.Logger) *LeaveServiceServer {
	return &LeaveServiceServer{
		logger: logger,
	}
}

// ApplyLeave creates a new leave request.
func (s *LeaveServiceServer) ApplyLeave(ctx context.Context, req *pb.ApplyLeaveRequest) (*pb.ApplyLeaveResponse, error) {
	// TODO: Implement after wiring command handler
	return nil, status.Errorf(codes.Unimplemented, "ApplyLeave not yet implemented")
}

// GetLeaveRequest retrieves a leave request.
func (s *LeaveServiceServer) GetLeaveRequest(ctx context.Context, req *pb.GetLeaveRequestRequest) (*pb.GetLeaveRequestResponse, error) {
	// TODO: Implement after wiring query handler
	return nil, status.Errorf(codes.Unimplemented, "GetLeaveRequest not yet implemented")
}

// ListLeaveRequests lists leave requests.
func (s *LeaveServiceServer) ListLeaveRequests(ctx context.Context, req *pb.ListLeaveRequestsRequest) (*pb.ListLeaveRequestsResponse, error) {
	// TODO: Implement after wiring query handler
	return nil, status.Errorf(codes.Unimplemented, "ListLeaveRequests not yet implemented")
}

// ApproveLeave approves a leave request.
func (s *LeaveServiceServer) ApproveLeave(ctx context.Context, req *pb.ApproveLeaveRequest) (*pb.ApproveLeaveResponse, error) {
	// TODO: Implement after wiring command handler
	return nil, status.Errorf(codes.Unimplemented, "ApproveLeave not yet implemented")
}

// RejectLeave rejects a leave request.
func (s *LeaveServiceServer) RejectLeave(ctx context.Context, req *pb.RejectLeaveRequest) (*pb.RejectLeaveResponse, error) {
	// TODO: Implement after wiring command handler
	return nil, status.Errorf(codes.Unimplemented, "RejectLeave not yet implemented")
}

// CancelLeave cancels a leave request.
func (s *LeaveServiceServer) CancelLeave(ctx context.Context, req *pb.CancelLeaveRequest) (*pb.CancelLeaveResponse, error) {
	// TODO: Implement after wiring command handler
	return nil, status.Errorf(codes.Unimplemented, "CancelLeave not yet implemented")
}

// GetLeaveBalance retrieves leave balance for an employee.
func (s *LeaveServiceServer) GetLeaveBalance(ctx context.Context, req *pb.GetLeaveBalanceRequest) (*pb.GetLeaveBalanceResponse, error) {
	// TODO: Implement after wiring query handler
	return nil, status.Errorf(codes.Unimplemented, "GetLeaveBalance not yet implemented")
}

// ListLeaveTypes lists all leave types for a tenant.
func (s *LeaveServiceServer) ListLeaveTypes(ctx context.Context, req *pb.ListLeaveTypesRequest) (*pb.ListLeaveTypesResponse, error) {
	// TODO: Implement after wiring query handler
	return nil, status.Errorf(codes.Unimplemented, "ListLeaveTypes not yet implemented")
}

// Helper: Extract tenant_id and actor_id from context
func extractContext(ctx context.Context) (tenantID, actorID string, err error) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok || tenantID == "" {
		return "", "", fmt.Errorf("tenant_id not found in context")
	}

	actorID, ok = ctx.Value("actor_id").(string)
	if !ok || actorID == "" {
		return "", "", fmt.Errorf("actor_id not found in context")
	}

	return tenantID, actorID, nil
}

// Helper: Convert domain status to proto enum
func statusToPB(s string) pb.LeaveStatus {
	switch s {
	case "PENDING":
		return pb.LeaveStatus_LEAVE_STATUS_PENDING
	case "APPROVED":
		return pb.LeaveStatus_LEAVE_STATUS_APPROVED
	case "REJECTED":
		return pb.LeaveStatus_LEAVE_STATUS_REJECTED
	case "CANCELLED":
		return pb.LeaveStatus_LEAVE_STATUS_CANCELLED
	default:
		return pb.LeaveStatus_LEAVE_STATUS_UNSPECIFIED
	}
}

// Helper: Convert proto enum to domain status string
func statusFromPB(s pb.LeaveStatus) string {
	switch s {
	case pb.LeaveStatus_LEAVE_STATUS_PENDING:
		return "PENDING"
	case pb.LeaveStatus_LEAVE_STATUS_APPROVED:
		return "APPROVED"
	case pb.LeaveStatus_LEAVE_STATUS_REJECTED:
		return "REJECTED"
	case pb.LeaveStatus_LEAVE_STATUS_CANCELLED:
		return "CANCELLED"
	default:
		return "PENDING"
	}
}

// Helper: Create a LeaveRequest proto from domain data
func leaveRequestToProto(id, tenantID, employeeID, leaveTypeID, status, startDate, endDate, reason, rejectionReason string, daysCount int32) *pb.LeaveRequest {
	resp := &pb.LeaveRequest{
		Id:              id,
		TenantId:        tenantID,
		EmployeeId:      employeeID,
		LeaveTypeId:     leaveTypeID,
		Status:          statusToPB(status),
		StartDate:       startDate,
		EndDate:         endDate,
		DaysCount:       daysCount,
		Reason:          reason,
		RejectionReason: rejectionReason,
		CreatedAt:       timestamppb.Now(),
		UpdatedAt:       timestamppb.Now(),
	}
	return resp
}
