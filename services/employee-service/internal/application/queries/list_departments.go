package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

// ListDepartmentsQuery is the input DTO.
type ListDepartmentsQuery struct {
	TenantID string
}

// DepartmentSummary is a brief department record.
type DepartmentSummary struct {
	ID          string
	Name        string
	Description string
	HeadID      *string
	CreatedAt   int64
}

// ListDepartmentsResult is the output DTO.
type ListDepartmentsResult struct {
	Departments []*DepartmentSummary
	Total       int32
}

// ListDepartmentsHandler implements the list departments use case.
type ListDepartmentsHandler struct {
	departmentRepo domain.DepartmentRepository
}

// NewListDepartmentsHandler creates a new list departments handler.
func NewListDepartmentsHandler(departmentRepo domain.DepartmentRepository) *ListDepartmentsHandler {
	return &ListDepartmentsHandler{departmentRepo: departmentRepo}
}

// Handle executes the list departments query.
func (h *ListDepartmentsHandler) Handle(ctx context.Context, query ListDepartmentsQuery) (*ListDepartmentsResult, error) {
	// Parse tenant ID
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Fetch departments
	departments, err := h.departmentRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}

	// Build result
	summaries := make([]*DepartmentSummary, len(departments))
	for i, dept := range departments {
		summary := &DepartmentSummary{
			ID:          dept.ID().String(),
			Name:        dept.Name(),
			Description: dept.Description(),
			CreatedAt:   dept.CreatedAt().Unix(),
		}
		if dept.HeadID() != nil {
			headIDStr := dept.HeadID().String()
			summary.HeadID = &headIDStr
		}
		summaries[i] = summary
	}

	return &ListDepartmentsResult{
		Departments: summaries,
		Total:       int32(len(summaries)),
	}, nil
}
