package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Briyantama/HRIS-Stery/internal/testdb"
	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
	"github.com/Briyantama/HRIS-Stery/services/leave/infrastructure/postgres"
)

// Runs against a migrated throwaway database (skipped without TEST_DATABASE_URL). It proves the SQL
// enforces the leave rules the handlers rely on: relative balance writes, the status guard, and the
// overlap check.
func TestLeaveRules(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.NewMigratedURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	requests := postgres.NewLeaveRequestRepository(pool)
	balances := postgres.NewLeaveBalanceRepository(pool)

	tenant := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	emp := domain.GenerateEmployeeID()
	leaveType := domain.GenerateLeaveTypeID()
	bal, _ := domain.NewLeaveBalance(tenant, emp, leaveType, 2026, 10)
	if err := balances.Create(ctx, bal); err != nil {
		t.Fatalf("create balance: %v", err)
	}

	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
	newReq := func(start, end time.Time, days int) *domain.LeaveRequest {
		r, err := domain.NewLeaveRequest(tenant, emp, leaveType, start, end, days, "trip", emp.String())
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	create := func(r *domain.LeaveRequest) error {
		return requests.CreateAndUpdateBalanceAtomically(ctx, r, domain.BalanceDelta{Pending: float64(r.DaysCount())})
	}
	balance := func() (used, pending float64) {
		t.Helper()
		b, err := balances.GetByEmployeeAndType(ctx, tenant, emp, leaveType, 2026)
		if err != nil || b == nil {
			t.Fatalf("get balance: %v, %v", b, err)
		}
		return b.UsedDays(), b.PendingDays()
	}

	// Overlap: 2-4 March is taken, so 4-6 March (shares the 4th) is refused and leaves no trace.
	first := newReq(day(3, 2), day(3, 4), 3)
	if err := create(first); err != nil {
		t.Fatalf("first request: %v", err)
	}
	if err := create(newReq(day(3, 4), day(3, 6), 3)); !errors.Is(err, domain.ErrLeaveOverlapExists) {
		t.Fatalf("overlapping request err = %v, want ErrLeaveOverlapExists", err)
	}
	if _, pending := balance(); pending != 3 {
		t.Fatalf("pending = %v after a refused overlap, want 3", pending)
	}
	if err := create(newReq(day(3, 5), day(3, 6), 2)); err != nil {
		t.Fatalf("adjacent request must be allowed: %v", err)
	}

	// Insufficient balance: 5 pending of 10, so 6 more days must fail atomically.
	if err := create(newReq(day(4, 1), day(4, 6), 6)); !errors.Is(err, domain.ErrInsufficientLeaveBalance) {
		t.Fatalf("over-entitlement err = %v, want ErrInsufficientLeaveBalance", err)
	}
	if n, err := requests.CountByTenant(ctx, tenant, domain.LeaveFilters{}); err != nil || n != 2 {
		t.Fatalf("requests after refused create = %d, %v; want 2", n, err)
	}

	// Status guard: the first decision wins, the second changes neither request nor balance.
	if err := first.Approve(domain.GenerateEmployeeID()); err != nil {
		t.Fatal(err)
	}
	approve := domain.BalanceDelta{Used: 3, Pending: -3}
	if err := requests.TransitionAndUpdateBalanceAtomically(ctx, first, domain.StatusPending, approve); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if err := requests.TransitionAndUpdateBalanceAtomically(ctx, first, domain.StatusPending, approve); !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Fatalf("second decision err = %v, want ErrInvalidStatusTransition", err)
	}
	if used, pending := balance(); used != 3 || pending != 2 {
		t.Fatalf("used/pending = %v/%v after a refused second decision, want 3/2", used, pending)
	}

	// A cancelled request no longer blocks its dates.
	if err := first.Cancel("plans"); err != nil {
		t.Fatal(err)
	}
	if err := requests.TransitionAndUpdateBalanceAtomically(ctx, first, domain.StatusApproved, domain.BalanceDelta{Used: -3}); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if err := create(newReq(day(3, 2), day(3, 3), 2)); err != nil {
		t.Fatalf("dates of a cancelled request must be free: %v", err)
	}
}

// Concurrent creations must all land in pending_days: with absolute writes the last one wins.
func TestConcurrentCreatesDoNotLoseBalanceUpdates(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.NewMigratedURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	requests := postgres.NewLeaveRequestRepository(pool)
	balances := postgres.NewLeaveBalanceRepository(pool)

	tenant := domain.MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	emp := domain.GenerateEmployeeID()
	leaveType := domain.GenerateLeaveTypeID()
	bal, _ := domain.NewLeaveBalance(tenant, emp, leaveType, 2026, 10)
	if err := balances.Create(ctx, bal); err != nil {
		t.Fatal(err)
	}

	const n = 6
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d := time.Date(2026, 5, 1+i, 0, 0, 0, 0, time.UTC)
			r, err := domain.NewLeaveRequest(tenant, emp, leaveType, d, d, 1, "trip", emp.String())
			if err == nil {
				err = requests.CreateAndUpdateBalanceAtomically(ctx, r, domain.BalanceDelta{Pending: 1})
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent create: %v", err)
		}
	}

	got, err := balances.GetByEmployeeAndType(ctx, tenant, emp, leaveType, 2026)
	if err != nil || got == nil {
		t.Fatalf("get balance: %v, %v", got, err)
	}
	if got.PendingDays() != n {
		t.Fatalf("pending = %v, want %d", got.PendingDays(), n)
	}
}
