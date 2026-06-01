//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/infrastructure"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Test suite uses real Postgres + NATS from Docker Compose
// Requires: DATABASE_URL, NATS_URL environment variables

var (
	pool       *pgxpool.Pool
	natsConn   *nats.Conn
	logger     *zap.Logger
	tenantID   domain.TenantID
	tenantID2  domain.TenantID
	deptID     domain.DepartmentID
	posID      domain.PositionID
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error

	// Initialize logger
	logger, err = zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Connect to Postgres
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable"
	}
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logger.Fatal("parse database URL", zap.Error(err))
	}
	pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Fatal("connect to database", zap.Error(err))
	}
	defer func() {
		_ = pool.Close()
	}()

	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	natsConn, err = nats.Connect(natsURL)
	if err != nil {
		logger.Fatal("connect to NATS", zap.Error(err))
	}
	defer func() {
		_ = natsConn.Close()
	}()

	// Create test tenant IDs
	tenantID = domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440001")
	tenantID2 = domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440002")
	deptID = domain.GenerateDepartmentID()
	posID = domain.GeneratePositionID()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// TestCreateEmployeeEndToEnd verifies full create flow: command → repo → DB
func TestCreateEmployeeEndToEnd(t *testing.T) {
	ctx := context.Background()

	// Create repositories
	empRepo := postgres.NewEmployeeRepository(pool)
	deptRepo := postgres.NewDepartmentRepository(pool)
	posRepo := postgres.NewPositionRepository(pool)

	// Create event publisher
	eventPub, err := infrastructure.NewNATSPublisher(natsConn)
	if err != nil {
		t.Fatalf("NewNATSPublisher failed: %v", err)
	}

	// Create department
	dept, err := domain.NewDepartment(tenantID, "Engineering")
	if err != nil {
		t.Fatalf("NewDepartment failed: %v", err)
	}
	if err := deptRepo.Create(ctx, dept); err != nil {
		t.Fatalf("Create department failed: %v", err)
	}

	// Create position
	pos, err := domain.NewPosition(tenantID, "Software Engineer", domain.LevelSenior)
	if err != nil {
		t.Fatalf("NewPosition failed: %v", err)
	}
	if err := posRepo.Create(ctx, pos); err != nil {
		t.Fatalf("Create position failed: %v", err)
	}

	// Create employee via command
	handler := commands.NewCreateEmployeeHandler(empRepo, deptRepo, posRepo, eventPub)
	result, err := handler.Handle(ctx, commands.CreateEmployeeCommand{
		TenantID:     tenantID.String(),
		Email:        "alice@example.com",
		FullName:     "Alice Engineer",
		DepartmentID: dept.ID().String(),
		PositionID:   pos.ID().String(),
		ActorID:      "test-user",
	})
	if err != nil {
		t.Fatalf("CreateEmployee command failed: %v", err)
	}

	// Verify employee persisted in database
	emp, err := empRepo.GetByID(ctx, tenantID, domain.MustNewEmployeeID(result.EmployeeID))
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if emp.Email() != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", emp.Email())
	}
	if emp.FullName() != "Alice Engineer" {
		t.Errorf("expected full name Alice Engineer, got %s", emp.FullName())
	}
}

// TestGetEmployeeSuccess verifies retrieval of single employee
func TestGetEmployeeSuccess(t *testing.T) {
	ctx := context.Background()

	// Create and persist employee
	empRepo := postgres.NewEmployeeRepository(pool)
	deptRepo := postgres.NewDepartmentRepository(pool)
	posRepo := postgres.NewPositionRepository(pool)
	eventPub, _ := infrastructure.NewNATSPublisher(natsConn)

	dept, _ := domain.NewDepartment(tenantID, "Engineering")
	_ = deptRepo.Create(ctx, dept)
	pos, _ := domain.NewPosition(tenantID, "Senior Engineer", domain.LevelSenior)
	_ = posRepo.Create(ctx, pos)

	handler := commands.NewCreateEmployeeHandler(empRepo, deptRepo, posRepo, eventPub)
	result, _ := handler.Handle(ctx, commands.CreateEmployeeCommand{
		TenantID:     tenantID.String(),
		Email:        "bob@example.com",
		FullName:     "Bob Senior",
		DepartmentID: dept.ID().String(),
		PositionID:   pos.ID().String(),
		ActorID:      "test-user",
	})

	// Get employee via query
	queryHandler := queries.NewGetEmployeeHandler(empRepo)
	empQuery, err := queryHandler.Handle(ctx, queries.GetEmployeeQuery{
		TenantID:   tenantID.String(),
		EmployeeID: result.EmployeeID,
	})
	if err != nil {
		t.Fatalf("GetEmployee query failed: %v", err)
	}

	if empQuery.Email != "bob@example.com" {
		t.Errorf("expected email bob@example.com, got %s", empQuery.Email)
	}
}

