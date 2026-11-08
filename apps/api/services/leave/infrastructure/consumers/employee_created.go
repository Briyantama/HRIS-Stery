package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
	"github.com/Briyantama/stery-kit/go/events"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EmployeeCreatedType is the event type that creates the initial leave balances of a new employee.
const EmployeeCreatedType = "hris.workforce.employee.created"

// EmployeeCreated initializes leave balances for every leave type of the tenant when an employee is created.
type EmployeeCreated struct {
	pool          *pgxpool.Pool
	balanceRepo   domain.LeaveBalanceRepository
	leaveTypeRepo domain.LeaveTypeRepository
}

// NewEmployeeCreated creates the consumer handler.
func NewEmployeeCreated(pool *pgxpool.Pool, balanceRepo domain.LeaveBalanceRepository, leaveTypeRepo domain.LeaveTypeRepository) *EmployeeCreated {
	return &EmployeeCreated{pool: pool, balanceRepo: balanceRepo, leaveTypeRepo: leaveTypeRepo}
}

type employeeCreatedPayload struct {
	EmployeeID string `json:"employee_id"`
	TenantID   string `json:"tenant_id"`
}

// Handle is an events.Handler. It returns an error (entry stays pending) when a database step fails,
// and acknowledges events whose payload can never be processed, after logging them.
func (c *EmployeeCreated) Handle(ctx context.Context, env events.Envelope) error {
	done, err := c.isProcessed(ctx, env.EventID)
	if err != nil {
		return fmt.Errorf("checking event idempotency: %w", err)
	}
	if done {
		return nil
	}

	var payload employeeCreatedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		slog.ErrorContext(ctx, "dropping employee created event with unreadable payload", "event_id", env.EventID, "tenant_id", env.TenantID)
		return nil
	}
	tenantID, terr := domain.NewTenantID(payload.TenantID)
	employeeID, eerr := domain.NewEmployeeID(payload.EmployeeID)
	if terr != nil || eerr != nil {
		slog.ErrorContext(ctx, "dropping employee created event with invalid ids", "event_id", env.EventID, "tenant_id", env.TenantID)
		return nil
	}

	leaveTypes, err := c.leaveTypeRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("listing leave types: %w", err)
	}
	year := time.Now().Year()
	for _, lt := range leaveTypes {
		balance, err := domain.NewLeaveBalance(tenantID, employeeID, lt.ID(), year, lt.MaxDaysPerYear())
		if err != nil {
			return fmt.Errorf("creating leave balance: %w", err)
		}
		// Create is idempotent (ON CONFLICT DO NOTHING), so a retry after a partial run is safe.
		if err := c.balanceRepo.Create(ctx, balance); err != nil {
			return fmt.Errorf("persisting leave balance: %w", err)
		}
	}

	if _, err := c.pool.Exec(ctx, `INSERT INTO leave.processed_events (event_id, processed_at) VALUES ($1, NOW()) ON CONFLICT DO NOTHING`, env.EventID); err != nil {
		return fmt.Errorf("marking event as processed: %w", err)
	}
	return nil
}

func (c *EmployeeCreated) isProcessed(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := c.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM leave.processed_events WHERE event_id = $1)`, eventID).Scan(&exists)
	return exists, err
}
