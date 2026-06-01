package postgres

import (
	"context"
	"fmt"
	"time"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LeaveBalanceRepository implements domain.LeaveBalanceRepository using PostgreSQL with RLS.
type LeaveBalanceRepository struct {
	pool *pgxpool.Pool
}

// NewLeaveBalanceRepository creates a new Postgres leave balance repository.
func NewLeaveBalanceRepository(pool *pgxpool.Pool) *LeaveBalanceRepository {
	return &LeaveBalanceRepository{pool: pool}
}

// Create persists a new leave balance with RLS enforcement.
func (r *LeaveBalanceRepository) Create(ctx context.Context, balance *domain.LeaveBalance) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(balance.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			INSERT INTO leave.leave_balances
			(id, tenant_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		_, err := tx.Exec(ctx, query,
			balance.ID().String(),
			balance.TenantID().String(),
			balance.EmployeeID().String(),
			balance.LeaveTypeID().String(),
			balance.Year(),
			balance.EntitledDays(),
			balance.UsedDays(),
			balance.PendingDays(),
			balance.CreatedAt(),
			balance.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert leave balance: %w", err)
		}
		return nil
	})
}

// GetByEmployeeAndType retrieves a balance for a specific employee and leave type.
func (r *LeaveBalanceRepository) GetByEmployeeAndType(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, leaveTypeID domain.LeaveTypeID, year int) (*domain.LeaveBalance, error) {
	var balance *domain.LeaveBalance
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days, created_at, updated_at
			FROM leave.leave_balances
			WHERE employee_id = $1 AND leave_type_id = $2 AND year = $3
		`
		var (
			balanceID, empID, typeID string
			queryYear                int
			entitledDays, usedDays, pendingDays float64
			createdAt, updatedAt     time.Time
		)
		rowErr := tx.QueryRow(ctx, query, employeeID.String(), leaveTypeID.String(), year).Scan(
			&balanceID, &tenantID, &empID, &typeID, &queryYear, &entitledDays, &usedDays, &pendingDays, &createdAt, &updatedAt,
		)
		if rowErr == pgx.ErrNoRows {
			return nil
		}
		if rowErr != nil {
			return fmt.Errorf("query leave balance: %w", rowErr)
		}

		balance = domain.RehydrateLeaveBalance(
			domain.MustNewLeaveBalanceID(balanceID),
			domain.MustNewTenantID(tenantID.String()),
			domain.MustNewEmployeeID(empID),
			domain.MustNewLeaveTypeID(typeID),
			queryYear,
			entitledDays,
			usedDays,
			pendingDays,
			createdAt,
			updatedAt,
		)
		return nil
	})
	return balance, err
}

// ListByEmployee retrieves all balances for a specific employee in a year.
func (r *LeaveBalanceRepository) ListByEmployee(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, year int) ([]*domain.LeaveBalance, error) {
	var balances []*domain.LeaveBalance
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days, created_at, updated_at
			FROM leave.leave_balances
			WHERE employee_id = $1 AND year = $2
			ORDER BY leave_type_id ASC
		`
		rows, err := tx.Query(ctx, query, employeeID.String(), year)
		if err != nil {
			return fmt.Errorf("query leave balances: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				balanceID, empID, typeID string
				queryYear                int
				entitledDays, usedDays, pendingDays float64
				createdAt, updatedAt     time.Time
			)
			if err := rows.Scan(
				&balanceID, &tenantID, &empID, &typeID, &queryYear, &entitledDays, &usedDays, &pendingDays, &createdAt, &updatedAt,
			); err != nil {
				return fmt.Errorf("scan row: %w", err)
			}

			balance := domain.RehydrateLeaveBalance(
				domain.MustNewLeaveBalanceID(balanceID),
				domain.MustNewTenantID(tenantID.String()),
				domain.MustNewEmployeeID(empID),
				domain.MustNewLeaveTypeID(typeID),
				queryYear,
				entitledDays,
				usedDays,
				pendingDays,
				createdAt,
				updatedAt,
			)
			balances = append(balances, balance)
		}

		return rows.Err()
	})
	return balances, err
}

// Update persists changes to a leave balance.
func (r *LeaveBalanceRepository) Update(ctx context.Context, balance *domain.LeaveBalance) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(balance.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			UPDATE leave.leave_balances
			SET used_days = $1, pending_days = $2, updated_at = $3
			WHERE id = $4
		`
		result, err := tx.Exec(ctx, query,
			balance.UsedDays(),
			balance.PendingDays(),
			balance.UpdatedAt(),
			balance.ID().String(),
		)
		if err != nil {
			return fmt.Errorf("update leave balance: %w", err)
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf("leave balance not found")
		}

		return nil
	})
}
