package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

// ListEmployeesQuery is the input DTO.
type ListEmployeesQuery struct {
	TenantID     string
	Status       *string // optional filter
	DepartmentID *string // optional filter
	SearchText   string
	Limit        int32
	Offset       int32
}

// EmployeeSummary is a brief employee record.
type EmployeeSummary struct {
	ID           string
	Email        string
	FullName     string
	DepartmentID string
	PositionID   string
	Status       string
	CreatedAt    int64
}

// ListEmployeesResult is the output DTO.
type ListEmployeesResult struct {
	Employees []*EmployeeSummary
	Total     int32
	Limit     int32
	Offset    int32
}

// ListEmployeesHandler implements the list employees use case.
type ListEmployeesHandler struct {
	employeeRepo domain.EmployeeRepository
}

// NewListEmployeesHandler creates a new list employees handler.
func NewListEmployeesHandler(employeeRepo domain.EmployeeRepository) *ListEmployeesHandler {
	return &ListEmployeesHandler{employeeRepo: employeeRepo}
}

// Handle executes the list employees query.
func (h *ListEmployeesHandler) Handle(ctx context.Context, query ListEmployeesQuery) (*ListEmployeesResult, error) {
	// Parse tenant ID
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Build filters
	filters := domain.EmployeeFilters{
		SearchText: query.SearchText,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}

	// Optional status filter
	if query.Status != nil {
		status := domain.EmploymentStatus(*query.Status)
		if !status.IsValid() {
			return nil, fmt.Errorf("invalid status: %s", *query.Status)
		}
		filters.Status = &status
	}

	// Optional department filter
	if query.DepartmentID != nil {
		deptID, err := domain.NewDepartmentID(*query.DepartmentID)
		if err != nil {
			return nil, fmt.Errorf("invalid department_id: %w", err)
		}
		filters.DepartmentID = &deptID
	}

	// Set defaults
	if filters.Limit == 0 {
		filters.Limit = 50
	}
	if filters.Limit > 1000 {
		filters.Limit = 1000
	}

	// Fetch employees
	employees, err := h.employeeRepo.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}

	// Build result
	summaries := make([]*EmployeeSummary, len(employees))
	for i, emp := range employees {
		summaries[i] = &EmployeeSummary{
			ID:           emp.ID().String(),
			Email:        emp.Email(),
			FullName:     emp.FullName(),
			DepartmentID: emp.DepartmentID().String(),
			PositionID:   emp.PositionID().String(),
			Status:       string(emp.Status()),
			CreatedAt:    emp.CreatedAt().Unix(),
		}
	}

	return &ListEmployeesResult{
		Employees: summaries,
		Total:     int32(len(employees)),
		Limit:     filters.Limit,
		Offset:    filters.Offset,
	}, nil
}
