package postgres

import (
	"context"
	"fmt"
	"time"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DepartmentRepository implements domain.DepartmentRepository using PostgreSQL with RLS.
type DepartmentRepository struct {
	pool *pgxpool.Pool
}

// NewDepartmentRepository creates a new Postgres department repository.
func NewDepartmentRepository(pool *pgxpool.Pool) *DepartmentRepository {
	return &DepartmentRepository{pool: pool}
}

// Create persists a new department.
func (r *DepartmentRepository) Create(ctx context.Context, department *domain.Department) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(department.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			INSERT INTO employee.departments (id, tenant_id, name, description, head_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, now(), now())
		`
		_, err := tx.Exec(ctx, query,
			department.ID().String(),
			department.TenantID().String(),
			department.Name(),
			department.Description(),
			managerIDOrNil(department.HeadID()),
		)
		if err != nil {
			return fmt.Errorf("insert department: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a department by ID within a tenant.
func (r *DepartmentRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.DepartmentID) (*domain.Department, error) {
	var department *domain.Department
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, name, description, head_id, created_at, updated_at
			FROM employee.departments
			WHERE id = $1
		`
		var (
			deptID               string
			name, desc           string
			headID               *string
			createdAt, updatedAt time.Time
		)
		rowErr := tx.QueryRow(ctx, query, id.String()).Scan(&deptID, &tenantID, &name, &desc, &headID, &createdAt, &updatedAt)
		if rowErr == pgx.ErrNoRows {
			return fmt.Errorf("department not found")
		}
		if rowErr != nil {
			return fmt.Errorf("query department: %w", rowErr)
		}

		var hid *domain.EmployeeID
		if headID != nil {
			h := domain.MustNewEmployeeID(*headID)
			hid = &h
		}

		dept, err := domain.RehydrateDepartment(
			domain.MustNewDepartmentID(deptID),
			domain.MustNewTenantID(tenantID.String()),
			name, desc, hid, createdAt, updatedAt,
		)
		if err != nil {
			return fmt.Errorf("rehydrate department: %w", err)
		}
		department = dept
		return nil
	})
	if err != nil {
		return nil, err
	}
	return department, nil
}

// ListByTenant retrieves all departments for a tenant.
func (r *DepartmentRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.Department, error) {
	var departments []*domain.Department
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, name, description, head_id, created_at, updated_at
			FROM employee.departments
			WHERE tenant_id = $1
			ORDER BY name
		`
		rows, err := tx.Query(ctx, query, tenantID.String())
		if err != nil {
			return fmt.Errorf("query departments: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				deptID               string
				name, desc           string
				headID               *string
				createdAt, updatedAt time.Time
			)
			if err := rows.Scan(&deptID, &tenantID, &name, &desc, &headID, &createdAt, &updatedAt); err != nil {
				return fmt.Errorf("scan department: %w", err)
			}

			var hid *domain.EmployeeID
			if headID != nil {
				h := domain.MustNewEmployeeID(*headID)
				hid = &h
			}

			dept, err := domain.RehydrateDepartment(
				domain.MustNewDepartmentID(deptID),
				domain.MustNewTenantID(tenantID.String()),
				name, desc, hid, createdAt, updatedAt,
			)
			if err != nil {
				return fmt.Errorf("rehydrate department: %w", err)
			}
			departments = append(departments, dept)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return departments, nil
}

// Update persists changes to a department.
func (r *DepartmentRepository) Update(ctx context.Context, department *domain.Department) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(department.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			UPDATE employee.departments
			SET name = $1, description = $2, head_id = $3, updated_at = now()
			WHERE id = $4
		`
		result, err := tx.Exec(ctx, query,
			department.Name(),
			department.Description(),
			managerIDOrNil(department.HeadID()),
			department.ID().String(),
		)
		if err != nil {
			return fmt.Errorf("update department: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("department not found")
		}
		return nil
	})
}