// TestListEmployeesTenantScoping verifies RLS isolation: Tenant A cannot see Tenant B's employees
func TestListEmployeesTenantScoping(t *testing.T) {
	ctx := context.Background()

	empRepo := postgres.NewEmployeeRepository(pool)
	deptRepo := postgres.NewDepartmentRepository(pool)
	posRepo := postgres.NewPositionRepository(pool)
	eventPub, _ := infrastructure.NewNATSPublisher(natsConn)

	// Create employees for Tenant 1
	dept1, _ := domain.NewDepartment(tenantID, "Dept1")
	_ = deptRepo.Create(ctx, dept1)
	pos1, _ := domain.NewPosition(tenantID, "Pos1", domain.LevelJunior)
	_ = posRepo.Create(ctx, pos1)

	handler := commands.NewCreateEmployeeHandler(empRepo, deptRepo, posRepo, eventPub)
	_, _ = handler.Handle(ctx, commands.CreateEmployeeCommand{
		TenantID:     tenantID.String(),
		Email:        "emp1@tenant1.com",
		FullName:     "Employee 1",
		DepartmentID: dept1.ID().String(),
		PositionID:   pos1.ID().String(),
		ActorID:      "test-user",
	})

	// Create employees for Tenant 2
	dept2, _ := domain.NewDepartment(tenantID2, "Dept2")
	_ = deptRepo.Create(ctx, dept2)
	pos2, _ := domain.NewPosition(tenantID2, "Pos2", domain.LevelJunior)
	_ = posRepo.Create(ctx, pos2)

	_, _ = handler.Handle(ctx, commands.CreateEmployeeCommand{
		TenantID:     tenantID2.String(),
		Email:        "emp2@tenant2.com",
		FullName:     "Employee 2",
		DepartmentID: dept2.ID().String(),
		PositionID:   pos2.ID().String(),
		ActorID:      "test-user",
	})

	// List employees for Tenant 1
	listHandler := queries.NewListEmployeesHandler(empRepo)
	result1, err := listHandler.Handle(ctx, queries.ListEmployeesQuery{
		TenantID: tenantID.String(),
		Limit:    100,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("List for tenant 1 failed: %v", err)
	}

	// List employees for Tenant 2
	result2, err := listHandler.Handle(ctx, queries.ListEmployeesQuery{
		TenantID: tenantID2.String(),
		Limit:    100,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("List for tenant 2 failed: %v", err)
	}

	// Verify isolation
	if len(result1.Employees) == 0 {
		t.Errorf("tenant 1 should have at least 1 employee, got %d", len(result1.Employees))
	}
	if len(result2.Employees) == 0 {
		t.Errorf("tenant 2 should have at least 1 employee, got %d", len(result2.Employees))
	}

	// Verify tenant 1 employees are not in tenant 2's list
	for _, emp2 := range result2.Employees {
		if emp2.Email == "emp1@tenant1.com" {
			t.Errorf("tenant 1 employee found in tenant 2 list — RLS failed")
		}
	}
}

// TestUserRegisteredEventCreatesEmployeeShell verifies event consumer creates exactly one employee per event
func TestUserRegisteredEventCreatesEmployeeShell(t *testing.T) {
	ctx := context.Background()

	empRepo := postgres.NewEmployeeRepository(pool)
	deptRepo := postgres.NewDepartmentRepository(pool)
	posRepo := postgres.NewPositionRepository(pool)
	eventPub, _ := infrastructure.NewNATSPublisher(natsConn)

	// Create consumer
	handler := commands.NewCreateEmployeeHandler(empRepo, deptRepo, posRepo, eventPub)
	consumer := infrastructure.NewUserRegisteredConsumer(pool, handler, logger)

	// Subscribe to events
	if err := consumer.Subscribe(natsConn); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer consumer.Close()

	// Publish hris.identity.user.registered event
	eventID := fmt.Sprintf("event-%d", time.Now().UnixNano())
	envelope := map[string]interface{}{
		"event_id":       eventID,
		"event_type":     "hris.identity.user.registered",
		"schema_version": 1,
		"tenant_id":      tenantID.String(),
		"actor_id":       "auth-service",
		"occurred_at":    time.Now().UTC().Format(time.RFC3339),
		"payload": map[string]interface{}{
			"user_id":   "user-123",
			"tenant_id": tenantID.String(),
			"email":     "newuser@example.com",
			"full_name": "New User",
		},
	}

	data, _ := json.Marshal(envelope)
	if err := natsConn.Publish("hris.identity.user.registered", data); err != nil {
		t.Fatalf("Publish event failed: %v", err)
	}
	natsConn.Flush()

	// Wait for consumer to process
	time.Sleep(2 * time.Second)

	// Verify employee was created
	emp, err := empRepo.GetByTenantAndEmail(ctx, tenantID, "newuser@example.com")
	if err != nil {
		t.Fatalf("GetByTenantAndEmail failed: %v", err)
	}
	if emp == nil {
		t.Fatalf("expected employee to be created from event, got nil")
	}
	if emp.Email() != "newuser@example.com" {
		t.Errorf("expected email newuser@example.com, got %s", emp.Email())
	}
}

// TestIdempotencyDuplicateEventIgnored verifies duplicate events don't create duplicate employees
func TestIdempotencyDuplicateEventIgnored(t *testing.T) {
	ctx := context.Background()

	empRepo := postgres.NewEmployeeRepository(pool)
	deptRepo := postgres.NewDepartmentRepository(pool)
	posRepo := postgres.NewPositionRepository(pool)
	eventPub, _ := infrastructure.NewNATSPublisher(natsConn)

	// Create consumer
	handler := commands.NewCreateEmployeeHandler(empRepo, deptRepo, posRepo, eventPub)
	consumer := infrastructure.NewUserRegisteredConsumer(pool, handler, logger)

	// Subscribe
	if err := consumer.Subscribe(natsConn); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	defer consumer.Close()

	// Publish same event twice with same event_id
	eventID := fmt.Sprintf("idempotent-event-%d", time.Now().UnixNano())
	envelope := map[string]interface{}{
		"event_id":       eventID,
		"event_type":     "hris.identity.user.registered",
		"schema_version": 1,
		"tenant_id":      tenantID.String(),
		"actor_id":       "auth-service",
		"occurred_at":    time.Now().UTC().Format(time.RFC3339),
		"payload": map[string]interface{}{
			"user_id":   "user-456",
			"tenant_id": tenantID.String(),
			"email":     "dupuser@example.com",
			"full_name": "Dup User",
		},
	}

	data, _ := json.Marshal(envelope)

	// Publish twice
	_ = natsConn.Publish("hris.identity.user.registered", data)
	_ = natsConn.Publish("hris.identity.user.registered", data)
	natsConn.Flush()

	// Wait for consumer
	time.Sleep(2 * time.Second)

	// Count employees with this email
	emps, err := empRepo.ListByTenant(ctx, tenantID, domain.EmployeeFilters{})
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}

	dupCount := 0
	for _, emp := range emps {
		if emp.Email() == "dupuser@example.com" {
			dupCount++
		}
	}

	if dupCount != 1 {
		t.Errorf("expected 1 employee with email dupuser@example.com, got %d — idempotency failed", dupCount)
	}
}

// TestCrossTenantAccessDeniedByRLS verifies direct SQL queries respect RLS
func TestCrossTenantAccessDeniedByRLS(t *testing.T) {
	ctx := context.Background()

	empRepo := postgres.NewEmployeeRepository(pool)
	deptRepo := postgres.NewDepartmentRepository(pool)
	posRepo := postgres.NewPositionRepository(pool)
	eventPub, _ := infrastructure.NewNATSPublisher(natsConn)

	// Create employee in Tenant 1
	dept1, _ := domain.NewDepartment(tenantID, "RLS-Test-Dept")
	_ = deptRepo.Create(ctx, dept1)
	pos1, _ := domain.NewPosition(tenantID, "RLS-Test-Pos", domain.LevelJunior)
	_ = posRepo.Create(ctx, pos1)

	handler := commands.NewCreateEmployeeHandler(empRepo, deptRepo, posRepo, eventPub)
	result, _ := handler.Handle(ctx, commands.CreateEmployeeCommand{
		TenantID:     tenantID.String(),
		Email:        "rls-test@tenant1.com",
		FullName:     "RLS Test User",
		DepartmentID: dept1.ID().String(),
		PositionID:   pos1.ID().String(),
		ActorID:      "test-user",
	})

	empID := domain.MustNewEmployeeID(result.EmployeeID)

	// Try to get Tenant 1's employee as Tenant 2 (should fail via RLS)
	_, err := empRepo.GetByID(ctx, tenantID2, empID)
	if err == nil {
		t.Errorf("cross-tenant GetByID should fail due to RLS, but got no error — RLS not enforced")
	}
}

// TestRLSContextSetBeforeQuery verifies WithTenantTx properly sets session config
func TestRLSContextSetBeforeQuery(t *testing.T) {
	ctx := context.Background()

	deptRepo := postgres.NewDepartmentRepository(pool)

	// Create department with RLS context
	dept, err := domain.NewDepartment(tenantID, "RLS-Context-Test")
	if err != nil {
		t.Fatalf("NewDepartment failed: %v", err)
	}

	if err := deptRepo.Create(ctx, dept); err != nil {
		t.Fatalf("Create department with RLS context failed: %v", err)
	}

	// Verify the department was created and is retrievable only for its tenant
	retrieved, err := deptRepo.GetByID(ctx, tenantID, dept.ID())
	if err != nil {
		t.Fatalf("GetByID with correct tenant failed: %v", err)
	}
	if retrieved == nil {
		t.Fatalf("expected department to be retrieved, got nil")
	}

	// Try to retrieve with wrong tenant (should return nil due to RLS)
	wrongTenant, _ := deptRepo.GetByID(ctx, tenantID2, dept.ID())
	if wrongTenant != nil {
		t.Errorf("cross-tenant GetByID should return nil due to RLS, got %v", wrongTenant)
	}
}
