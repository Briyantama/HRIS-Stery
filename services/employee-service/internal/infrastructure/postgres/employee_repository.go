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

// EmployeeRepository implements domain.EmployeeRepository using PostgreSQL with RLS.
type EmployeeRepository struct {
	pool *pgxpool.Pool
}

// NewEmployeeRepository creates a new Postgres employee repository.
func NewEmployeeRepository(pool *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{pool: pool}
}

// Create persists a new employee.
func (r *EmployeeRepository) Create(ctx context.Context, employee *domain.Employee) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(employee.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			INSERT INTO employee.employees
			(id, tenant_id, email, phone, full_name, gender, birth_date, national_id, tax_id,
			 department_id, position_id, manager_id, status, contract_type, join_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		`
		_, err := tx.Exec(ctx, query,
			employee.ID().String(),
			employee.TenantID().String(),
			employee.Email(),
			employee.Phone(),
			employee.FullName(),
			string(employee.Gender()),
			employee.BirthDate(),
			employee.NationalID(),
			employee.TaxID(),
			employee.DepartmentID().String(),
			employee.PositionID().String(),
			managerIDOrNil(employee.ManagerID()),
			string(employee.Status()),
			string(employee.ContractType()),
			employee.JoinDate(),
			employee.CreatedAt(),
			employee.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert employee: %w", err)
		}
		return nil
	})
}

// GetByID retrieves an employee by ID within a tenant.
func (r *EmployeeRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.EmployeeID) (*domain.Employee, error) {
	var employee *domain.Employee
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, email, phone, full_name, gender, birth_date, national_id, tax_id,
			       department_id, position_id, manager_id, status, contract_type, join_date, termination_date, created_at, updated_at
			FROM employee.employees
			WHERE id = $1
		`
		var (
			empID, deptID, posID           string
			email, phone, fullName         string
			gender, status, contract       string
			managerID                      *string
			terminationDate                *time.Time
			birthDate                      *time.Time
			nationalID, taxID              string
			joinDate, createdAt, updatedAt time.Time
		)
		rowErr := tx.QueryRow(ctx, query, id.String()).Scan(
			&empID, &tenantID, &email, &phone, &fullName, &gender, &birthDate, &nationalID, &taxID,
			&deptID, &posID, &managerID, &status, &contract, &joinDate, &terminationDate, &createdAt, &updatedAt,
		)
		if rowErr == pgx.ErrNoRows {
			return fmt.Errorf("employee not found")
		}
		if rowErr != nil {
			return fmt.Errorf("query employee: %w", rowErr)
		}

		// Reconstruct employee domain object
		var mid *domain.EmployeeID
		if managerID != nil {
			m := domain.MustNewEmployeeID(*managerID)
			mid = &m
		}

		emp, err := domain.RehydrateEmployee(
			domain.MustNewEmployeeID(empID),
			domain.MustNewTenantID(tenantID.String()),
			email, phone, fullName,
			domain.Gender(gender),
			birthDate,
			nationalID, taxID,
			domain.MustNewDepartmentID(deptID),
			domain.MustNewPositionID(posID),
			mid,
			domain.EmploymentStatus(status),
			domain.ContractType(contract),
			joinDate,
			terminationDate,
			createdAt, updatedAt,
		)
		if err != nil {
			return fmt.Errorf("rehydrate employee: %w", err)
		}
		employee = emp
		return nil
	})
	if err != nil {
		return nil, err
	}
	return employee, nil
}

