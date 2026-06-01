package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

var errNotFound = errors.New("not found")

type mockEmployeeRepository struct {
	employees map[string]*domain.Employee
	storage   map[string][]*domain.Employee
}

func newMockEmployeeRepository() *mockEmployeeRepository {
	return &mockEmployeeRepository{
		employees: make(map[string]*domain.Employee),
		storage:   make(map[string][]*domain.Employee),
	}
}

func (m *mockEmployeeRepository) Create(ctx context.Context, employee *domain.Employee) error {
	m.employees[employee.ID().String()] = employee
	return nil
}

func (m *mockEmployeeRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.EmployeeID) (*domain.Employee, error) {
	if emp, ok := m.employees[id.String()]; ok {
		return emp, nil
	}
	return nil, errNotFound
}

func (m *mockEmployeeRepository) GetByTenantAndEmail(ctx context.Context, tenantID domain.TenantID, email string) (*domain.Employee, error) {
	for _, emp := range m.employees {
		if emp.Email() == email && emp.TenantID().Equals(tenantID) {
			return emp, nil
		}
	}
	return nil, nil
}

func (m *mockEmployeeRepository) Update(ctx context.Context, employee *domain.Employee) error {
	m.employees[employee.ID().String()] = employee
	return nil
}

func (m *mockEmployeeRepository) Delete(ctx context.Context, tenantID domain.TenantID, id domain.EmployeeID) error {
	delete(m.employees, id.String())
	return nil
}

func (m *mockEmployeeRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID, filters domain.EmployeeFilters) ([]*domain.Employee, error) {
	return []*domain.Employee{}, nil
}

func (m *mockEmployeeRepository) GetByManager(ctx context.Context, tenantID domain.TenantID, managerID domain.EmployeeID) ([]*domain.Employee, error) {
	return []*domain.Employee{}, nil
}

type mockDepartmentRepository struct {
	departments map[string]*domain.Department
}

func newMockDepartmentRepository() *mockDepartmentRepository {
	return &mockDepartmentRepository{
		departments: make(map[string]*domain.Department),
	}
}

func (m *mockDepartmentRepository) Create(ctx context.Context, department *domain.Department) error {
	m.departments[department.ID().String()] = department
	return nil
}

func (m *mockDepartmentRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.DepartmentID) (*domain.Department, error) {
	if dept, ok := m.departments[id.String()]; ok {
		return dept, nil
	}
	return nil, errNotFound
}

func (m *mockDepartmentRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.Department, error) {
	return []*domain.Department{}, nil
}

func (m *mockDepartmentRepository) Update(ctx context.Context, department *domain.Department) error {
	m.departments[department.ID().String()] = department
	return nil
}

type mockPositionRepository struct {
	positions map[string]*domain.Position
}

func newMockPositionRepository() *mockPositionRepository {
	return &mockPositionRepository{
		positions: make(map[string]*domain.Position),
	}
}

func (m *mockPositionRepository) Create(ctx context.Context, position *domain.Position) error {
	m.positions[position.ID().String()] = position
	return nil
}

func (m *mockPositionRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.PositionID) (*domain.Position, error) {
	if pos, ok := m.positions[id.String()]; ok {
		return pos, nil
	}
	return nil, errNotFound
}

func (m *mockPositionRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.Position, error) {
	return []*domain.Position{}, nil
}

func (m *mockPositionRepository) Update(ctx context.Context, position *domain.Position) error {
	m.positions[position.ID().String()] = position
	return nil
}

type mockEventPublisher struct {
	events []domain.DomainEvent
}

func newMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{events: []domain.DomainEvent{}}
}

func (m *mockEventPublisher) PublishAsync(ctx context.Context, event domain.DomainEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventPublisher) PublishSync(ctx context.Context, event domain.DomainEvent) error {
	m.events = append(m.events, event)
	return nil
}

func TestCreateEmployeeSuccess(t *testing.T) {
	ctx := context.Background()
	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	deptID := domain.GenerateDepartmentID()
	posID := domain.GeneratePositionID()

	empRepo := newMockEmployeeRepository()
	deptRepo := newMockDepartmentRepository()
	posRepo := newMockPositionRepository()
	eventPub := newMockEventPublisher()

	// Create test department and position
	dept, _ := domain.NewDepartment(tenantID, "Engineering")
	_ = deptRepo.Create(ctx, dept)
	deptID = dept.ID()

	pos, _ := domain.NewPosition(tenantID, "Software Engineer", domain.LevelSenior)
	_ = posRepo.Create(ctx, pos)
	posID = pos.ID()

	handler := NewCreateEmployeeHandler(empRepo, deptRepo, posRepo, eventPub)
	cmd := CreateEmployeeCommand{
		TenantID:     tenantID.String(),
		Email:        "john@example.com",
		FullName:     "John Doe",
		DepartmentID: deptID.String(),
		PositionID:   posID.String(),
		ActorID:      "test-user",
	}

	result, err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("CreateEmployee failed: %v", err)
	}

	if result.Email != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", result.Email)
	}
	if result.FullName != "John Doe" {
		t.Errorf("expected full name John Doe, got %s", result.FullName)
	}
	if len(eventPub.events) != 1 {
		t.Errorf("expected 1 event published, got %d", len(eventPub.events))
	}
}

func TestCreateEmployeeDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	deptID := domain.GenerateDepartmentID()
	posID := domain.GeneratePositionID()

	empRepo := newMockEmployeeRepository()
	deptRepo := newMockDepartmentRepository()
	posRepo := newMockPositionRepository()
	eventPub := newMockEventPublisher()

	// Pre-populate with existing employee
	dept, _ := domain.NewDepartment(tenantID, "Engineering")
	_ = deptRepo.Create(ctx, dept)
	deptID = dept.ID()

	pos, _ := domain.NewPosition(tenantID, "Software Engineer", domain.LevelSenior)
	_ = posRepo.Create(ctx, pos)
	posID = pos.ID()

	existing, _ := domain.NewEmployee(tenantID, "john@example.com", "John Doe", deptID, posID)
	_ = empRepo.Create(ctx, existing)

	handler := NewCreateEmployeeHandler(empRepo, deptRepo, posRepo, eventPub)
	cmd := CreateEmployeeCommand{
		TenantID:     tenantID.String(),
		Email:        "john@example.com", // Duplicate
		FullName:     "Jane Doe",
		DepartmentID: deptID.String(),
		PositionID:   posID.String(),
		ActorID:      "test-user",
	}

	_, err := handler.Handle(ctx, cmd)
	if err == nil {
		t.Errorf("expected error for duplicate email, got nil")
	}
}
