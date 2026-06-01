package grpc

import (
	"context"
	"time"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/attendance/v1"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/application/queries"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AttendanceServiceServer implements the gRPC attendance service.
type AttendanceServiceServer struct {
	pb.UnimplementedAttendanceServiceServer
	checkInHandler         *commands.CheckInHandler
	checkOutHandler        *commands.CheckOutHandler
	overrideHandler        *commands.OverrideAttendanceHandler
	getDailySummaryHandler *queries.GetDailySummaryHandler
	listAttendanceHandler  *queries.ListAttendanceHandler
}

// NewAttendanceServiceServer creates a new attendance service server.
func NewAttendanceServiceServer(
	checkInHandler *commands.CheckInHandler,
	checkOutHandler *commands.CheckOutHandler,
	overrideHandler *commands.OverrideAttendanceHandler,
	getAttendanceHandler *queries.GetAttendanceHandler,
	listAttendanceHandler *queries.ListAttendanceHandler,
	getDailySummaryHandler *queries.GetDailySummaryHandler,
) *AttendanceServiceServer {
	return &AttendanceServiceServer{
		checkInHandler:         checkInHandler,
		checkOutHandler:        checkOutHandler,
		overrideHandler:        overrideHandler,
		getDailySummaryHandler: getDailySummaryHandler,
		listAttendanceHandler:  listAttendanceHandler,
	}
}

// CheckIn handles check-in requests (uses current date and time).
func (s *AttendanceServiceServer) CheckIn(ctx context.Context, req *pb.CheckInRequest) (*pb.CheckInResponse, error) {
	now := time.Now()
	cmd := commands.CheckInCommand{
		TenantID:   req.TenantId,
		EmployeeID: req.EmployeeId,
		Date:       now,
		CheckInAt:  now,
		Latitude:   latPtrFromProto(req.Latitude),
		Longitude:  lonPtrFromProto(req.Longitude),
		Notes:      req.Notes,
		ActorID:    req.EmployeeId,
	}

	result, err := s.checkInHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check in failed: %v", err)
	}

	return &pb.CheckInResponse{
		Record: &pb.AttendanceRecord{
			Id:         result.AttendanceID,
			TenantId:   result.TenantID,
			EmployeeId: result.EmployeeID,
			CheckInAt:  timestamppb.New(result.CheckInAt),
			Status:     statusToEnum(result.Status),
		},
	}, nil
}

// CheckOut handles check-out requests (uses current date and time).
func (s *AttendanceServiceServer) CheckOut(ctx context.Context, req *pb.CheckOutRequest) (*pb.CheckOutResponse, error) {
	now := time.Now()
	cmd := commands.CheckOutCommand{
		TenantID:   req.TenantId,
		EmployeeID: req.EmployeeId,
		Date:       now,
		CheckOutAt: now,
		Latitude:   latPtrFromProto(req.Latitude),
		Longitude:  lonPtrFromProto(req.Longitude),
		Notes:      req.Notes,
		ActorID:    req.EmployeeId,
	}

	result, err := s.checkOutHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check out failed: %v", err)
	}

	return &pb.CheckOutResponse{
		Record: &pb.AttendanceRecord{
			Id:               result.AttendanceID,
			TenantId:         result.TenantID,
			EmployeeId:       result.EmployeeID,
			CheckOutAt:       timestamppb.New(result.CheckOutAt),
			WorkDurationMins: formatWorkDuration(result.WorkDurationMins),
			Status:           statusToEnum(result.Status),
		},
	}, nil
}

// OverrideAttendance handles attendance override requests.
func (s *AttendanceServiceServer) OverrideAttendance(ctx context.Context, req *pb.OverrideAttendanceRequest) (*pb.OverrideAttendanceResponse, error) {
	dateStr := req.Date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	cmd := commands.OverrideAttendanceCommand{
		TenantID:   req.TenantId,
		EmployeeID: req.EmployeeId,
		Date:       date,
		CheckInAt:  tsToPtr(req.CheckInAt),
		CheckOutAt: tsToPtr(req.CheckOutAt),
		Status:     statusFromEnum(req.Status),
		Reason:     req.Reason,
		ActorID:    req.AdminId,
	}

	result, err := s.overrideHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "override failed: %v", err)
	}

	return &pb.OverrideAttendanceResponse{
		Record: &pb.AttendanceRecord{
			Id:         result.AttendanceID,
			TenantId:   result.TenantID,
			EmployeeId: result.EmployeeID,
			Status:     statusToEnum(result.Status),
		},
	}, nil
}

