package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Briyantama/HRIS-Stery/services/employee/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
	"github.com/Briyantama/stery-kit/go/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRegisteredType is the event type that creates an employee shell record.
const UserRegisteredType = "hris.identity.user.registered"

// unassigned names the per-tenant department and position that hold shell employees until HR
// moves them during onboarding.
const unassigned = "Unassigned"

type userRegisteredPayload struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// UserRegistered creates an employee shell record for every newly registered user.
type UserRegistered struct {
	pool      *pgxpool.Pool
	depts     domain.DepartmentRepository
	positions domain.PositionRepository
	create    *commands.CreateEmployeeHandler
}

// NewUserRegistered creates the consumer; pool backs the processed_events idempotency table.
func NewUserRegistered(
	pool *pgxpool.Pool,
	depts domain.DepartmentRepository,
	positions domain.PositionRepository,
	create *commands.CreateEmployeeHandler,
) *UserRegistered {
	return &UserRegistered{pool: pool, depts: depts, positions: positions, create: create}
}

// Handle processes one hris.identity.user.registered event. It returns an error (entry stays
// pending) for transient failures and nil for events that can never succeed.
func (c *UserRegistered) Handle(ctx context.Context, env events.Envelope) error {
	eventID, err := uuid.Parse(env.EventID)
	if err != nil {
		slog.ErrorContext(ctx, "skip user registered event: invalid event id", "event_id", env.EventID)
		return nil
	}
	var p userRegisteredPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		slog.ErrorContext(ctx, "skip user registered event: invalid payload", "event_id", env.EventID)
		return nil
	}
	if p.TenantID == "" {
		p.TenantID = env.TenantID
	}
	tenantID, err := domain.NewTenantID(p.TenantID)
	if err != nil || p.Email == "" || p.FullName == "" {
		slog.ErrorContext(ctx, "skip user registered event: missing or invalid fields", "event_id", env.EventID)
		return nil
	}

	done, err := c.isProcessed(ctx, eventID)
	if err != nil {
		return err
	}
	if done {
		return nil
	}

	deptID, err := c.defaultDepartment(ctx, tenantID)
	if err != nil {
		return err
	}
	posID, err := c.defaultPosition(ctx, tenantID)
	if err != nil {
		return err
	}

	_, err = c.create.Handle(ctx, commands.CreateEmployeeCommand{
		TenantID:     tenantID.String(),
		Email:        p.Email,
		FullName:     p.FullName,
		DepartmentID: deptID,
		PositionID:   posID,
		ActorID:      "system",
	})
	// An existing employee with this email means an earlier delivery already did the work.
	if err != nil && !errors.Is(err, domain.ErrConflict) {
		return fmt.Errorf("create employee shell: %w", err)
	}
	if err == nil {
		slog.InfoContext(ctx, "created employee shell", "tenant_id", tenantID.String(), "event_id", env.EventID)
	}

	return c.markProcessed(ctx, eventID, env.EventType)
}

// defaultDepartment returns the tenant's "Unassigned" department id, creating it on first use.
// ponytail: two API replicas racing on a new tenant hit the unique (tenant_id, name) constraint; the
// loser's entry stays pending and succeeds on retry.
func (c *UserRegistered) defaultDepartment(ctx context.Context, tenantID domain.TenantID) (string, error) {
	list, err := c.depts.ListByTenant(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("list departments: %w", err)
	}
	for _, d := range list {
		if d.Name() == unassigned {
			return d.ID().String(), nil
		}
	}
	d, err := domain.NewDepartment(tenantID, unassigned)
	if err != nil {
		return "", err
	}
	if err := c.depts.Create(ctx, d); err != nil {
		return "", fmt.Errorf("create default department: %w", err)
	}
	return d.ID().String(), nil
}

// defaultPosition returns the tenant's "Unassigned" position id, creating it on first use.
func (c *UserRegistered) defaultPosition(ctx context.Context, tenantID domain.TenantID) (string, error) {
	list, err := c.positions.ListByTenant(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("list positions: %w", err)
	}
	for _, p := range list {
		if p.Title() == unassigned {
			return p.ID().String(), nil
		}
	}
	p, err := domain.NewPosition(tenantID, unassigned, domain.LevelJunior)
	if err != nil {
		return "", err
	}
	if err := c.positions.Create(ctx, p); err != nil {
		return "", fmt.Errorf("create default position: %w", err)
	}
	return p.ID().String(), nil
}

func (c *UserRegistered) isProcessed(ctx context.Context, eventID uuid.UUID) (bool, error) {
	var exists bool
	err := c.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM employee.processed_events WHERE event_id = $1)`, eventID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check processed event: %w", err)
	}
	return exists, nil
}

func (c *UserRegistered) markProcessed(ctx context.Context, eventID uuid.UUID, eventType string) error {
	_, err := c.pool.Exec(ctx, `INSERT INTO employee.processed_events (event_id, event_type) VALUES ($1, $2) ON CONFLICT DO NOTHING`, eventID, eventType)
	if err != nil {
		return fmt.Errorf("mark event processed: %w", err)
	}
	return nil
}
