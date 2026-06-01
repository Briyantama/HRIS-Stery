package grpc

import (
	"context"
	"time"

	employeev1 "github.com/hris-stery/hris-stery/gen/go/hris/employee/v1"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/queries"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EmployeeServiceServer implements employeev1.EmployeeServiceServer.
type EmployeeServiceServer struct {
	employeev1.UnimplementedEmployeeServiceServer
	createEmployeeHandler    *commands.CreateEmployeeHandler
	terminateEmployeeHandler *commands.TerminateEmployeeHandler
	createDepartmentHandler  *commands.CreateDepartmentHandler
	createPositionHandler    *commands.CreatePositionHandler
	getEmployeeHandler       *queries.GetEmployeeHandler
	listEmployeesHandler     *queries.ListEmployeesHandler
	listDepartmentsHandler   *queries.ListDepartmentsHandler
	listPositionsHandler     *queries.ListPositionsHandler
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
func (s *EmployeeServiceServer) CreateEmployee(ctx context.Context, req *employeev1.CreateEmployeeRequest) (*employeev1.CreateEmployeeResponse, error) {
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

	emp := &employeev1.Employee{
		Id:           result.EmployeeID,
		TenantId:     req.TenantId,
		FullName:     result.FullName,
		Email:        result.Email,
		Phone:        req.Phone,
		DepartmentId: result.DepartmentID,
		PositionId:   result.PositionID,
	}

	return &employeev1.CreateEmployeeResponse{
		Employee: emp,
	}, nil
}

// GetEmployee retrieves an employee by ID.
func (s *EmployeeServiceServer) GetEmployee(ctx context.Context, req *employeev1.GetEmployeeRequest) (*employeev1.GetEmployeeResponse, error) {
	if req.TenantId == "" || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and id are required")
	}

	result, err := s.getEmployeeHandler.Handle(ctx, queries.GetEmployeeQuery{
		TenantID:   req.TenantId,
		EmployeeID: req.Id,
	})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "employee not found: %v", err)
	}

	emp := &employeev1.Employee{
		Id:           result.ID,
		TenantId:     req.TenantId,
		FullName:     result.FullName,
		Email:        result.Email,
		Phone:        result.Phone,
		DepartmentId: result.DepartmentID,
		PositionId:   result.PositionID,
	}

	return &employeev1.GetEmployeeResponse{
		Employee: emp,
	}, nil
}

// ListEmployees lists employees with optional filters.
func (s *EmployeeServiceServer) ListEmployees(ctx context.Context, req *employeev1.ListEmployeesRequest) (*employeev1.ListEmployeesResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	var deptID *string
	if req.DepartmentId != "" {
		deptID = &req.DepartmentId
	}
	result, err := s.listEmployeesHandler.Handle(ctx, queries.ListEmployeesQuery{
		TenantID:     req.TenantId,
		DepartmentID: deptID,
		SearchText:   req.Search,
		Limit:        req.PageSize,
		Offset:       0, // TODO: Implement pagination with page_token
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list employees failed: %v", err)
	}

	employees := make([]*employeev1.Employee, len(result.Employees))
	for i, emp := range result.Employees {
		employees[i] = &employeev1.Employee{
			Id:           emp.ID,
			TenantId:     req.TenantId,
			Email:        emp.Email,
			FullName:     emp.FullName,
			DepartmentId: emp.DepartmentID,
			PositionId:   emp.PositionID,
		}
	}

	return &employeev1.ListEmployeesResponse{
		Employees:     employees,
		TotalCount:    int32(result.Total),
		NextPageToken: "", // TODO: Implement pagination tokens
	}, nil
}

// TerminateEmployee terminates an employee.
func (s *EmployeeServiceServer) TerminateEmployee(ctx context.Context, req *employeev1.TerminateEmployeeRequest) (*employeev1.TerminateEmployeeResponse, error) {
	if req.TenantId == "" || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and id are required")
	}

	terminationDate := time.Now().UTC()
	if req.TerminationDate != "" {
		t, err := time.Parse(time.RFC3339, req.TerminationDate)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid termination_date format")
		}
		terminationDate = t
	}

	result, err := s.terminateEmployeeHandler.Handle(ctx, commands.TerminateEmployeeCommand{
		TenantID:        req.TenantId,
		EmployeeID:      req.Id,
		TerminationDate: terminationDate,
		ActorID:         "system", // TODO: Extract from context
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "terminate employee failed: %v", err)
	}

	emp := &employeev1.Employee{
		Id:              result.EmployeeID,
		TenantId:        req.TenantId,
		Email:           result.Email,
		TerminationDate: terminationDate.Format(time.RFC3339),
	}

	return &employeev1.TerminateEmployeeResponse{
		Employee: emp,
	}, nil
}

// CreateDepartment creates a new department.
func (s *EmployeeServiceServer) CreateDepartment(ctx context.Context, req *employeev1.CreateDepartmentRequest) (*employeev1.CreateDepartmentResponse, error) {
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

	dept := &employeev1.Department{
		Id:          result.DepartmentID,
		TenantId:    req.TenantId,
		Name:        result.Name,
		Description: result.Description,
	}

	return &employeev1.CreateDepartmentResponse{
		Department: dept,
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

	departments := make([]*employeev1.Department, len(result.Departments))
	for i, dept := range result.Departments {
		departments[i] = &employeev1.Department{
			Id:          dept.ID,
			TenantId:    req.TenantId,
			Name:        dept.Name,
			Description: dept.Description,
		}
		if dept.HeadID != nil {
			departments[i].HeadId = *dept.HeadID
		}
	}

	return &employeev1.ListDepartmentsResponse{
		Departments: departments,
	}, nil
}

// CreatePosition creates a new position.
func (s *EmployeeServiceServer) CreatePosition(ctx context.Context, req *employeev1.CreatePositionRequest) (*employeev1.CreatePositionResponse, error) {
	if req.TenantId == "" || req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and title are required")
	}
	if req.Level == "" {
		return nil, status.Error(codes.InvalidArgument, "level is required")
	}

	result, err := s.createPositionHandler.Handle(ctx, commands.CreatePositionCommand{
		TenantID: req.TenantId,
		Title:    req.Title,
		Level:    req.Level,
		ActorID:  "system", // TODO: Extract from context
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create position failed: %v", err)
	}

	pos := &employeev1.Position{
		Id:       result.PositionID,
		TenantId: req.TenantId,
		Title:    result.Title,
		Level:    result.Level,
	}

	return &employeev1.CreatePositionResponse{
		Position: pos,
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

	positions := make([]*employeev1.Position, len(result.Positions))
	for i, pos := range result.Positions {
		positions[i] = &employeev1.Position{
			Id:       pos.ID,
			TenantId: req.TenantId,
			Title:    pos.Title,
			Level:    pos.Level,
		}
	}

	return &employeev1.ListPositionsResponse{
		Positions: positions,
	}, nil
}
