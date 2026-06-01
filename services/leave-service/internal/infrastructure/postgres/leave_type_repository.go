package postgres

import (
	"context"
	"fmt"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LeaveTypeRepository implements domain.LeaveTypeRepository using PostgreSQL with RLS.
type LeaveTypeRepository struct {
	pool *pgxpool.Pool
}

// NewLeaveTypeRepository creates a new Postgres leave type repository.
func NewLeaveTypeRepository(pool *pgxpool.Pool) *LeaveTypeRepository {
	return &LeaveTypeRepository{pool: pool}
}

// Create persists a new leave type with RLS enforcement.
func (r *LeaveTypeRepository) Create(ctx context.Context, leaveType *domain.LeaveType) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(leaveType.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			INSERT INTO leave.leave_types
			(id, tenant_id, code, name, max_days_per_year, requires_document, is_paid, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err := tx.Exec(ctx, query,
			leaveType.ID().String(),
			leaveType.TenantID().String(),
			leaveType.Code(),
			leaveType.Name(),
			leaveType.MaxDaysPerYear(),
			leaveType.RequiresDocument(),
			leaveType.IsPaid(),
			leaveType.IsActive(),
		)
		if err != nil {
			return fmt.Errorf("insert leave type: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a leave type by ID within a tenant.
func (r *LeaveTypeRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.LeaveTypeID) (*domain.LeaveType, error) {
	var leaveType *domain.LeaveType
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, code, name, max_days_per_year, requires_document, is_paid, is_active
			FROM leave.leave_types
			WHERE id = $1
		`
		var (
			typeID, code, name string
			maxDaysPerYear     float64
			requiresDocument   bool
			isPaid             bool
			isActive           bool
		)
		rowErr := tx.QueryRow(ctx, query, id.String()).Scan(
			&typeID, &tenantID, &code, &name, &maxDaysPerYear, &requiresDocument, &isPaid, &isActive,
		)
		if rowErr == pgx.ErrNoRows {
			return nil
		}
		if rowErr != nil {
			return fmt.Errorf("query leave type: %w", rowErr)
		}

		leaveType = domain.RehydrateLeaveType(
			domain.MustNewLeaveTypeID(typeID),
			domain.MustNewTenantID(tenantID.String()),
			code,
			name,
			maxDaysPerYear,
			requiresDocument,
			isPaid,
			isActive,
		)
		return nil
	})
	return leaveType, err
}

// ListByTenant retrieves all leave types for a tenant.
func (r *LeaveTypeRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.LeaveType, error) {
	var types []*domain.LeaveType
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, code, name, max_days_per_year, requires_document, is_paid, is_active
			FROM leave.leave_types
			WHERE is_active = true
			ORDER BY name ASC
		`
		rows, err := tx.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("query leave types: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				typeID, code, name string
				maxDaysPerYear     float64
				requiresDocument   bool
				isPaid             bool
				isActive           bool
			)
			if err := rows.Scan(&typeID, &tenantID, &code, &name, &maxDaysPerYear, &requiresDocument, &isPaid, &isActive); err != nil {
				return fmt.Errorf("scan row: %w", err)
			}

			lt := domain.RehydrateLeaveType(
				domain.MustNewLeaveTypeID(typeID),
				domain.MustNewTenantID(tenantID.String()),
				code,
				name,
				maxDaysPerYear,
				requiresDocument,
				isPaid,
				isActive,
			)
			types = append(types, lt)
		}

		return rows.Err()
	})
	return types, err
}
