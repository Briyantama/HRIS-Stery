// Package postgres provides shared database utilities for HRIS-Stery services.
// All services must use WithTenantTx before executing any tenant-scoped query.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantID is a typed wrapper to prevent accidental raw string comparisons.
type TenantID string

func (t TenantID) String() string { return string(t) }

// WithTenantTx acquires a connection, sets the RLS tenant context, begins a
// transaction, and passes it to fn. The transaction is committed on success
// and rolled back on error. set_config and all queries share one transaction,
// which is required for PgBouncer transaction pooling mode.
func WithTenantTx(ctx context.Context, pool *pgxpool.Pool, tenantID TenantID, fn func(ctx context.Context, tx pgx.Tx) error) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if _, err = tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenantID.String()); err != nil {
		return fmt.Errorf("set RLS tenant context: %w", err)
	}

	if err = fn(ctx, tx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	committed = true
	return nil
}
