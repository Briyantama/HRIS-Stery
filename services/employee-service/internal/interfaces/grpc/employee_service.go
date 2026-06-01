package grpc

import (
	"context"
	"time"

	employeev1 "github.com/hris-stery/hris-stery/gen/go/hris/employee/v1"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/queries"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EmployeeServiceServer implements employeev1.EmployeeServiceServer.
type EmployeeServiceServer struct {
	employeev1.UnimplementedEmployeeServiceServer
	createEmployeeHandler     *commands.CreateEmployeeHandler
	terminateEmployeeHandler  *commands.TerminateEmployeeHandler
	createDepartmentHandler   *commands.CreateDepartmentHandler
	createPositionHandler     *commands.CreatePositionHandler
	getEmployeeHandler        *queries.GetEmployeeHandler
	listEmployeesHandler      *queries.ListEmployeesHandler
	listDepartmentsHandler    *queries.ListDepartmentsHandler
	listPositionsHandler      *queries.ListPositionsHandler
}

// NewEmployeeServiceServer creates a new employee service gRPC server.
func NewEmployeeServiceServer(
	createEmployeeHandler *commands.CreateEmployeeHandler,
	terminateEmployeeHandler *commands.TerminateEmployeeHandler,
	createDepartmentHandler *commands.CreateDepartmentHandler,
	createPositionHandler *commands.CreatePositionHandler,
	getEmployeeHandler *queries.GetEmployeeHandler,
	listEmployeesHandler *queries.ListEmployeesHandler,
	listDepartmentsHandler *queries.ListDepartmentsHandler,
	listPositionsHandler *queries.ListPositionsHandler,
) *EmployeeServiceServer {
	return &EmployeeServiceServer{
		createEmployeeHandler:    createEmployeeHandler,
		terminateEmployeeHandler: terminateEmployeeHandler,
		createDepartmentHandler:  createDepartmentHandler,
		createPositionHandler:    createPositionHandler,
		getEmployeeHandler:       getEmployeeHandler,
		listEmployeesHandler:     listEmployeesHandler,
		listDepartmentsHandler:   listDepartmentsHandler,
		listPositionsHandler:     listPositionsHandler,
	}
}

// CreateEmployee creates a new employee.
func (s *EmployeeServiceServer) CreateEmployee(ctx context.Context, req *employeev1.CreateEmployeeRequest) (*employeev1.EmployeeResponse, error) {
	if req.TenantId == "" || req.Email == "" || req.FullName == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id, email, and full_name are required")
	}
	if req.DepartmentId == "" || req.PositionId == "" {
		return nil, status.Error(codes.InvalidArgument, "department_id and position_id are required")
	}

	result, err := s.createEmployeeHandler.Handle(ctx, commands.CreateEmployeeCommand{
		TenantID:     req.TenantId,
		Email:        req.Email,
		FullName:     req.FullName,
		Phone:        req.Phone,
		DepartmentID: req.DepartmentId,
		PositionID:   req.PositionId,
		ManagerID:    req.ManagerId,
		ContractType: req.ContractType.String(),
		ActorID:      "system", // TODO: Extract from context
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create employee failed: %v", err)
	}

	return &employeev1.EmployeeResponse{
		Id:           result.EmployeeID,
		Email:        result.Email,
		FullName:     result.FullName,
		DepartmentId: result.DepartmentID,
		PositionId:   result.PositionID,
	}, nil
}

// GetEmployee retrieves an employee by ID.
func (s *EmployeeServiceServer) GetEmployee(ctx context.Context, req *employeev1.GetEmployeeRequest) (*employeev1.EmployeeDetailResponse, error) {
	if req.TenantId == "" || req.EmployeeId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and employee_id are required")
	}

	result, err := s.getEmployeeHandler.Handle(ctx, queries.GetEmployeeQuery{
		TenantID:   req.TenantId,
		EmployeeID: req.EmployeeId,
	})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "employee not found: %v", err)
	}

	resp := &employeev1.EmployeeDetailResponse{
		Id:           result.ID,
		Email:        result.Email,
		FullName:     result.FullName,
		Phone:        result.Phone,
		DepartmentId: result.DepartmentID,
		PositionId:   result.PositionID,
		Status:       result.Status,
		ContractType: result.ContractType,
		JoinDate:     timestamppb.Unix(result.JoinDate, 0),
		CreatedAt:    timestamppb.Unix(result.CreatedAt, 0),
		UpdatedAt:    timestamppb.Unix(result.UpdatedAt, 0),
	}
	if result.ManagerID != nil {
		resp.ManagerId = *result.ManagerID
	}
	if result.TerminationDate != nil {
		resp.TerminationDate = timestamppb.Unix(*result.TerminationDate, 0)
	}
	return resp, nil
}

