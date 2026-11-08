package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Briyantama/stery-kit/go/pgtenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Briyantama/HRIS-Stery/services/attendance/domain"
)

const recordColumns = `id, tenant_id, employee_id, date, check_in_at, check_out_at,
	check_in_latitude, check_in_longitude, check_out_latitude, check_out_longitude,
	status, work_duration_mins, notes, created_by, is_override, created_at, updated_at`

// AttendanceRepository implements domain.AttendanceRepository using PostgreSQL with RLS.
type AttendanceRepository struct {
	pool *pgxpool.Pool
}

// NewAttendanceRepository creates a new Postgres attendance repository.
func NewAttendanceRepository(pool *pgxpool.Pool) *AttendanceRepository {
	return &AttendanceRepository{pool: pool}
}

// scanRecord reads one row selected with recordColumns.
func scanRecord(row pgx.Row) (*domain.AttendanceRecord, error) {
	var (
		id, tenantID, employeeID                         string
		date                                             time.Time
		checkInAt, checkOutAt                            *time.Time
		checkInLat, checkInLon, checkOutLat, checkOutLon *float64
		status                                           string
		workDurationMins                                 *int
		notes, createdBy                                 string
		isOverride                                       bool
		createdAt, updatedAt                             time.Time
	)
	if err := row.Scan(
		&id, &tenantID, &employeeID, &date, &checkInAt, &checkOutAt,
		&checkInLat, &checkInLon, &checkOutLat, &checkOutLon,
		&status, &workDurationMins, &notes, &createdBy, &isOverride, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	return domain.RehydrateAttendanceRecord(
		domain.MustNewAttendanceID(id),
		domain.MustNewTenantID(tenantID),
		domain.MustNewEmployeeID(employeeID),
		date, checkInAt, checkOutAt,
		checkInLat, checkInLon, checkOutLat, checkOutLon,
		domain.AttendanceStatus(status), workDurationMins, notes, createdBy, isOverride, createdAt, updatedAt,
	), nil
}

func collect(rows pgx.Rows) ([]*domain.AttendanceRecord, error) {
	defer rows.Close()
	var records []*domain.AttendanceRecord
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scan attendance row: %w", err)
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// Create persists a new attendance record with RLS enforcement.
func (r *AttendanceRepository) Create(ctx context.Context, a *domain.AttendanceRecord) error {
	return pgtenant.WithTenantTx(ctx, r.pool, a.TenantID().String(), "", func(tx pgx.Tx) error {
		inLat, inLon := a.CheckInLocation()
		outLat, outLon := a.CheckOutLocation()
		_, err := tx.Exec(ctx, `
			INSERT INTO attendance.attendance_records (`+recordColumns+`)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
			a.ID().String(), a.TenantID().String(), a.EmployeeID().String(), a.Date(),
			a.CheckInAt(), a.CheckOutAt(), inLat, inLon, outLat, outLon,
			string(a.Status()), a.WorkDurationMins(), a.Notes(), a.CreatedBy(), a.IsOverride(),
			a.CreatedAt(), a.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert attendance: %w", err)
		}
		return nil
	})
}

// GetByID retrieves an attendance record by ID within a tenant; domain.ErrNotFound when absent.
func (r *AttendanceRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.AttendanceID) (*domain.AttendanceRecord, error) {
	var rec *domain.AttendanceRecord
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		var err error
		rec, err = scanRecord(tx.QueryRow(ctx,
			`SELECT `+recordColumns+` FROM attendance.attendance_records WHERE tenant_id = $1 AND id = $2`, tenantID.String(), id.String()))
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("query attendance: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// GetByEmployeeAndDate returns the record of an employee on a date, or (nil, nil) when there is none.
// A regular record wins over manual overrides when both exist.
func (r *AttendanceRepository) GetByEmployeeAndDate(ctx context.Context, tenantID domain.TenantID, employeeID domain.EmployeeID, date time.Time) (*domain.AttendanceRecord, error) {
	var rec *domain.AttendanceRecord
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		var err error
		rec, err = scanRecord(tx.QueryRow(ctx,
			`SELECT `+recordColumns+` FROM attendance.attendance_records
			 WHERE tenant_id = $1 AND employee_id = $2 AND date = $3
			 ORDER BY is_override, created_at LIMIT 1`, tenantID.String(), employeeID.String(), date))
		if errors.Is(err, pgx.ErrNoRows) {
			rec = nil
			return nil
		}
		if err != nil {
			return fmt.Errorf("query attendance: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// where builds the shared WHERE clause of ListByTenant and CountByTenant. The tenant filter is
// explicit on top of RLS, which does not apply to table owners.
func where(tenantID domain.TenantID, f domain.AttendanceFilters) (string, []any) {
	args := []any{tenantID.String()}
	conds := []string{"tenant_id = $1"}
	add := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}
	if f.EmployeeID != nil {
		add("employee_id = $%d", f.EmployeeID.String())
	}
	if f.DateFrom != nil {
		add("date >= $%d", *f.DateFrom)
	}
	if f.DateTo != nil {
		add("date <= $%d", *f.DateTo)
	}
	if f.Status != nil {
		add("status = $%d", string(*f.Status))
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

// ListByTenant retrieves attendance records for a tenant, newest first, with optional filters and paging.
func (r *AttendanceRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID, f domain.AttendanceFilters) ([]*domain.AttendanceRecord, error) {
	var records []*domain.AttendanceRecord
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		cond, args := where(tenantID, f)
		query := `SELECT ` + recordColumns + ` FROM attendance.attendance_records` + cond +
			` ORDER BY date DESC, created_at DESC, id`
		if f.Limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
			args = append(args, f.Limit, f.Offset)
		}
		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("query attendance records: %w", err)
		}
		records, err = collect(rows)
		return err
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CountByTenant counts the records matching the filters, ignoring paging.
func (r *AttendanceRepository) CountByTenant(ctx context.Context, tenantID domain.TenantID, f domain.AttendanceFilters) (int, error) {
	var n int
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		cond, args := where(tenantID, f)
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM attendance.attendance_records`+cond, args...).Scan(&n); err != nil {
			return fmt.Errorf("count attendance records: %w", err)
		}
		return nil
	})
	return n, err
}

// Update persists changes to an attendance record; domain.ErrNotFound when it does not exist.
func (r *AttendanceRepository) Update(ctx context.Context, a *domain.AttendanceRecord) error {
	return pgtenant.WithTenantTx(ctx, r.pool, a.TenantID().String(), "", func(tx pgx.Tx) error {
		inLat, inLon := a.CheckInLocation()
		outLat, outLon := a.CheckOutLocation()
		res, err := tx.Exec(ctx, `
			UPDATE attendance.attendance_records
			SET check_in_at = $1, check_out_at = $2,
			    check_in_latitude = $3, check_in_longitude = $4,
			    check_out_latitude = $5, check_out_longitude = $6,
			    status = $7, work_duration_mins = $8, notes = $9,
			    is_override = $10, updated_at = $11
			WHERE id = $12 AND tenant_id = $13`,
			a.CheckInAt(), a.CheckOutAt(), inLat, inLon, outLat, outLon,
			string(a.Status()), a.WorkDurationMins(), a.Notes(), a.IsOverride(), a.UpdatedAt(), a.ID().String(), a.TenantID().String(),
		)
		if err != nil {
			return fmt.Errorf("update attendance: %w", err)
		}
		if res.RowsAffected() == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
}
