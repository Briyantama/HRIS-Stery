package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
	"github.com/Briyantama/stery-kit/go/pgtenant"
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
	return pgtenant.WithTenantTx(ctx, r.pool, balance.TenantID().String(), "", func(tx pgx.Tx) error {
		query := `
			INSERT INTO leave.leave_balances
			(id, tenant_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, employee_id, leave_type_id, year) DO NOTHING
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
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days, created_at, updated_at
			FROM leave.leave_balances
			WHERE employee_id = $1 AND leave_type_id = $2 AND year = $3
		`
		var (
			balanceID, empID, typeID            string
			queryYear                           int
			entitledDays, usedDays, pendingDays float64
			createdAt, updatedAt                time.Time
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
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
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
				balanceID, empID, typeID            string
				queryYear                           int
				entitledDays, usedDays, pendingDays float64
				createdAt, updatedAt                time.Time
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

// applyBalanceDelta adds the delta to the request's balance row in SQL. The WHERE clause keeps used and
// pending non-negative and, when the delta grows the committed total, within the entitlement, so a
// concurrent change can neither be overwritten nor push the balance over its limit; no row matched
// means a guard failed.
func applyBalanceDelta(ctx context.Context, tx pgx.Tx, request *domain.LeaveRequest, delta domain.BalanceDelta) error {
	query := `
		UPDATE leave.leave_balances
		SET used_days = used_days + $1, pending_days = pending_days + $2, updated_at = now()
		WHERE employee_id = $3 AND leave_type_id = $4 AND year = $5
		  AND used_days + $1 >= 0 AND pending_days + $2 >= 0
		  AND ($1 + $2 <= 0 OR used_days + pending_days + $1 + $2 <= entitled_days)
	`
	result, err := tx.Exec(ctx, query,
		delta.Used,
		delta.Pending,
		request.EmployeeID().String(),
		request.LeaveTypeID().String(),
		request.StartDate().Year(),
	)
	if err != nil {
		return fmt.Errorf("update leave balance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("leave balance for %d: %w", request.StartDate().Year(), domain.ErrInsufficientLeaveBalance)
	}

	return nil
}

// CreateWithTx persists a new leave balance within an existing transaction.
// Supports atomic operations wrapping multiple repositories.
func (r *LeaveBalanceRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, balance *domain.LeaveBalance) error {
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
}

// GetByEmployeeAndTypeWithTx retrieves a balance for a specific employee and leave type within an existing transaction.
// Supports atomic operations wrapping multiple repositories.
func (r *LeaveBalanceRepository) GetByEmployeeAndTypeWithTx(ctx context.Context, tx pgx.Tx, tenantID domain.TenantID, employeeID domain.EmployeeID, leaveTypeID domain.LeaveTypeID, year int) (*domain.LeaveBalance, error) {
	query := `
		SELECT id, tenant_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days, created_at, updated_at
		FROM leave.leave_balances
		WHERE employee_id = $1 AND leave_type_id = $2 AND year = $3
	`
	var (
		balanceID, empID, typeID            string
		queryYear                           int
		entitledDays, usedDays, pendingDays float64
		createdAt, updatedAt                time.Time
	)
	rowErr := tx.QueryRow(ctx, query, employeeID.String(), leaveTypeID.String(), year).Scan(
		&balanceID, &tenantID, &empID, &typeID, &queryYear, &entitledDays, &usedDays, &pendingDays, &createdAt, &updatedAt,
	)
	if rowErr == pgx.ErrNoRows {
		return nil, nil
	}
	if rowErr != nil {
		return nil, fmt.Errorf("query leave balance: %w", rowErr)
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
	return balance, nil
}
