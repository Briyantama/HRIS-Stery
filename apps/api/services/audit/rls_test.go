package audit_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Briyantama/HRIS-Stery/internal/testdb"
)

// audit.entries must be tenant-isolated for the app role, and accept rows beyond the monthly partitions.
func TestEntries_RLSAndDefaultPartition(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.NewMigratedURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	a, b := uuid.NewString(), uuid.NewString()
	// Run as hris_app (NOBYPASSRLS), scoped to one tenant, like the service does.
	asApp := func(tenantID string, fn func(tx pgx.Tx) error) error {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx, `SET LOCAL ROLE hris_app`); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenantID); err != nil {
			return err
		}
		if err := fn(tx); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	insert := func(tenantID, at string) error {
		return asApp(tenantID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO audit.entries (tenant_id, actor_id, action, resource_type, created_at)
				VALUES ($1, 'u', 'X', 'r', $2::timestamptz)`, tenantID, at)
			return err
		})
	}

	if err := insert(a, "2026-01-15"); err != nil {
		t.Fatalf("insert in monthly partition: %v", err)
	}
	if err := insert(a, "2031-01-15"); err != nil {
		t.Fatalf("insert past last partition: %v", err)
	}
	if err := insert(b, "2026-01-15"); err != nil {
		t.Fatalf("insert tenant b: %v", err)
	}
	if err := insert(b, "2026-01-16"); err != nil { // sanity: same tenant ok
		t.Fatal(err)
	}
	// Writing another tenant's row must be rejected by WITH CHECK.
	err = asApp(a, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO audit.entries (tenant_id, actor_id, action, resource_type)
			VALUES ($1, 'u', 'X', 'r')`, b)
		return err
	})
	if err == nil {
		t.Fatal("cross-tenant insert succeeded; want RLS violation")
	}

	var n int
	if err := asApp(a, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT COUNT(*) FROM audit.entries`).Scan(&n)
	}); err != nil || n != 2 {
		t.Fatalf("tenant a sees %d rows, err=%v; want 2", n, err)
	}
}
