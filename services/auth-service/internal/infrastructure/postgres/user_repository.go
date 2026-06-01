package postgres

import (
	"context"
	"fmt"
	"time"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository implements domain.UserRepository using PostgreSQL with RLS.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new Postgres user repository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create persists a new user.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(user.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			INSERT INTO auth.users (id, tenant_id, email, password_hash, full_name, is_active, email_verified, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err := tx.Exec(ctx, query,
			user.ID().String(),
			user.TenantID().String(),
			user.Email(),
			user.PasswordHash(),
			user.FullName(),
			user.IsActive(),
			user.EmailVerified(),
			user.CreatedAt(),
			user.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert user: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a user by ID within a tenant.
func (r *UserRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.UserID) (*domain.User, error) {
	var user *domain.User
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, email, password_hash, full_name, is_active, email_verified, last_login_at, last_login_ip_addr, created_at, updated_at
			FROM auth.users
			WHERE id = $1
		`
		var (
			userID, tenantIDStr, email, passwordHash, fullName string
			isActive, emailVerified                            bool
			lastLoginAt                                        *time.Time
			lastLoginIP                                        *string
			createdAt, updatedAt                               time.Time
		)
		rowErr := tx.QueryRow(ctx, query, id.String()).Scan(
			&userID, &tenantIDStr, &email, &passwordHash, &fullName,
			&isActive, &emailVerified, &lastLoginAt, &lastLoginIP, &createdAt, &updatedAt,
		)
		if rowErr == pgx.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		if rowErr != nil {
			return fmt.Errorf("query user: %w", rowErr)
		}

		lastLoginIPStr := ""
		if lastLoginIP != nil {
			lastLoginIPStr = *lastLoginIP
		}

		var rehydrateErr error
		user, rehydrateErr = domain.RehydrateUser(
			domain.MustNewUserID(userID),
			domain.MustNewTenantID(tenantIDStr),
			email, passwordHash, fullName,
			isActive, emailVerified,
			createdAt, updatedAt,
			lastLoginAt, lastLoginIPStr,
		)
		return rehydrateErr
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByTenantAndEmail retrieves a user by tenant and email.
func (r *UserRepository) GetByTenantAndEmail(ctx context.Context, tenantID domain.TenantID, email string) (*domain.User, error) {
	var user *domain.User
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, tenant_id, email, password_hash, full_name, is_active, email_verified, last_login_at, last_login_ip_addr, created_at, updated_at
			FROM auth.users
			WHERE tenant_id = $1 AND email = $2
		`
		var (
			userID, tenantIDStr, emailVal, passwordHash, fullName string
			isActive, emailVerified                               bool
			lastLoginAt                                           *time.Time
			lastLoginIP                                           *string
			createdAt, updatedAt                                  time.Time
		)
		rowErr := tx.QueryRow(ctx, query, tenantID.String(), email).Scan(
			&userID, &tenantIDStr, &emailVal, &passwordHash, &fullName,
			&isActive, &emailVerified, &lastLoginAt, &lastLoginIP, &createdAt, &updatedAt,
		)
		if rowErr == pgx.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		if rowErr != nil {
			return fmt.Errorf("query user by email: %w", rowErr)
		}

		lastLoginIPStr := ""
		if lastLoginIP != nil {
			lastLoginIPStr = *lastLoginIP
		}

		var rehydrateErr error
		user, rehydrateErr = domain.RehydrateUser(
			domain.MustNewUserID(userID),
			domain.MustNewTenantID(tenantIDStr),
			emailVal, passwordHash, fullName,
			isActive, emailVerified,
			createdAt, updatedAt,
			lastLoginAt, lastLoginIPStr,
		)
		return rehydrateErr
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update persists changes to a user.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(user.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			UPDATE auth.users
			SET full_name = $1, is_active = $2, email_verified = $3, last_login_at = $4, last_login_ip_addr = $5, updated_at = $6
			WHERE id = $7
		`
		result, err := tx.Exec(ctx, query,
			user.FullName(),
			user.IsActive(),
			user.EmailVerified(),
			user.LastLoginAt(),
			user.LastLoginIPAddr(),
			user.UpdatedAt(),
			user.ID().String(),
		)
		if err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("user not found")
		}
		return nil
	})
}

// Delete removes a user.
func (r *UserRepository) Delete(ctx context.Context, tenantID domain.TenantID, id domain.UserID) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		result, err := tx.Exec(ctx, `DELETE FROM auth.users WHERE id = $1`, id.String())
		if err != nil {
			return fmt.Errorf("delete user: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("user not found")
		}
		return nil
	})
}

// GetRoles retrieves all roles assigned to a user.
func (r *UserRepository) GetRoles(ctx context.Context, tenantID domain.TenantID, userID domain.UserID) ([]*domain.Role, error) {
	var roles []*domain.Role
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT r.id, r.tenant_id, r.name, r.description, r.is_system, r.created_at
			FROM auth.roles r
			JOIN auth.user_roles ur ON r.id = ur.role_id
			WHERE ur.user_id = $1
		`
		rows, err := tx.Query(ctx, query, userID.String())
		if err != nil {
			return fmt.Errorf("query user roles: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var roleID, roleTenantID, name, description string
			var isSystem bool
			var createdAt time.Time
			if err := rows.Scan(&roleID, &roleTenantID, &name, &description, &isSystem, &createdAt); err != nil {
				return fmt.Errorf("scan role: %w", err)
			}
			roles = append(roles, domain.NewRole(
				domain.MustNewRoleID(roleID),
				domain.MustNewTenantID(roleTenantID),
				name,
				description,
				isSystem,
			))
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// SetRoles updates the roles assigned to a user.
func (r *UserRepository) SetRoles(ctx context.Context, tenantID domain.TenantID, userID domain.UserID, roleIDs []domain.RoleID) error {
	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM auth.user_roles WHERE user_id = $1`, userID.String()); err != nil {
			return fmt.Errorf("delete user roles: %w", err)
		}
		for _, roleID := range roleIDs {
			if _, err := tx.Exec(ctx,
				`INSERT INTO auth.user_roles (user_id, role_id, granted_at) VALUES ($1, $2, NOW())`,
				userID.String(), roleID.String(),
			); err != nil {
				return fmt.Errorf("insert user role: %w", err)
			}
		}
		return nil
	})
}
