package postgres

import (
	"context"
	"fmt"

	"github.com/Briyantama/stery-kit/go/pgtenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EmployeeValidator treats an employee as known once leave balances exist for them; the
// employee.created consumer creates those balances.
type EmployeeValidator struct {
	pool *pgxpool.Pool
}

// NewEmployeeValidator creates an EmployeeValidator.
func NewEmployeeValidator(pool *pgxpool.Pool) *EmployeeValidator {
	return &EmployeeValidator{pool: pool}
}

// ValidateEmployeeExists reports whether the employee has leave balances in the tenant.
func (v *EmployeeValidator) ValidateEmployeeExists(ctx context.Context, tenantID, employeeID string) (bool, error) {
	var exists bool
	err := pgtenant.WithTenantTx(ctx, v.pool, tenantID, "", func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM leave.leave_balances WHERE employee_id = $1)`, employeeID).Scan(&exists)
	})
	if err != nil {
		return false, fmt.Errorf("checking employee balances: %w", err)
	}
	return exists, nil
}
