package postgres

import (
	"context"
	"fmt"
	"time"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AttendanceRepository implements domain.AttendanceRepository using PostgreSQL with RLS.
type AttendanceRepository struct {
	pool *pgxpool.Pool
}

// NewAttendanceRepository creates a new Postgres attendance repository.
func NewAttendanceRepository(pool *pgxpool.Pool) *AttendanceRepository {
	return &AttendanceRepository{pool: pool}
}

// Create persists a new attendance record with RLS enforcement.
func (r *AttendanceRepository) Create(ctx context.Context, attendance *domain.AttendanceRecord) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(attendance.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			INSERT INTO attendance.attendance_records
			(id, tenant_id, employee_id, date, check_in_at, check_out_at,
			 check_in_latitude, check_in_longitude, check_out_latitude, check_out_longitude,
			 status, work_duration_mins, notes, created_by, is_override, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		`
		inLat, inLon := attendance.CheckInLocation()
		outLat, outLon := attendance.CheckOutLocation()
		_, err := tx.Exec(ctx, query,
			attendance.ID().String(),
			attendance.TenantID().String(),
			attendance.EmployeeID().String(),
			attendance.Date(),
			attendance.CheckInAt(),
			attendance.CheckOutAt(),
			inLat,
			inLon,
			outLat,
			outLon,
			string(attendance.Status()),
			attendance.WorkDurationMins(),
			attendance.Notes(),
			attendance.CreatedBy(),
			attendance.IsOverride(),
			attendance.CreatedAt(),
			attendance.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert attendance: %w", err)
		}
		return nil
	})
}

// GetByID retrieves an attendance record by ID within a tenant.
func (r *AttendanceRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.AttendanceID) (*domain.AttendanceRecord, error) {
	var attendance *domain.AttendanceRecord
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, date, check_in_at, check_out_at,
			       check_in_latitude, check_in_longitude, check_out_latitude, check_out_longitude,
			       status, work_duration_mins, notes, created_by, is_override, created_at, updated_at
			FROM attendance.attendance_records
			WHERE id = $1
		`
		var (
			attendanceID, empID string
			date               time.Time
			checkInAt, checkOutAt *time.Time
			checkInLat, checkInLon, checkOutLat, checkOutLon *float64
			status             string
			workDurationMins   *int
			notes, createdBy   string
			isOverride         bool
			createdAt, updatedAt time.Time
		)
		rowErr := tx.QueryRow(ctx, query, id.String()).Scan(
			&attendanceID, &tenantID, &empID, &date, &checkInAt, &checkOutAt,
			&checkInLat, &checkInLon, &checkOutLat, &checkOutLon,
			&status, &workDurationMins, &notes, &createdBy, &isOverride, &createdAt, &updatedAt,
		)
		if rowErr == pgx.ErrNoRows {
			return fmt.Errorf("attendance not found")
		}
		if rowErr != nil {
			return fmt.Errorf("query attendance: %w", rowErr)
		}

		attendance = domain.RehydrateAttendanceRecord(
			domain.MustNewAttendanceID(attendanceID),
			domain.MustNewTenantID(tenantID.String()),
			domain.MustNewEmployeeID(empID),
			date,
			checkInAt,
			checkOutAt,
			checkInLat,
			checkInLon,
			checkOutLat,
			checkOutLon,
			domain.AttendanceStatus(status),
			workDurationMins,
			notes,
			createdBy,
			isOverride,
			createdAt,
			updatedAt,
		)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return attendance, nil
}

// GetByEmployeeAndDate retrieves an attendance record for a specific employee on a specific date.
func (r *AttendanceRepository) GetByEmployeeAndDate(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, date time.Time) (*domain.AttendanceRecord, error) {
	var attendance *domain.AttendanceRecord
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, date, check_in_at, check_out_at,
			       check_in_latitude, check_in_longitude, check_out_latitude, check_out_longitude,
			       status, work_duration_mins, notes, created_by, is_override, created_at, updated_at
			FROM attendance.attendance_records
			WHERE employee_id = $1 AND date = $2
		`
		var (
			attendanceID, empID string
			queryDate          time.Time
			checkInAt, checkOutAt *time.Time
			checkInLat, checkInLon, checkOutLat, checkOutLon *float64
			status             string
			workDurationMins   *int
			notes, createdBy   string
			isOverride         bool
			createdAt, updatedAt time.Time
		)
		rowErr := tx.QueryRow(ctx, query, employeeID.String(), date).Scan(
			&attendanceID, &tenantID, &empID, &queryDate, &checkInAt, &checkOutAt,
			&checkInLat, &checkInLon, &checkOutLat, &checkOutLon,
			&status, &workDurationMins, &notes, &createdBy, &isOverride, &createdAt, &updatedAt,
		)
		if rowErr == pgx.ErrNoRows {
			return nil
		}
		if rowErr != nil {
			return fmt.Errorf("query attendance: %w", rowErr)
		}

		attendance = domain.RehydrateAttendanceRecord(
			domain.MustNewAttendanceID(attendanceID),
			domain.MustNewTenantID(tenantID.String()),
			domain.MustNewEmployeeID(empID),
			queryDate,
			checkInAt,
			checkOutAt,
			checkInLat,
			checkInLon,
			checkOutLat,
			checkOutLon,
			domain.AttendanceStatus(status),
			workDurationMins,
			notes,
			createdBy,
			isOverride,
			createdAt,
			updatedAt,
		)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return attendance, nil
}

// ListByTenant retrieves all attendance records for a tenant with optional filters.
func (r *AttendanceRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID, filters domain.AttendanceFilters) ([]*domain.AttendanceRecord, error) {
	var records []*domain.AttendanceRecord
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, date, check_in_at, check_out_at,
			       check_in_latitude, check_in_longitude, check_out_latitude, check_out_longitude,
			       status, work_duration_mins, notes, created_by, is_override, created_at, updated_at
			FROM attendance.attendance_records
			WHERE 1=1
		`
		args := []interface{}{}
		argCount := 1

		if filters.EmployeeID != nil {
			query += fmt.Sprintf(" AND employee_id = $%d", argCount)
			args = append(args, filters.EmployeeID.String())
			argCount++
		}

		if filters.DateFrom != nil {
			query += fmt.Sprintf(" AND date >= $%d", argCount)
			args = append(args, *filters.DateFrom)
			argCount++
		}

		if filters.DateTo != nil {
			query += fmt.Sprintf(" AND date <= $%d", argCount)
			args = append(args, *filters.DateTo)
			argCount++
		}

		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argCount)
			args = append(args, string(*filters.Status))
			argCount++
		}

		query += " ORDER BY date DESC, created_at DESC"
		if filters.Limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
			args = append(args, filters.Limit, filters.Offset)
		}

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("query attendance records: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				attendanceID, empID string
				date               time.Time
				checkInAt, checkOutAt *time.Time
				checkInLat, checkInLon, checkOutLat, checkOutLon *float64
				status             string
				workDurationMins   *int
				notes, createdBy   string
				isOverride         bool
				createdAt, updatedAt time.Time
			)
			if err := rows.Scan(
				&attendanceID, &tenantID, &empID, &date, &checkInAt, &checkOutAt,
				&checkInLat, &checkInLon, &checkOutLat, &checkOutLon,
				&status, &workDurationMins, &notes, &createdBy, &isOverride, &createdAt, &updatedAt,
			); err != nil {
				return fmt.Errorf("scan row: %w", err)
			}

			record := domain.RehydrateAttendanceRecord(
				domain.MustNewAttendanceID(attendanceID),
				domain.MustNewTenantID(tenantID.String()),
				domain.MustNewEmployeeID(empID),
				date,
				checkInAt,
				checkOutAt,
				checkInLat,
				checkInLon,
				checkOutLat,
				checkOutLon,
				domain.AttendanceStatus(status),
				workDurationMins,
				notes,
				createdBy,
				isOverride,
				createdAt,
				updatedAt,
			)
			records = append(records, record)
		}

		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

