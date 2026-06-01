package postgres

import (
	"context"
	"fmt"
	"time"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PositionRepository implements domain.PositionRepository using PostgreSQL with RLS.
type PositionRepository struct {
	pool *pgxpool.Pool
}

// NewPositionRepository creates a new Postgres position repository.
func NewPositionRepository(pool *pgxpool.Pool) *PositionRepository {
	return &PositionRepository{pool: pool}
}

// Create persists a new position.
func (r *PositionRepository) Create(ctx context.Context, position *domain.Position) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(position.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			INSERT INTO employee.positions (id, tenant_id, title, description, level, created_at)
			VALUES ($1, $2, $3, $4, $5, now())
		`
		_, err := tx.Exec(ctx, query,
			position.ID().String(),
			position.TenantID().String(),
			position.Title(),
			position.Description(),
			string(position.Level()),
		)
		if err != nil {
			return fmt.Errorf("insert position: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a position by ID within a tenant.
func (r *PositionRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.PositionID) (*domain.Position, error) {
	var position *domain.Position
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, title, description, level, created_at
			FROM employee.positions
			WHERE id = $1
		`
		var (
			posID              string
			title, desc, level string
			createdAt          time.Time
		)
		rowErr := tx.QueryRow(ctx, query, id.String()).Scan(&posID, &tenantID, &title, &desc, &level, &createdAt)
		if rowErr == pgx.ErrNoRows {
			return fmt.Errorf("position not found")
		}
		if rowErr != nil {
			return fmt.Errorf("query position: %w", rowErr)
		}

		pos, err := domain.RehydratePosition(
			domain.MustNewPositionID(posID),
			domain.MustNewTenantID(tenantID.String()),
			title, desc, domain.PositionLevel(level), createdAt,
		)
		if err != nil {
			return fmt.Errorf("rehydrate position: %w", err)
		}
		position = pos
		return nil
	})
	if err != nil {
		return nil, err
	}
	return position, nil
}

// ListByTenant retrieves all positions for a tenant.
func (r *PositionRepository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]*domain.Position, error) {
	var positions []*domain.Position
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, title, description, level, created_at
			FROM employee.positions
			WHERE tenant_id = $1
			ORDER BY title
		`
		rows, err := tx.Query(ctx, query, tenantID.String())
		if err != nil {
			return fmt.Errorf("query positions: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				posID              string
				title, desc, level string
				createdAt          time.Time
			)
			if err := rows.Scan(&posID, &tenantID, &title, &desc, &level, &createdAt); err != nil {
				return fmt.Errorf("scan position: %w", err)
			}

			pos, err := domain.RehydratePosition(
				domain.MustNewPositionID(posID),
				domain.MustNewTenantID(tenantID.String()),
				title, desc, domain.PositionLevel(level), createdAt,
			)
			if err != nil {
				return fmt.Errorf("rehydrate position: %w", err)
			}
			positions = append(positions, pos)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return positions, nil
}

// Update persists changes to a position.
func (r *PositionRepository) Update(ctx context.Context, position *domain.Position) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(position.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			UPDATE employee.positions
			SET title = $1, description = $2, level = $3
			WHERE id = $4
		`
		result, err := tx.Exec(ctx, query,
			position.Title(),
			position.Description(),
			string(position.Level()),
			position.ID().String(),
		)
		if err != nil {
			return fmt.Errorf("update position: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("position not found")
		}
		return nil
	})
}
