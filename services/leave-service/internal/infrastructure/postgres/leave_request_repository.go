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
			requestID, empID, leaveTypeID   string
			startDate, endDate              time.Time
			daysCount                       int
			status, reason, rejectionReason string
			approvedByID                    *string
			approvedAt                      *time.Time
			createdAt, updatedAt            time.Time
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
				requestID, empID, leaveTypeID   string
				startDate, endDate              time.Time
				daysCount                       int
				status, reason, rejectionReason string
				approvedByID                    *string
				approvedAt                      *time.Time
				createdAt, updatedAt            time.Time
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
		return r.updateWithTx(ctx, tx, request)
	})
}

// updateWithTx persists changes to a leave request within an existing transaction.
// Supports atomic operations wrapping multiple repositories.
func (r *LeaveRequestRepository) updateWithTx(ctx context.Context, tx pgx.Tx, request *domain.LeaveRequest) error {
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
}

// CreateWithTx persists a new leave request within an existing transaction.
// Supports atomic operations wrapping multiple repositories.
func (r *LeaveRequestRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, request *domain.LeaveRequest) error {
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
}

// GetByIDWithTx retrieves a leave request by ID within an existing transaction.
// Supports atomic operations wrapping multiple repositories.
func (r *LeaveRequestRepository) GetByIDWithTx(ctx context.Context, tx pgx.Tx, tenantID domain.TenantID, id domain.LeaveRequestID) (*domain.LeaveRequest, error) {
	query := `
		SELECT id, tenant_id, employee_id, leave_type_id, start_date, end_date, days_count, status, reason,
		       rejection_reason, approved_by_id, approved_at, created_at, updated_at
		FROM leave.leave_requests
		WHERE id = $1
	`
	var (
		requestID, empID, leaveTypeID   string
		startDate, endDate              time.Time
		daysCount                       int
		status, reason, rejectionReason string
		approvedByID                    *string
		approvedAt                      *time.Time
		createdAt, updatedAt            time.Time
	)
	rowErr := tx.QueryRow(ctx, query, id.String()).Scan(
		&requestID, &tenantID, &empID, &leaveTypeID, &startDate, &endDate, &daysCount,
		&status, &reason, &rejectionReason, &approvedByID, &approvedAt, &createdAt, &updatedAt,
	)
	if rowErr == pgx.ErrNoRows {
		return nil, nil
	}
	if rowErr != nil {
		return nil, fmt.Errorf("query leave request: %w", rowErr)
	}

	var approvedByEmpID *domain.EmployeeID
	if approvedByID != nil {
		a := domain.MustNewEmployeeID(*approvedByID)
		approvedByEmpID = &a
	}

	request := domain.RehydrateLeaveRequest(
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
	return request, nil
}

// CreateAndUpdateBalanceAtomically wraps create leave request and update balance in a single transaction.
// This ensures all-or-nothing semantics: if creation succeeds but balance update fails, the request insert rolls back.
func (r *LeaveRequestRepository) CreateAndUpdateBalanceAtomically(
	ctx context.Context,
	request *domain.LeaveRequest,
	balance *domain.LeaveBalance,
	balanceRepo domain.LeaveBalanceRepository,
) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(request.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		if err := r.CreateWithTx(ctx, tx, request); err != nil {
			return fmt.Errorf("create leave request: %w", err)
		}
		// Cast to concrete type to access updateWithTx method
		if repo, ok := balanceRepo.(*LeaveBalanceRepository); ok {
			if err := repo.updateWithTx(ctx, tx, balance); err != nil {
				return fmt.Errorf("update balance: %w", err)
			}
		}
		return nil
	})
}

// UpdateAndGetBalanceAtomically updates a leave request and retrieves updated balance in a single transaction.
// Used for approve/reject/cancel operations where balance state must be consistent.
func (r *LeaveRequestRepository) UpdateAndGetBalanceAtomically(
	ctx context.Context,
	request *domain.LeaveRequest,
	balanceRepo *LeaveBalanceRepository,
	tenantID domain.TenantID,
	employeeID domain.EmployeeID,
	leaveTypeID domain.LeaveTypeID,
	year int,
) (*domain.LeaveBalance, error) {
	var balance *domain.LeaveBalance
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(request.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		// Get balance for update
		b, err := balanceRepo.GetByEmployeeAndTypeWithTx(ctx, tx, tenantID, employeeID, leaveTypeID, year)
		if err != nil {
			return fmt.Errorf("get balance: %w", err)
		}
		if b == nil {
			return fmt.Errorf("balance not found")
		}
		balance = b

		// Update request status
		if err := r.updateWithTx(ctx, tx, request); err != nil {
			return fmt.Errorf("update request: %w", err)
		}

		return nil
	})
	return balance, err
}

// ApproveAndUpdateBalanceAtomically approves a leave request and updates balance in a single transaction.
// This ensures all-or-nothing semantics: if balance update fails, request approval rolls back.
func (r *LeaveRequestRepository) ApproveAndUpdateBalanceAtomically(
	ctx context.Context,
	request *domain.LeaveRequest,
	balance *domain.LeaveBalance,
	balanceRepo domain.LeaveBalanceRepository,
) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(request.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		// Update request status (approve)
		if err := r.updateWithTx(ctx, tx, request); err != nil {
			return fmt.Errorf("update request: %w", err)
		}

		// Update balance (pending -> used)
		if repo, ok := balanceRepo.(*LeaveBalanceRepository); ok {
			if err := repo.updateWithTx(ctx, tx, balance); err != nil {
				return fmt.Errorf("update balance: %w", err)
			}
		}

		return nil
	})
}

// RejectAndUpdateBalanceAtomically rejects a leave request and updates balance in a single transaction.
// This ensures all-or-nothing semantics: if balance update fails, request rejection rolls back.
func (r *LeaveRequestRepository) RejectAndUpdateBalanceAtomically(
	ctx context.Context,
	request *domain.LeaveRequest,
	balance *domain.LeaveBalance,
	balanceRepo domain.LeaveBalanceRepository,
) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(request.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		// Update request status (reject)
		if err := r.updateWithTx(ctx, tx, request); err != nil {
			return fmt.Errorf("update request: %w", err)
		}

		// Update balance (remove pending)
		if repo, ok := balanceRepo.(*LeaveBalanceRepository); ok {
			if err := repo.updateWithTx(ctx, tx, balance); err != nil {
				return fmt.Errorf("update balance: %w", err)
			}
		}

		return nil
	})
}

// CancelAndUpdateBalanceAtomically cancels a leave request and updates balance in a single transaction.
// This ensures all-or-nothing semantics: if balance update fails, request cancellation rolls back.
func (r *LeaveRequestRepository) CancelAndUpdateBalanceAtomically(
	ctx context.Context,
	request *domain.LeaveRequest,
	balance *domain.LeaveBalance,
	balanceRepo domain.LeaveBalanceRepository,
) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(request.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		// Update request status (cancel)
		if err := r.updateWithTx(ctx, tx, request); err != nil {
			return fmt.Errorf("update request: %w", err)
		}

		// Update balance (remove used or pending)
		if repo, ok := balanceRepo.(*LeaveBalanceRepository); ok {
			if err := repo.updateWithTx(ctx, tx, balance); err != nil {
				return fmt.Errorf("update balance: %w", err)
			}
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
				requestID, empID, leaveTypeID   string
				startDate, endDate              time.Time
				daysCount                       int
				status, reason, rejectionReason string
				approvedByID                    *string
				approvedAt                      *time.Time
				createdAt, updatedAt            time.Time
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
