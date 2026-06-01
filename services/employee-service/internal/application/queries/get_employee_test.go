package queries

import (
	"context"
	"errors"
	"testing"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

var errNotFound = errors.New("not found")

type mockEmployeeRepositoryForQuery struct {
	employees map[string]*domain.Employee
}

func newMockEmployeeRepositoryForQuery() *mockEmployeeRepositoryForQuery {
	return &mockEmployeeRepositoryForQuery{
		employees: make(map[string]*domain.Employee),
	}
}

func (m *mockEmployeeRepositoryForQuery) Create(ctx context.Context, employee *domain.Employee) error {
	m.employees[employee.ID().String()] = employee
	return nil
}

func (m *mockEmployeeRepositoryForQuery) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.EmployeeID) (*domain.Employee, error) {
	if emp, ok := m.employees[id.String()]; ok {
		return emp, nil
	}
	return nil, errNotFound
}

func (m *mockEmployeeRepositoryForQuery) GetByTenantAndEmail(ctx context.Context, tenantID domain.TenantID, email string) (*domain.Employee, error) {
	for _, emp := range m.employees {
		if emp.Email() == email && emp.TenantID().Equals(tenantID) {
			return emp, nil
		}
	}
	return nil, nil
}

func (m *mockEmployeeRepositoryForQuery) Update(ctx context.Context, employee *domain.Employee) error {
	m.employees[employee.ID().String()] = employee
	return nil
}

func (m *mockEmployeeRepositoryForQuery) Delete(ctx context.Context, tenantID domain.TenantID, id domain.EmployeeID) error {
	delete(m.employees, id.String())
	return nil
}

func (m *mockEmployeeRepositoryForQuery) ListByTenant(ctx context.Context, tenantID domain.TenantID, filters domain.EmployeeFilters) ([]*domain.Employee, error) {
	var results []*domain.Employee
	for _, emp := range m.employees {
		if emp.TenantID().Equals(tenantID) {
			results = append(results, emp)
		}
	}
	return results, nil
}

func (m *mockEmployeeRepositoryForQuery) GetByManager(ctx context.Context, tenantID domain.TenantID, managerID domain.EmployeeID) ([]*domain.Employee, error) {
	return []*domain.Employee{}, nil
}

func TestGetEmployeeSuccess(t *testing.T) {
	ctx := context.Background()
	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	deptID := domain.GenerateDepartmentID()
	posID := domain.GeneratePositionID()

	repo := newMockEmployeeRepositoryForQuery()

	// Create test employee
	emp, err := domain.NewEmployee(tenantID, "john@example.com", "John Doe", deptID, posID)
	if err != nil {
		t.Fatalf("NewEmployee failed: %v", err)
	}
	_ = repo.Create(ctx, emp)

	handler := NewGetEmployeeHandler(repo)
	query := GetEmployeeQuery{
		TenantID:   tenantID.String(),
		EmployeeID: emp.ID().String(),
	}

	result, err := handler.Handle(ctx, query)
	if err != nil {
		t.Fatalf("GetEmployee failed: %v", err)
	}

	if result.Email != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", result.Email)
	}
	if result.FullName != "John Doe" {
		t.Errorf("expected full name John Doe, got %s", result.FullName)
	}
}

func TestGetEmployeeNotFound(t *testing.T) {
	ctx := context.Background()
	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	repo := newMockEmployeeRepositoryForQuery()

	handler := NewGetEmployeeHandler(repo)
	query := GetEmployeeQuery{
		TenantID:   tenantID.String(),
		EmployeeID: domain.GenerateEmployeeID().String(),
	}

	_, err := handler.Handle(ctx, query)
	if err == nil {
		t.Errorf("expected error for non-existent employee, got nil")
	}
}

func TestListEmployeesSuccess(t *testing.T) {
	ctx := context.Background()
	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	deptID := domain.GenerateDepartmentID()
	posID := domain.GeneratePositionID()

	repo := newMockEmployeeRepositoryForQuery()

	// Create test employees
	emp1, _ := domain.NewEmployee(tenantID, "john@example.com", "John Doe", deptID, posID)
	emp2, _ := domain.NewEmployee(tenantID, "jane@example.com", "Jane Smith", deptID, posID)
	_ = repo.Create(ctx, emp1)
	_ = repo.Create(ctx, emp2)

	handler := NewListEmployeesHandler(repo)
	query := ListEmployeesQuery{
		TenantID:   tenantID.String(),
		Status:     nil,
		DepartmentID: nil,
		SearchText: "",
		Offset:     0,
		Limit:      10,
	}

	result, err := handler.Handle(ctx, query)
	if err != nil {
		t.Fatalf("ListEmployees failed: %v", err)
	}

	if len(result.Employees) != 2 {
		t.Errorf("expected 2 employees, got %d", len(result.Employees))
	}
}