// ListEmployees lists employees with optional filters.
func (s *EmployeeServiceServer) ListEmployees(ctx context.Context, req *employeev1.ListEmployeesRequest) (*employeev1.ListEmployeesResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	result, err := s.listEmployeesHandler.Handle(ctx, queries.ListEmployeesQuery{
		TenantID:     req.TenantId,
		Status:       getStringPtr(req.Status),
		DepartmentID: getStringPtr(req.DepartmentId),
		SearchText:   req.SearchText,
		Limit:        req.PageSize,
		Offset:       req.PageOffset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list employees failed: %v", err)
	}

	employees := make([]*employeev1.EmployeeResponse, len(result.Employees))
	for i, emp := range result.Employees {
		employees[i] = &employeev1.EmployeeResponse{
			Id:           emp.ID,
			Email:        emp.Email,
			FullName:     emp.FullName,
			DepartmentId: emp.DepartmentID,
			PositionId:   emp.PositionID,
			Status:       emp.Status,
		}
	}

	return &employeev1.ListEmployeesResponse{
		Employees:  employees,
		Total:      result.Total,
		PageSize:   result.Limit,
		PageOffset: result.Offset,
	}, nil
}

// TerminateEmployee terminates an employee.
func (s *EmployeeServiceServer) TerminateEmployee(ctx context.Context, req *employeev1.TerminateEmployeeRequest) (*employeev1.TerminateEmployeeResponse, error) {
	if req.TenantId == "" || req.EmployeeId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and employee_id are required")
	}

	terminationDate := time.Now().UTC()
	if req.TerminationDate != nil {
		terminationDate = req.TerminationDate.AsTime()
	}

	result, err := s.terminateEmployeeHandler.Handle(ctx, commands.TerminateEmployeeCommand{
		TenantID:        req.TenantId,
		EmployeeID:      req.EmployeeId,
		TerminationDate: terminationDate,
		ActorID:         "system", // TODO: Extract from context
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "terminate employee failed: %v", err)
	}

	return &employeev1.TerminateEmployeeResponse{
		EmployeeId:      result.EmployeeID,
		Email:           result.Email,
		TerminationDate: timestamppb.Unix(result.TerminationDate.Unix(), 0),
	}, nil
}

// CreateDepartment creates a new department.
func (s *EmployeeServiceServer) CreateDepartment(ctx context.Context, req *employeev1.CreateDepartmentRequest) (*employeev1.DepartmentResponse, error) {
	if req.TenantId == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and name are required")
	}

	result, err := s.createDepartmentHandler.Handle(ctx, commands.CreateDepartmentCommand{
		TenantID:    req.TenantId,
		Name:        req.Name,
		Description: req.Description,
		ActorID:     "system", // TODO: Extract from context
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create department failed: %v", err)
	}

	return &employeev1.DepartmentResponse{
		Id:          result.DepartmentID,
		Name:        result.Name,
		Description: result.Description,
	}, nil
}

// ListDepartments lists all departments for a tenant.
func (s *EmployeeServiceServer) ListDepartments(ctx context.Context, req *employeev1.ListDepartmentsRequest) (*employeev1.ListDepartmentsResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	result, err := s.listDepartmentsHandler.Handle(ctx, queries.ListDepartmentsQuery{
		TenantID: req.TenantId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list departments failed: %v", err)
	}

	departments := make([]*employeev1.DepartmentResponse, len(result.Departments))
	for i, dept := range result.Departments {
		departments[i] = &employeev1.DepartmentResponse{
			Id:          dept.ID,
			Name:        dept.Name,
			Description: dept.Description,
		}
		if dept.HeadID != nil {
			departments[i].HeadId = *dept.HeadID
		}
	}

	return &employeev1.ListDepartmentsResponse{
		Departments: departments,
		Total:       result.Total,
	}, nil
}

// CreatePosition creates a new position.
func (s *EmployeeServiceServer) CreatePosition(ctx context.Context, req *employeev1.CreatePositionRequest) (*employeev1.PositionResponse, error) {
	if req.TenantId == "" || req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and title are required")
	}
	if req.Level == "" {
		return nil, status.Error(codes.InvalidArgument, "level is required")
	}

	result, err := s.createPositionHandler.Handle(ctx, commands.CreatePositionCommand{
		TenantID:    req.TenantId,
		Title:       req.Title,
		Description: req.Description,
		Level:       req.Level,
		ActorID:     "system", // TODO: Extract from context
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create position failed: %v", err)
	}

	return &employeev1.PositionResponse{
		Id:          result.PositionID,
		Title:       result.Title,
		Description: result.Description,
		Level:       result.Level,
	}, nil
}

// ListPositions lists all positions for a tenant.
func (s *EmployeeServiceServer) ListPositions(ctx context.Context, req *employeev1.ListPositionsRequest) (*employeev1.ListPositionsResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	result, err := s.listPositionsHandler.Handle(ctx, queries.ListPositionsQuery{
		TenantID: req.TenantId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list positions failed: %v", err)
	}

	positions := make([]*employeev1.PositionResponse, len(result.Positions))
	for i, pos := range result.Positions {
		positions[i] = &employeev1.PositionResponse{
			Id:          pos.ID,
			Title:       pos.Title,
			Description: pos.Description,
			Level:       pos.Level,
		}
	}

	return &employeev1.ListPositionsResponse{
		Positions: positions,
		Total:     result.Total,
	}, nil
}

// Helper function to convert empty string to nil pointer.
func getStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
