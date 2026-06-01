package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(entry.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		// Marshal changes JSONB
		var changesJSON *string
		if entry.Changes() != nil && len(entry.Changes()) > 0 {
			data, err := json.Marshal(entry.Changes())
			if err != nil {
				return fmt.Errorf("marshal changes: %w", err)
			}
			changesStr := string(data)
			changesJSON = &changesStr
		}

		query := `
			INSERT INTO audit.entries (
				entry_id, tenant_id, actor_id, action, resource_type,
				resource_id, description, success, error_message, changes, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`

		_, err := tx.Exec(ctx, query,
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

// GetByID retrieves an audit entry by ID within a tenant.
func (r *AuditRepository) GetByID(ctx context.Context, tenantID domain.TenantID, entryID domain.AuditEntryID) (*domain.AuditEntry, error) {
	var entry *domain.AuditEntry
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT entry_id, tenant_id, actor_id, action, resource_type,
				   resource_id, description, success, error_message, changes, created_at
			FROM audit.entries
			WHERE entry_id = $1 AND tenant_id = $2
		`

		var (
			entryIDStr, tenantIDStr, actorID, actionStr, resourceTypeStr string
			resourceID, description, errorMsg                            *string
			success                                                      bool
			changesJSON                                                  *string
			createdAt                                                    interface{}
		)

		rowErr := tx.QueryRow(ctx, query, entryID.String(), tenantID.String()).Scan(
			&entryIDStr, &tenantIDStr, &actorID, &actionStr, &resourceTypeStr,
			&resourceID, &description, &success, &errorMsg, &changesJSON, &createdAt,
		)
		if rowErr == pgx.ErrNoRows {
			return domain.ErrAuditEntryNotFound
		}
		if rowErr != nil {
			return fmt.Errorf("query audit entry: %w", rowErr)
		}

		// Unmarshal changes
		var changes map[string]string
		if changesJSON != nil {
			if err := json.Unmarshal([]byte(*changesJSON), &changes); err != nil {
				return fmt.Errorf("unmarshal changes: %w", err)
			}
		}

		// Reconstruct entry from stored data
		resourceIDStr := ""
		if resourceID != nil {
			resourceIDStr = *resourceID
		}
		descriptionStr := ""
		if description != nil {
			descriptionStr = *description
		}
		errorMsgStr := ""
		if errorMsg != nil {
			errorMsgStr = *errorMsg
		}

		var rehydrateErr error
		entry, rehydrateErr = domain.RehydrateAuditEntry(
			domain.MustNewAuditEntryID(entryIDStr),
			domain.MustNewTenantID(tenantIDStr),
			actorID,
			domain.AuditAction(actionStr),
			domain.ResourceType(resourceTypeStr),
			resourceIDStr,
			descriptionStr,
			success,
			errorMsgStr,
			changes,
			createdAt,
		)
		return rehydrateErr
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// Query retrieves audit entries matching filters within a tenant.
func (r *AuditRepository) Query(ctx context.Context, tenantID domain.TenantID, filters domain.QueryFilters) ([]*domain.AuditEntry, int, error) {
	var entries []*domain.AuditEntry
	var totalCount int

	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		// Build WHERE clause dynamically
		whereClause := "tenant_id = $1"
		args := []interface{}{tenantID.String()}
		argIdx := 2

		if filters.ActorID != "" {
			whereClause += fmt.Sprintf(" AND actor_id = $%d", argIdx)
			args = append(args, filters.ActorID)
			argIdx++
		}
		if filters.Action != "" {
			whereClause += fmt.Sprintf(" AND action = $%d", argIdx)
			args = append(args, string(filters.Action))
			argIdx++
		}
		if filters.ResourceType != "" {
			whereClause += fmt.Sprintf(" AND resource_type = $%d", argIdx)
			args = append(args, string(filters.ResourceType))
			argIdx++
		}
		if filters.ResourceID != "" {
			whereClause += fmt.Sprintf(" AND resource_id = $%d", argIdx)
			args = append(args, filters.ResourceID)
			argIdx++
		}

		// Get total count
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit.entries WHERE %s", whereClause)
		countErr := tx.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
		if countErr != nil {
			return fmt.Errorf("count audit entries: %w", countErr)
		}

		// Get paginated results
		limit := filters.Limit
		if limit <= 0 {
			limit = 100 // default limit
		}
		if limit > 10000 {
			limit = 10000 // max limit
		}

		offset := filters.Offset
		if offset < 0 {
			offset = 0
		}

		selectQuery := fmt.Sprintf(`
			SELECT entry_id, tenant_id, actor_id, action, resource_type,
				   resource_id, description, success, error_message, changes, created_at
			FROM audit.entries
			WHERE %s
			ORDER BY created_at DESC
			LIMIT $%d OFFSET $%d
		`, whereClause, argIdx, argIdx+1)

		args = append(args, limit, offset)

		rows, err := tx.Query(ctx, selectQuery, args...)
		if err != nil {
			return fmt.Errorf("query audit entries: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				entryIDStr, tenantIDStr, actorID, actionStr, resourceTypeStr string
				resourceID, description, errorMsg                            *string
				success                                                      bool
				changesJSON                                                  *string
				createdAt                                                    interface{}
			)

			err := rows.Scan(
				&entryIDStr, &tenantIDStr, &actorID, &actionStr, &resourceTypeStr,
				&resourceID, &description, &success, &errorMsg, &changesJSON, &createdAt,
			)
			if err != nil {
				return fmt.Errorf("scan audit entry: %w", err)
			}

			// Unmarshal changes
			var changes map[string]string
			if changesJSON != nil {
				if err := json.Unmarshal([]byte(*changesJSON), &changes); err != nil {
					return fmt.Errorf("unmarshal changes: %w", err)
				}
			}

			resourceIDStr := ""
			if resourceID != nil {
				resourceIDStr = *resourceID
			}
			descriptionStr := ""
			if description != nil {
				descriptionStr = *description
			}
			errorMsgStr := ""
			if errorMsg != nil {
				errorMsgStr = *errorMsg
			}

			entry, rehydrateErr := domain.RehydrateAuditEntry(
				domain.MustNewAuditEntryID(entryIDStr),
				domain.MustNewTenantID(tenantIDStr),
				actorID,
				domain.AuditAction(actionStr),
				domain.ResourceType(resourceTypeStr),
				resourceIDStr,
				descriptionStr,
				success,
				errorMsgStr,
				changes,
				createdAt,
			)
			if rehydrateErr != nil {
				return rehydrateErr
			}
			entries = append(entries, entry)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, 0, err
	}
	return entries, totalCount, nil
}
