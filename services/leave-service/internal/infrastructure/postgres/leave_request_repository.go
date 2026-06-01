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

// LeaveRequestRepository implements domain.LeaveRequestRepository using PostgreSQL with RLS.
type LeaveRequestRepository struct {
	pool *pgxpool.Pool
}

// NewLeaveRequestRepository creates a new Postgres leave request repository.
func NewLeaveRequestRepository(pool *pgxpool.Pool) *LeaveRequestRepository {
	return &LeaveRequestRepository{pool: pool}
}

// Create persists a new leave request with RLS enforcement.
func (r *LeaveRequestRepository) Create(ctx context.Context, request *domain.LeaveRequest) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(request.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			INSERT INTO leave.leave_requests
			(id, tenant_id, employee_id, leave_type_id, start_date, end_date, days_count, status, reason, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err := tx.Exec(ctx, query,
			request.ID().String(),
			request.TenantID().String(),
			request.EmployeeID().String(),
			request.LeaveTypeID().String(),
			request.StartDate(),
			request.EndDate(),
			request.DaysCount(),
			string(request.Status()),
			request.Reason(),
			request.CreatedAt(),
			request.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert leave request: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a leave request by ID within a tenant.
func (r *LeaveRequestRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.LeaveRequestID) (*domain.LeaveRequest, error) {
	var request *domain.LeaveRequest
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, leave_type_id, start_date, end_date, days_count, status, reason,
			       rejection_reason, approved_by_id, approved_at, created_at, updated_at
			FROM leave.leave_requests
			WHERE id = $1
		`
		var (
			requestID, empID, leaveTypeID string
			startDate, endDate            time.Time
			daysCount                     int
			status, reason, rejectionReason string
			approvedByID                  *string
			approvedAt                    *time.Time
			createdAt, updatedAt          time.Time
		)
		rowErr := tx.QueryRow(ctx, query, id.String()).Scan(
			&requestID, &tenantID, &empID, &leaveTypeID, &startDate, &endDate, &daysCount,
			&status, &reason, &rejectionReason, &approvedByID, &approvedAt, &createdAt, &updatedAt,
		)
		if rowErr == pgx.ErrNoRows {
			return nil
		}
		if rowErr != nil {
			return fmt.Errorf("query leave request: %w", rowErr)
		}

		var approvedByEmpID *domain.EmployeeID
		if approvedByID != nil {
			a := domain.MustNewEmployeeID(*approvedByID)
			approvedByEmpID = &a
		}

		request = domain.RehydrateLeaveRequest(
			domain.MustNewLeaveRequestID(requestID),
			domain.MustNewTenantID(tenantID.String()),
			domain.MustNewEmployeeID(empID),
			domain.MustNewLeaveTypeID(leaveTypeID),
			startDate,
			endDate,
			daysCount,
			domain.LeaveStatus(status),
			reason,
			rejectionReason,
			approvedByEmpID,
			approvedAt,
			createdAt,
			updatedAt,
		)
		return nil
	})
	return request, err
}

// ListByTenant retrieves all leave requests for a tenant with optional filters.
func (r *LeaveRequestRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID, filters domain.LeaveFilters) ([]*domain.LeaveRequest, error) {
	var records []*domain.LeaveRequest
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, leave_type_id, start_date, end_date, days_count, status, reason,
			       rejection_reason, approved_by_id, approved_at, created_at, updated_at
			FROM leave.leave_requests
			WHERE 1=1
		`
		args := []interface{}{}
		argCount := 1

		if filters.EmployeeID != nil {
			query += fmt.Sprintf(" AND employee_id = $%d", argCount)
			args = append(args, filters.EmployeeID.String())
			argCount++
		}

		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argCount)
			args = append(args, string(*filters.Status))
			argCount++
		}

		if filters.ApprovedByID != nil {
			query += fmt.Sprintf(" AND approved_by_id = $%d", argCount)
			args = append(args, filters.ApprovedByID.String())
			argCount++
		}

		if filters.DateFrom != nil {
			query += fmt.Sprintf(" AND start_date >= $%d", argCount)
			args = append(args, *filters.DateFrom)
			argCount++
		}

		if filters.DateTo != nil {
			query += fmt.Sprintf(" AND end_date <= $%d", argCount)
			args = append(args, *filters.DateTo)
			argCount++
		}

		query += " ORDER BY start_date DESC, created_at DESC"
		if filters.Limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
			args = append(args, filters.Limit, filters.Offset)
		}

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("query leave requests: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				requestID, empID, leaveTypeID string
				startDate, endDate            time.Time
				daysCount                     int
				status, reason, rejectionReason string
				approvedByID                  *string
				approvedAt                    *time.Time
				createdAt, updatedAt          time.Time
			)
			if err := rows.Scan(
				&requestID, &tenantID, &empID, &leaveTypeID, &startDate, &endDate, &daysCount,
				&status, &reason, &rejectionReason, &approvedByID, &approvedAt, &createdAt, &updatedAt,
			); err != nil {
				return fmt.Errorf("scan row: %w", err)
			}

			var approvedByEmpID *domain.EmployeeID
			if approvedByID != nil {
				a := domain.MustNewEmployeeID(*approvedByID)
				approvedByEmpID = &a
			}

			record := domain.RehydrateLeaveRequest(
				domain.MustNewLeaveRequestID(requestID),
				domain.MustNewTenantID(tenantID.String()),
				domain.MustNewEmployeeID(empID),
				domain.MustNewLeaveTypeID(leaveTypeID),
				startDate,
				endDate,
				daysCount,
				domain.LeaveStatus(status),
				reason,
				rejectionReason,
				approvedByEmpID,
				approvedAt,
				createdAt,
				updatedAt,
			)
			records = append(records, record)
		}

		return rows.Err()
	})
	return records, err
}