// GetByTenantAndEmail retrieves an employee by email within a tenant.
func (r *EmployeeRepository) GetByTenantAndEmail(ctx context.Context, tenantID domain.TenantID, email string) (*domain.Employee, error) {
	var employee *domain.Employee
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, email, phone, full_name, gender, birth_date, national_id, tax_id,
			       department_id, position_id, manager_id, status, contract_type, join_date, termination_date, created_at, updated_at
			FROM employee.employees
			WHERE tenant_id = $1 AND email = $2
		`
		var (
			empID, deptID, posID           string
			emailVal, phone, fullName      string
			gender, status, contract       string
			managerID                      *string
			terminationDate                *time.Time
			birthDate                      *time.Time
			nationalID, taxID              string
			joinDate, createdAt, updatedAt time.Time
		)
		rowErr := tx.QueryRow(ctx, query, tenantID.String(), email).Scan(
			&empID, &tenantID, &emailVal, &phone, &fullName, &gender, &birthDate, &nationalID, &taxID,
			&deptID, &posID, &managerID, &status, &contract, &joinDate, &terminationDate, &createdAt, &updatedAt,
		)
		if rowErr == pgx.ErrNoRows {
			return nil
		}
		if rowErr != nil {
			return fmt.Errorf("query employee by email: %w", rowErr)
		}

		var mid *domain.EmployeeID
		if managerID != nil {
			m := domain.MustNewEmployeeID(*managerID)
			mid = &m
		}

		emp, err := domain.RehydrateEmployee(
			domain.MustNewEmployeeID(empID),
			domain.MustNewTenantID(tenantID.String()),
			emailVal, phone, fullName,
			domain.Gender(gender),
			birthDate,
			nationalID, taxID,
			domain.MustNewDepartmentID(deptID),
			domain.MustNewPositionID(posID),
			mid,
			domain.EmploymentStatus(status),
			domain.ContractType(contract),
			joinDate,
			terminationDate,
			createdAt, updatedAt,
		)
		if err != nil {
			return fmt.Errorf("rehydrate employee: %w", err)
		}
		employee = emp
		return nil
	})
	if err != nil {
		return nil, err
	}
	return employee, nil
}

// Update persists changes to an employee.
func (r *EmployeeRepository) Update(ctx context.Context, employee *domain.Employee) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(employee.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			UPDATE employee.employees
			SET email = $1, phone = $2, full_name = $3, gender = $4, birth_date = $5, national_id = $6, tax_id = $7,
			    department_id = $8, position_id = $9, manager_id = $10, status = $11, contract_type = $12,
			    join_date = $13, termination_date = $14, updated_at = $15
			WHERE id = $16
		`
		result, err := tx.Exec(ctx, query,
			employee.Email(),
			employee.Phone(),
			employee.FullName(),
			string(employee.Gender()),
			employee.BirthDate(),
			employee.NationalID(),
			employee.TaxID(),
			employee.DepartmentID().String(),
			employee.PositionID().String(),
			managerIDOrNil(employee.ManagerID()),
			string(employee.Status()),
			string(employee.ContractType()),
			employee.JoinDate(),
			employee.TerminationDate(),
			employee.UpdatedAt(),
			employee.ID().String(),
		)
		if err != nil {
			return fmt.Errorf("update employee: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("employee not found")
		}
		return nil
	})
}

// Delete removes an employee record (soft or hard, depends on policy).
func (r *EmployeeRepository) Delete(ctx context.Context, tenantID domain.TenantID, id domain.EmployeeID) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `DELETE FROM employee.employees WHERE id = $1`
		result, err := tx.Exec(ctx, query, id.String())
		if err != nil {
			return fmt.Errorf("delete employee: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("employee not found")
		}
		return nil
	})
}

