package commands

import (
	"context"
	"testing"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
)

// Mock implementations for testing
type transition struct {
	request *domain.LeaveRequest
	from    domain.LeaveStatus
	delta   domain.BalanceDelta
}

type mockLeaveRequestRepo struct {
	requests      map[string]*domain.LeaveRequest
	created       []*domain.LeaveRequest
	createDeltas  []domain.BalanceDelta
	transitions   []transition
	createErr     error
	transitionErr error
}

func newMockLeaveRequestRepo() *mockLeaveRequestRepo {
	return &mockLeaveRequestRepo{
		requests: make(map[string]*domain.LeaveRequest),
		created:  []*domain.LeaveRequest{},
	}
}

func (m *mockLeaveRequestRepo) Create(ctx context.Context, req *domain.LeaveRequest) error {
	m.created = append(m.created, req)
	m.requests[req.ID().String()] = req
	return nil
}

func (m *mockLeaveRequestRepo) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.LeaveRequestID) (*domain.LeaveRequest, error) {
	if req, ok := m.requests[id.String()]; ok {
		return req, nil
	}
	return nil, nil
}

func (m *mockLeaveRequestRepo) ListByTenant(ctx context.Context, tenantID domain.TenantID, filters domain.LeaveFilters) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepo) CountByTenant(ctx context.Context, tenantID domain.TenantID, filters domain.LeaveFilters) (int, error) {
	return 0, nil
}

func (m *mockLeaveRequestRepo) ListByEmployee(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, year int) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepo) CreateAndUpdateBalanceAtomically(ctx context.Context, request *domain.LeaveRequest, delta domain.BalanceDelta) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.createDeltas = append(m.createDeltas, delta)
	return m.Create(ctx, request)
}

func (m *mockLeaveRequestRepo) TransitionAndUpdateBalanceAtomically(ctx context.Context, request *domain.LeaveRequest, from domain.LeaveStatus, delta domain.BalanceDelta) error {
	if m.transitionErr != nil {
		return m.transitionErr
	}
	m.transitions = append(m.transitions, transition{request: request, from: from, delta: delta})
	return nil
}

type mockLeaveBalanceRepo struct {
	balances map[string]*domain.LeaveBalance
}

func newMockLeaveBalanceRepo() *mockLeaveBalanceRepo {
	return &mockLeaveBalanceRepo{
		balances: make(map[string]*domain.LeaveBalance),
	}
}

func (m *mockLeaveBalanceRepo) Create(ctx context.Context, balance *domain.LeaveBalance) error {
	m.balances[balance.ID().String()] = balance
	return nil
}

func (m *mockLeaveBalanceRepo) GetByEmployeeAndType(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, leaveTypeID domain.LeaveTypeID, year int) (*domain.LeaveBalance, error) {
	for _, b := range m.balances {
		if b.EmployeeID().Equals(employeeID) && b.LeaveTypeID().Equals(leaveTypeID) && b.Year() == year {
			return b, nil
		}
	}
	return nil, nil
}

func (m *mockLeaveBalanceRepo) ListByEmployee(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, year int) ([]*domain.LeaveBalance, error) {
	return nil, nil
}

type mockEventPublisher struct {
	published []domain.DomainEvent
}

func newMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{
		published: []domain.DomainEvent{},
	}
}

func (m *mockEventPublisher) Publish(ctx context.Context, event domain.DomainEvent, tenantID domain.TenantID, actorID string) error {
	m.published = append(m.published, event)
	return nil
}

type mockEmployeeValidator struct {
	validEmployees map[string]bool
}

func newMockEmployeeValidator() *mockEmployeeValidator {
	return &mockEmployeeValidator{
		validEmployees: make(map[string]bool),
	}
}

func (m *mockEmployeeValidator) ValidateEmployeeExists(ctx context.Context, tenantID, employeeID string) (bool, error) {
	key := tenantID + ":" + employeeID
	return m.validEmployees[key], nil
}

// Tests

