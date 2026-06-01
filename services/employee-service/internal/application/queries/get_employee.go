package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

// GetEmployeeQuery is the input DTO.
type GetEmployeeQuery struct {
	TenantID   string
	EmployeeID string
}

// GetEmployeeResult is the output DTO.
type GetEmployeeResult struct {
	ID              string
	Email           string
	FullName        string
	Phone           string
	DepartmentID    string
	PositionID      string
	ManagerID       *string
	Status          string
	ContractType    string
	JoinDate        int64 // Unix timestamp
	TerminationDate *int64
	CreatedAt       int64
	UpdatedAt       int64
}

// GetEmployeeHandler implements the get employee use case.
type GetEmployeeHandler struct {
	employeeRepo domain.EmployeeRepository
}

// NewGetEmployeeHandler creates a new get employee handler.
func NewGetEmployeeHandler(employeeRepo domain.EmployeeRepository) *GetEmployeeHandler {
	return &GetEmployeeHandler{employeeRepo: employeeRepo}
}

// Handle executes the get employee query.
func (h *GetEmployeeHandler) Handle(ctx context.Context, query GetEmployeeQuery) (*GetEmployeeResult, error) {
	// Parse tenant ID
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Parse employee ID
	employeeID, err := domain.NewEmployeeID(query.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	// Fetch employee
	employee, err := h.employeeRepo.GetByID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// Build result
	result := &GetEmployeeResult{
		ID:           employee.ID().String(),
		Email:        employee.Email(),
		FullName:     employee.FullName(),
		Phone:        employee.Phone(),
		DepartmentID: employee.DepartmentID().String(),
		PositionID:   employee.PositionID().String(),
		Status:       string(employee.Status()),
		ContractType: string(employee.ContractType()),
		JoinDate:     employee.JoinDate().Unix(),
		CreatedAt:    employee.CreatedAt().Unix(),
		UpdatedAt:    employee.UpdatedAt().Unix(),
	}

	// Optional fields
	if employee.ManagerID() != nil {
		managerIDStr := employee.ManagerID().String()
		result.ManagerID = &managerIDStr
	}
	if employee.TerminationDate() != nil {
		terminationDate := employee.TerminationDate().Unix()
		result.TerminationDate = &terminationDate
	}

	return result, nil
}