// ListByTenant retrieves all employees for a tenant with optional filters.
func (r *EmployeeRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID, filters domain.EmployeeFilters) ([]*domain.Employee, error) {
	var employees []*domain.Employee
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, email, phone, full_name, gender, birth_date, national_id, tax_id,
			       department_id, position_id, manager_id, status, contract_type, join_date, termination_date, created_at, updated_at
			FROM employee.employees
			WHERE tenant_id = $1
		`
		args := []interface{}{tenantID.String()}
		argNum := 2

		// Apply filters
		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argNum)
			args = append(args, string(*filters.Status))
			argNum++
		}
		if filters.DepartmentID != nil {
			query += fmt.Sprintf(" AND department_id = $%d", argNum)
			args = append(args, filters.DepartmentID.String())
			argNum++
		}
		if filters.SearchText != "" {
			query += fmt.Sprintf(" AND (full_name ILIKE $%d OR email ILIKE $%d)", argNum, argNum+1)
			searchPattern := "%" + filters.SearchText + "%"
			args = append(args, searchPattern, searchPattern)
			argNum += 2
		}

		// Add pagination
		query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filters.Limit, filters.Offset)

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("query employees: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				empID, deptID, posID           string
				email, phone, fullName         string
				gender, status, contract       string
				managerID                      *string
				terminationDate                *time.Time
				birthDate                      *time.Time
				nationalID, taxID              string
				joinDate, createdAt, updatedAt time.Time
			)
			if err := rows.Scan(
				&empID, &tenantID, &email, &phone, &fullName, &gender, &birthDate, &nationalID, &taxID,
				&deptID, &posID, &managerID, &status, &contract, &joinDate, &terminationDate, &createdAt, &updatedAt,
			); err != nil {
				return fmt.Errorf("scan employee: %w", err)
			}

			var mid *domain.EmployeeID
			if managerID != nil {
				m := domain.MustNewEmployeeID(*managerID)
				mid = &m
			}

			emp, err := domain.RehydrateEmployee(
				domain.MustNewEmployeeID(empID),
				domain.MustNewTenantID(tenantID.String()),
				email, phone, fullName,
				domain.Gender(gender),
				birthDate,
				nationalID, taxID,
				domain.MustNewDepartmentID(deptID),
				domain.MustNewPositionID(posID),
				mid,
				domain.EmploymentStatus(status),
				domain.ContractType(contract),
				joinDate,
				terminationDate,
				createdAt, updatedAt,
			)
			if err != nil {
				return fmt.Errorf("rehydrate employee: %w", err)
			}
			employees = append(employees, emp)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return employees, nil
}

// GetByManager retrieves direct reports of a manager.
func (r *EmployeeRepository) GetByManager(ctx context.Context, tenantID domain.TenantID, managerID domain.EmployeeID) ([]*domain.Employee, error) {
	var employees []*domain.Employee
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, email, phone, full_name, gender, birth_date, national_id, tax_id,
			       department_id, position_id, manager_id, status, contract_type, join_date, termination_date, created_at, updated_at
			FROM employee.employees
			WHERE tenant_id = $1 AND manager_id = $2
		`
		rows, err := tx.Query(ctx, query, tenantID.String(), managerID.String())
		if err != nil {
			return fmt.Errorf("query by manager: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				empID, deptID, posID           string
				email, phone, fullName         string
				gender, status, contract       string
				managerIDVal                   *string
				terminationDate                *time.Time
				birthDate                      *time.Time
				nationalID, taxID              string
				joinDate, createdAt, updatedAt time.Time
			)
			if err := rows.Scan(
				&empID, &tenantID, &email, &phone, &fullName, &gender, &birthDate, &nationalID, &taxID,
				&deptID, &posID, &managerIDVal, &status, &contract, &joinDate, &terminationDate, &createdAt, &updatedAt,
			); err != nil {
				return fmt.Errorf("scan employee: %w", err)
			}

			var mid *domain.EmployeeID
			if managerIDVal != nil {
				m := domain.MustNewEmployeeID(*managerIDVal)
				mid = &m
			}

			emp, err := domain.RehydrateEmployee(
				domain.MustNewEmployeeID(empID),
				domain.MustNewTenantID(tenantID.String()),
				email, phone, fullName,
				domain.Gender(gender),
				birthDate,
				nationalID, taxID,
				domain.MustNewDepartmentID(deptID),
				domain.MustNewPositionID(posID),
				mid,
				domain.EmploymentStatus(status),
				domain.ContractType(contract),
				joinDate,
				terminationDate,
				createdAt, updatedAt,
			)
			if err != nil {
				return fmt.Errorf("rehydrate employee: %w", err)
			}
			employees = append(employees, emp)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return employees, nil
}

// Helper function to convert *EmployeeID to *string for NULL insertion.
func managerIDOrNil(mid *domain.EmployeeID) *string {
	if mid == nil {
		return nil
	}
	s := mid.String()
	return &s
}
