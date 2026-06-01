package commands

import (
	"context"
	"testing"
	"time"

	"github.com/hris-stery/hris-stery/services/attendance-service/internal/domain"
)

type mockAttendanceRepo struct {
	records map[string]*domain.AttendanceRecord
	created []*domain.AttendanceRecord
	updated []*domain.AttendanceRecord
}

func newMockAttendanceRepo() *mockAttendanceRepo {
	return &mockAttendanceRepo{
		records: make(map[string]*domain.AttendanceRecord),
		created: []*domain.AttendanceRecord{},
		updated: []*domain.AttendanceRecord{},
	}
}

func (m *mockAttendanceRepo) Create(ctx context.Context, attendance *domain.AttendanceRecord) error {
	m.created = append(m.created, attendance)
	m.records[attendance.ID().String()] = attendance
	return nil
}

func (m *mockAttendanceRepo) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.AttendanceID) (*domain.AttendanceRecord, error) {
	if record, ok := m.records[id.String()]; ok {
		return record, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockAttendanceRepo) GetByEmployeeAndDate(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, date time.Time) (*domain.AttendanceRecord, error) {
	for _, record := range m.records {
		if record.TenantID().Equals(tenantID) && record.EmployeeID().Equals(employeeID) && record.Date().Equal(date) {
			return record, nil
		}
	}
	return nil, nil
}

func (m *mockAttendanceRepo) ListByTenant(ctx context.Context, tenantID domain.TenantID, filters domain.AttendanceFilters) ([]*domain.AttendanceRecord, error) {
	var results []*domain.AttendanceRecord
	for _, record := range m.records {
		if record.TenantID().Equals(tenantID) {
			results = append(results, record)
		}
	}
	return results, nil
}

func (m *mockAttendanceRepo) Update(ctx context.Context, attendance *domain.AttendanceRecord) error {
	m.updated = append(m.updated, attendance)
	m.records[attendance.ID().String()] = attendance
	return nil
}

func (m *mockAttendanceRepo) ListByEmployee(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, from time.Time, to time.Time) ([]*domain.AttendanceRecord, error) {
	return []*domain.AttendanceRecord{}, nil
}

func (m *mockAttendanceRepo) ListActiveCheckIns(ctx context.Context, tenantID domain.TenantID) ([]*domain.AttendanceRecord, error) {
	return []*domain.AttendanceRecord{}, nil
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

func TestCheckInNewRecord(t *testing.T) {
	ctx := context.Background()
	repo := newMockAttendanceRepo()
	publisher := newMockEventPublisher()
	handler := NewCheckInHandler(repo, publisher)

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := domain.GenerateEmployeeID()
	date := time.Now()
	checkInTime := time.Now()

	cmd := CheckInCommand{
		TenantID:   tenantID.String(),
		EmployeeID: employeeID.String(),
		Date:       date,
		CheckInAt:  checkInTime,
		Notes:      "Office",
		ActorID:    "user-123",
	}

	result, err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	if result == nil {
		t.Fatalf("expected result, got nil")
	}

	if len(repo.created) != 1 {
		t.Errorf("expected 1 created record, got %d", len(repo.created))
	}

	if len(publisher.published) != 1 {
		t.Errorf("expected 1 published event, got %d", len(publisher.published))
	}
}

func TestCheckInExistingRecord(t *testing.T) {
	ctx := context.Background()
	repo := newMockAttendanceRepo()
	publisher := newMockEventPublisher()
	handler := NewCheckInHandler(repo, publisher)

	tenantID := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := domain.GenerateEmployeeID()
	date := time.Now()

	// Pre-create a record
	existingRecord, _ := domain.NewAttendanceRecord(tenantID, employeeID, date, "system")
	repo.records[existingRecord.ID().String()] = existingRecord

	checkInTime := time.Now()
	cmd := CheckInCommand{
		TenantID:   tenantID.String(),
		EmployeeID: employeeID.String(),
		Date:       date,
		CheckInAt:  checkInTime,
		ActorID:    "user-123",
	}

	result, err := handler.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	if result == nil {
		t.Fatalf("expected result, got nil")
	}

	if len(repo.created) != 0 {
		t.Errorf("expected 0 created records, got %d", len(repo.created))
	}

	if len(repo.updated) != 1 {
		t.Errorf("expected 1 updated record, got %d", len(repo.updated))
	}
}

func TestCheckInInvalidTenant(t *testing.T) {
	ctx := context.Background()
	repo := newMockAttendanceRepo()
	publisher := newMockEventPublisher()
	handler := NewCheckInHandler(repo, publisher)

	cmd := CheckInCommand{
		TenantID:   "invalid-tenant",
		EmployeeID: domain.GenerateEmployeeID().String(),
		Date:       time.Now(),
		CheckInAt:  time.Now(),
		ActorID:    "user-123",
	}

	_, err := handler.Handle(ctx, cmd)
	if err == nil {
		t.Fatalf("expected error for invalid tenant, got nil")
	}
}
