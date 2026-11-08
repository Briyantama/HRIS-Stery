package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Briyantama/HRIS-Stery/internal/testdb"
	"github.com/Briyantama/HRIS-Stery/services/attendance/domain"
	"github.com/Briyantama/HRIS-Stery/services/attendance/infrastructure/postgres"
)

// Runs against a migrated throwaway database (skipped without TEST_DATABASE_URL). It proves the
// migrations apply (including the partial unique index), the repository works under RLS, and the
// list/count paging agrees.
func TestRepository(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.NewMigratedURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repo := postgres.NewAttendanceRepository(pool)

	tenant := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	otherTenant := domain.MustNewTenantID("660e8400-e29b-41d4-a716-446655440000")
	emp := domain.GenerateEmployeeID()
	day := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)

	rec, _ := domain.NewAttendanceRecord(tenant, emp, day, "u1")
	if err := rec.CheckIn(time.Now().UTC(), nil, nil, "in"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, rec); err != nil {
		t.Fatalf("create: %v", err)
	}

	dup, _ := domain.NewAttendanceRecord(tenant, emp, day, "u1")
	if err := repo.Create(ctx, dup); err == nil {
		t.Error("second regular record for the same day must violate attendance_unique_per_day")
	}
	over, _ := domain.NewAttendanceRecord(tenant, emp, day, "admin")
	_ = over.Override(nil, nil, domain.StatusAbsent, "r")
	if err := repo.Create(ctx, over); err != nil {
		t.Errorf("override rows may repeat: %v", err)
	}

	got, err := repo.GetByEmployeeAndDate(ctx, tenant, emp, day)
	if err != nil || got == nil || got.ID() != rec.ID() {
		t.Fatalf("GetByEmployeeAndDate = %v, %v; want the regular record", got, err)
	}
	if _, err := repo.GetByID(ctx, otherTenant, rec.ID()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("cross-tenant GetByID err = %v, want ErrNotFound", err)
	}

	out := time.Now().UTC().Add(time.Hour)
	if err := got.CheckOut(out, nil, nil, "out"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	f := domain.AttendanceFilters{Limit: 1}
	page, err := repo.ListByTenant(ctx, tenant, f)
	total, cerr := repo.CountByTenant(ctx, tenant, f)
	if err != nil || cerr != nil || len(page) != 1 || total != 2 {
		t.Errorf("list = %d rows, count = %d, errs %v %v; want 1 row of 2", len(page), total, err, cerr)
	}
	if n, _ := repo.CountByTenant(ctx, otherTenant, f); n != 0 {
		t.Errorf("other tenant sees %d rows", n)
	}
}