// GetAttendanceSummary retrieves attendance summary for a period.
func (s *AttendanceServiceServer) GetAttendanceSummary(ctx context.Context, req *pb.GetAttendanceSummaryRequest) (*pb.GetAttendanceSummaryResponse, error) {
	now := time.Now()
	query := queries.GetDailySummaryQuery{
		TenantID: req.TenantId,
		Date:     now,
	}

	result, err := s.getDailySummaryHandler.Handle(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get summary failed: %v", err)
	}

	return &pb.GetAttendanceSummaryResponse{
		Summary: &pb.AttendanceSummary{
			TenantId:         result.TenantID,
			EmployeeId:       req.EmployeeId,
			Period:           req.Period,
			TotalWorkingDays: int32(result.TotalEmployees),
			PresentDays:      int32(result.PresentCount),
			AbsentDays:       int32(result.AbsentCount),
			LateDays:         int32(result.LateCount),
			OnLeaveDays:      int32(result.OnLeaveCount),
			TotalWorkMinutes: int32(result.AverageWorkMins),
		},
	}, nil
}

// ListAttendance retrieves attendance records with optional filters.
func (s *AttendanceServiceServer) ListAttendance(ctx context.Context, req *pb.ListAttendanceRequest) (*pb.ListAttendanceResponse, error) {
	query := queries.ListAttendanceQuery{
		TenantID:   req.TenantId,
		EmployeeID: strPtrFromProto(req.EmployeeId),
		Offset:     0,
		Limit:      int(req.PageSize),
	}

	result, err := s.listAttendanceHandler.Handle(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list attendance failed: %v", err)
	}

	records := make([]*pb.AttendanceRecord, len(result.Records))
	for i, dto := range result.Records {
		records[i] = dtoToProto(dto)
	}

	return &pb.ListAttendanceResponse{
		Records:       records,
		TotalCount:    int32(result.Total),
		NextPageToken: "",
	}, nil
}

// Helper functions for type conversion

func latPtrFromProto(lat float64) *float64 {
	if lat == 0 {
		return nil
	}
	return &lat
}

func lonPtrFromProto(lon float64) *float64 {
	if lon == 0 {
		return nil
	}
	return &lon
}

func strPtrFromProto(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func tsToPtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func statusFromEnum(s pb.AttendanceStatus) string {
	switch s {
	case pb.AttendanceStatus_ATTENDANCE_STATUS_PRESENT:
		return "PRESENT"
	case pb.AttendanceStatus_ATTENDANCE_STATUS_LATE:
		return "LATE"
	case pb.AttendanceStatus_ATTENDANCE_STATUS_ABSENT:
		return "ABSENT"
	case pb.AttendanceStatus_ATTENDANCE_STATUS_HALF_DAY:
		return "HALF_DAY"
	case pb.AttendanceStatus_ATTENDANCE_STATUS_ON_LEAVE:
		return "ON_LEAVE"
	default:
		return "ABSENT"
	}
}

func statusToEnum(s string) pb.AttendanceStatus {
	switch s {
	case "PRESENT":
		return pb.AttendanceStatus_ATTENDANCE_STATUS_PRESENT
	case "LATE":
		return pb.AttendanceStatus_ATTENDANCE_STATUS_LATE
	case "ABSENT":
		return pb.AttendanceStatus_ATTENDANCE_STATUS_ABSENT
	case "HALF_DAY":
		return pb.AttendanceStatus_ATTENDANCE_STATUS_HALF_DAY
	case "ON_LEAVE":
		return pb.AttendanceStatus_ATTENDANCE_STATUS_ON_LEAVE
	default:
		return pb.AttendanceStatus_ATTENDANCE_STATUS_ABSENT
	}
}

func formatWorkDuration(mins int) string {
	if mins <= 0 {
		return "0"
	}
	return string(rune(mins % 256))
}

func dtoToProto(dto *queries.AttendanceDTO) *pb.AttendanceRecord {
	rec := &pb.AttendanceRecord{
		Id:         dto.ID,
		TenantId:   dto.TenantID,
		EmployeeId: dto.EmployeeID,
		Date:       dto.Date,
		Status:     statusToEnum(dto.Status),
		Notes:      dto.Notes,
		IsOverride: dto.IsOverride,
		CreatedBy:  dto.CreatedBy,
	}

	if dto.CheckInAt != nil {
		t, err := time.Parse(time.RFC3339, *dto.CheckInAt)
		if err == nil {
			rec.CheckInAt = timestamppb.New(t)
		}
	}

	if dto.CheckOutAt != nil {
		t, err := time.Parse(time.RFC3339, *dto.CheckOutAt)
		if err == nil {
			rec.CheckOutAt = timestamppb.New(t)
		}
	}

	if dto.WorkDurationMins != nil {
		rec.WorkDurationMins = formatWorkDuration(*dto.WorkDurationMins)
	}

	return rec
}