func TestCreateLeaveRequestSuccess(t *testing.T) {
	ctx := context.Background()
	leaveRequestRepo := newMockLeaveRequestRepo()
	leaveBalanceRepo := newMockLeaveBalanceRepo()
	eventPublisher := newMockEventPublisher()
	employeeValidator := newMockEmployeeValidator()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := domain.GenerateEmployeeID()
	leaveTypeID := domain.GenerateLeaveTypeID()

	// Setup: create a balance for the employee
	balance, _ := domain.NewLeaveBalance(tenantID, employeeID, leaveTypeID, 2026, 20.0)
	leaveBalanceRepo.balances[balance.ID().String()] = balance

	// Setup: mark employee as valid
	key := tenantID.String() + ":" + employeeID.String()
	employeeValidator.validEmployees[key] = true

	handler := NewCreateLeaveRequestHandler(leaveRequestRepo, leaveBalanceRepo, eventPublisher, employeeValidator)

	cmd := CreateLeaveRequestCommand{
		TenantID:    tenantID.String(),
		EmployeeID:  employeeID.String(),
		LeaveTypeID: leaveTypeID.String(),
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, 5),
		DaysCount:   5,
		Reason:      "Vacation",
		ActorID:     employeeID.String(),
	}

	result, err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("CreateLeaveRequest failed: %v", err)
	}

	if result == nil {
		t.Fatalf("expected result, got nil")
	}

	if len(leaveRequestRepo.created) != 1 {
		t.Errorf("expected 1 created request, got %d", len(leaveRequestRepo.created))
	}

	if len(eventPublisher.published) != 1 {
		t.Errorf("expected 1 published event, got %d", len(eventPublisher.published))
	}

	// The balance change is a relative delta, not an absolute write
	if len(leaveRequestRepo.createDeltas) != 1 || leaveRequestRepo.createDeltas[0] != (domain.BalanceDelta{Pending: 5}) {
		t.Errorf("create deltas = %+v, want one {Pending: 5}", leaveRequestRepo.createDeltas)
	}
}

func TestCreateLeaveRequestInvalidEmployee(t *testing.T) {
	ctx := context.Background()
	leaveRequestRepo := newMockLeaveRequestRepo()
	leaveBalanceRepo := newMockLeaveBalanceRepo()
	eventPublisher := newMockEventPublisher()
	employeeValidator := newMockEmployeeValidator()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := domain.GenerateEmployeeID()
	leaveTypeID := domain.GenerateLeaveTypeID()

	// Employee is NOT marked as valid
	handler := NewCreateLeaveRequestHandler(leaveRequestRepo, leaveBalanceRepo, eventPublisher, employeeValidator)

	cmd := CreateLeaveRequestCommand{
		TenantID:    tenantID.String(),
		EmployeeID:  employeeID.String(),
		LeaveTypeID: leaveTypeID.String(),
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, 5),
		DaysCount:   5,
		Reason:      "Vacation",
		ActorID:     employeeID.String(),
	}

	_, err := handler.Handle(ctx, cmd)
	if err == nil {
		t.Fatalf("expected error for invalid employee, got nil")
	}
}

func TestCreateLeaveRequestInsufficientBalance(t *testing.T) {
	ctx := context.Background()
	leaveRequestRepo := newMockLeaveRequestRepo()
	leaveBalanceRepo := newMockLeaveBalanceRepo()
	eventPublisher := newMockEventPublisher()
	employeeValidator := newMockEmployeeValidator()

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := domain.GenerateEmployeeID()
	leaveTypeID := domain.GenerateLeaveTypeID()

	// Create a balance with only 3 days
	balance, _ := domain.NewLeaveBalance(tenantID, employeeID, leaveTypeID, 2026, 3.0)
	leaveBalanceRepo.balances[balance.ID().String()] = balance

	// Mark employee as valid
	key := tenantID.String() + ":" + employeeID.String()
	employeeValidator.validEmployees[key] = true

	handler := NewCreateLeaveRequestHandler(leaveRequestRepo, leaveBalanceRepo, eventPublisher, employeeValidator)

	cmd := CreateLeaveRequestCommand{
		TenantID:    tenantID.String(),
		EmployeeID:  employeeID.String(),
		LeaveTypeID: leaveTypeID.String(),
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, 5),
		DaysCount:   5, // Requesting more than available
		Reason:      "Vacation",
		ActorID:     employeeID.String(),
	}

	_, err := handler.Handle(ctx, cmd)
	if err == nil {
		t.Fatalf("expected error for insufficient balance, got nil")
	}
}