// Update persists changes to an attendance record.
func (r *AttendanceRepository) Update(ctx context.Context, attendance *domain.AttendanceRecord) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(attendance.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			UPDATE attendance.attendance_records
			SET check_in_at = $1, check_out_at = $2,
			    check_in_latitude = $3, check_in_longitude = $4,
			    check_out_latitude = $5, check_out_longitude = $6,
			    status = $7, work_duration_mins = $8, notes = $9,
			    is_override = $10, updated_at = $11
			WHERE id = $12
		`
		inLat, inLon := attendance.CheckInLocation()
		outLat, outLon := attendance.CheckOutLocation()

		result, err := tx.Exec(ctx, query,
			attendance.CheckInAt(),
			attendance.CheckOutAt(),
			inLat,
			inLon,
			outLat,
			outLon,
			string(attendance.Status()),
			attendance.WorkDurationMins(),
			attendance.Notes(),
			attendance.IsOverride(),
			attendance.UpdatedAt(),
			attendance.ID().String(),
		)
		if err != nil {
			return fmt.Errorf("update attendance: %w", err)
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf("attendance record not found")
		}

		return nil
	})
}

// ListByEmployee retrieves all attendance records for a specific employee within a date range.
func (r *AttendanceRepository) ListByEmployee(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, from time.Time, to time.Time) ([]*domain.AttendanceRecord, error) {
	var records []*domain.AttendanceRecord
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, date, check_in_at, check_out_at,
			       check_in_latitude, check_in_longitude, check_out_latitude, check_out_longitude,
			       status, work_duration_mins, notes, created_by, is_override, created_at, updated_at
			FROM attendance.attendance_records
			WHERE employee_id = $1 AND date >= $2 AND date <= $3
			ORDER BY date DESC
		`
		rows, err := tx.Query(ctx, query, employeeID.String(), from, to)
		if err != nil {
			return fmt.Errorf("query attendance records: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				attendanceID, empID string
				date               time.Time
				checkInAt, checkOutAt *time.Time
				checkInLat, checkInLon, checkOutLat, checkOutLon *float64
				status             string
				workDurationMins   *int
				notes, createdBy   string
				isOverride         bool
				createdAt, updatedAt time.Time
			)
			if err := rows.Scan(
				&attendanceID, &tenantID, &empID, &date, &checkInAt, &checkOutAt,
				&checkInLat, &checkInLon, &checkOutLat, &checkOutLon,
				&status, &workDurationMins, &notes, &createdBy, &isOverride, &createdAt, &updatedAt,
			); err != nil {
				return fmt.Errorf("scan row: %w", err)
			}

			record := domain.RehydrateAttendanceRecord(
				domain.MustNewAttendanceID(attendanceID),
				domain.MustNewTenantID(tenantID.String()),
				domain.MustNewEmployeeID(empID),
				date,
				checkInAt,
				checkOutAt,
				checkInLat,
				checkInLon,
				checkOutLat,
				checkOutLon,
				domain.AttendanceStatus(status),
				workDurationMins,
				notes,
				createdBy,
				isOverride,
				createdAt,
				updatedAt,
			)
			records = append(records, record)
		}

		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

// ListActiveCheckIns retrieves all currently checked-in records (no check-out) for a tenant.
func (r *AttendanceRepository) ListActiveCheckIns(ctx context.Context, tenantID domain.TenantID) ([]*domain.AttendanceRecord, error) {
	var records []*domain.AttendanceRecord
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, employee_id, date, check_in_at, check_out_at,
			       check_in_latitude, check_in_longitude, check_out_latitude, check_out_longitude,
			       status, work_duration_mins, notes, created_by, is_override, created_at, updated_at
			FROM attendance.attendance_records
			WHERE check_in_at IS NOT NULL AND check_out_at IS NULL AND is_override = false
			ORDER BY check_in_at DESC
		`
		rows, err := tx.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("query active checkins: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				attendanceID, empID string
				date               time.Time
				checkInAt, checkOutAt *time.Time
				checkInLat, checkInLon, checkOutLat, checkOutLon *float64
				status             string
				workDurationMins   *int
				notes, createdBy   string
				isOverride         bool
				createdAt, updatedAt time.Time
			)
			if err := rows.Scan(
				&attendanceID, &tenantID, &empID, &date, &checkInAt, &checkOutAt,
				&checkInLat, &checkInLon, &checkOutLat, &checkOutLon,
				&status, &workDurationMins, &notes, &createdBy, &isOverride, &createdAt, &updatedAt,
			); err != nil {
				return fmt.Errorf("scan row: %w", err)
			}

			record := domain.RehydrateAttendanceRecord(
				domain.MustNewAttendanceID(attendanceID),
				domain.MustNewTenantID(tenantID.String()),
				domain.MustNewEmployeeID(empID),
				date,
				checkInAt,
				checkOutAt,
				checkInLat,
				checkInLon,
				checkOutLat,
				checkOutLon,
				domain.AttendanceStatus(status),
				workDurationMins,
				notes,
				createdBy,
				isOverride,
				createdAt,
				updatedAt,
			)
			records = append(records, record)
		}

		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}
