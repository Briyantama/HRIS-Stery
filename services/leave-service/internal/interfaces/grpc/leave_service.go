package grpc

import (
	"context"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/leave/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
