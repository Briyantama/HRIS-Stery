package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Briyantama/stery-kit/go/pgtenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Briyantama/HRIS-Stery/services/audit/domain"
)

const entryColumns = `entry_id::text, tenant_id::text, actor_id, action, resource_type,
	resource_id::text, description, success, error_message, changes::text, created_at`

// AuditRepository implements domain.AuditRepository using PostgreSQL.
// INSERT-ONLY: audit entries are immutable and append-only.
type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository creates a new Postgres audit repository.
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// Record persists a new audit entry (INSERT-ONLY).
func (r *AuditRepository) Record(ctx context.Context, entry *domain.AuditEntry) error {
	return pgtenant.WithTenantTx(ctx, r.pool, entry.TenantID().String(), "", func(tx pgx.Tx) error {
		var changesJSON *string
		if len(entry.Changes()) > 0 {
			data, err := json.Marshal(entry.Changes())
			if err != nil {
				return fmt.Errorf("marshal changes: %w", err)
			}
			s := string(data)
			changesJSON = &s
		}

		// resource_id is a UUID column: an empty id is stored as NULL.
		_, err := tx.Exec(ctx, `
			INSERT INTO audit.entries (
				entry_id, tenant_id, actor_id, action, resource_type,
				resource_id, description, success, error_message, changes, created_at
			)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6::text, '')::uuid, $7, $8, $9, $10, $11)`,
			entry.ID().String(),
			entry.TenantID().String(),
			entry.ActorID(),
			string(entry.Action()),
			string(entry.ResourceType()),
			entry.ResourceID(),
			entry.Description(),
			entry.Success(),
			entry.ErrorMessage(),
			changesJSON,
			entry.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert audit entry: %w", err)
		}
		return nil
	})
}

// GetByID retrieves an audit entry by ID within a tenant; domain.ErrAuditEntryNotFound when absent.
func (r *AuditRepository) GetByID(ctx context.Context, tenantID domain.TenantID, entryID domain.AuditEntryID) (*domain.AuditEntry, error) {
	var entry *domain.AuditEntry
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx,
			`SELECT `+entryColumns+` FROM audit.entries WHERE entry_id = $1 AND tenant_id = $2`,
			entryID.String(), tenantID.String())
		e, err := scanEntry(row)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrAuditEntryNotFound
		}
		entry = e
		return err
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// Query retrieves audit entries matching filters within a tenant, newest first.
func (r *AuditRepository) Query(ctx context.Context, tenantID domain.TenantID, filters domain.QueryFilters) ([]*domain.AuditEntry, int, error) {
	var entries []*domain.AuditEntry
	var total int

	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		where := "tenant_id = $1"
		args := []any{tenantID.String()}
		add := func(col, val string) {
			if val != "" {
				args = append(args, val)
				where += fmt.Sprintf(" AND %s = $%d", col, len(args))
			}
		}
		add("actor_id", filters.ActorID)
		add("action", string(filters.Action))
		add("resource_type", string(filters.ResourceType))
		add("resource_id::text", filters.ResourceID)

		if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM audit.entries WHERE "+where, args...).Scan(&total); err != nil {
			return fmt.Errorf("count audit entries: %w", err)
		}

		limit := filters.Limit
		if limit <= 0 {
			limit = 100
		}
		if limit > 10000 {
			limit = 10000
		}
		offset := max(filters.Offset, 0)

		args = append(args, limit, offset)
		rows, err := tx.Query(ctx, fmt.Sprintf(
			`SELECT %s FROM audit.entries WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			entryColumns, where, len(args)-1, len(args)), args...)
		if err != nil {
			return fmt.Errorf("query audit entries: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			e, err := scanEntry(rows)
			if err != nil {
				return err
			}
			entries = append(entries, e)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

func scanEntry(row pgx.Row) (*domain.AuditEntry, error) {
	var (
		entryID, tenantID, actorID, action, resourceType string
		resourceID, description, errorMsg, changesJSON   *string
		success                                          bool
		createdAt                                        time.Time
	)
	if err := row.Scan(&entryID, &tenantID, &actorID, &action, &resourceType,
		&resourceID, &description, &success, &errorMsg, &changesJSON, &createdAt); err != nil {
		return nil, fmt.Errorf("scan audit entry: %w", err)
	}

	var changes map[string]string
	if changesJSON != nil {
		if err := json.Unmarshal([]byte(*changesJSON), &changes); err != nil {
			return nil, fmt.Errorf("unmarshal changes: %w", err)
		}
	}
	return domain.RehydrateAuditEntry(
		domain.MustNewAuditEntryID(entryID),
		domain.MustNewTenantID(tenantID),
		actorID,
		domain.AuditAction(action),
		domain.ResourceType(resourceType),
		deref(resourceID),
		deref(description),
		success,
		deref(errorMsg),
		changes,
		createdAt,
	), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