// Update persists changes to a leave request.
func (r *LeaveRequestRepository) Update(ctx context.Context, request *domain.LeaveRequest) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(request.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			UPDATE leave.leave_requests
			SET status = $1, rejection_reason = $2, approved_by_id = $3, approved_at = $4, updated_at = $5
			WHERE id = $6
		`
		var approvedByID *string
		if request.ApprovedByID() != nil {
			id := request.ApprovedByID().String()
			approvedByID = &id
		}

		result, err := tx.Exec(ctx, query,
			string(request.Status()),
			request.RejectionReason(),
			approvedByID,
			request.ApprovedAt(),
			request.UpdatedAt(),
			request.ID().String(),
		)
		if err != nil {
			return fmt.Errorf("update leave request: %w", err)
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf("leave request not found")
		}

		return nil
	})
}

// ListByEmployee retrieves all leave requests for a specific employee in a year.
func (r *LeaveRequestRepository) ListByEmployee(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, year int) ([]*domain.LeaveRequest, error) {
	var records []*domain.LeaveRequest
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, leave_type_id, start_date, end_date, days_count, status, reason,
			       rejection_reason, approved_by_id, approved_at, created_at, updated_at
			FROM leave.leave_requests
			WHERE employee_id = $1 AND EXTRACT(YEAR FROM start_date) = $2
			ORDER BY start_date DESC
		`
		rows, err := tx.Query(ctx, query, employeeID.String(), year)
		if err != nil {
			return fmt.Errorf("query leave requests by employee: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				requestID, empID, leaveTypeID string
				startDate, endDate            time.Time
				daysCount                     int
				status, reason, rejectionReason string
				approvedByID                  *string
				approvedAt                    *time.Time
				createdAt, updatedAt          time.Time
			)
			if err := rows.Scan(
				&requestID, &tenantID, &empID, &leaveTypeID, &startDate, &endDate, &daysCount,
				&status, &reason, &rejectionReason, &approvedByID, &approvedAt, &createdAt, &updatedAt,
			); err != nil {
				return fmt.Errorf("scan row: %w", err)
			}

			var approvedByEmpID *domain.EmployeeID
			if approvedByID != nil {
				a := domain.MustNewEmployeeID(*approvedByID)
				approvedByEmpID = &a
			}

			record := domain.RehydrateLeaveRequest(
				domain.MustNewLeaveRequestID(requestID),
				domain.MustNewTenantID(tenantID.String()),
				domain.MustNewEmployeeID(empID),
				domain.MustNewLeaveTypeID(leaveTypeID),
				startDate,
				endDate,
				daysCount,
				domain.LeaveStatus(status),
				reason,
				rejectionReason,
				approvedByEmpID,
				approvedAt,
				createdAt,
				updatedAt,
			)
			records = append(records, record)
		}

		return rows.Err()
	})
	return records, err
}
