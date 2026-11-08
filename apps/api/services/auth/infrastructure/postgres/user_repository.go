package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
	sqlcdb "github.com/Briyantama/HRIS-Stery/services/auth/infrastructure/postgres/sqlc"
	"github.com/Briyantama/stery-kit/go/pgtenant"
	"github.com/google/uuid"
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
	return pgtenant.WithTenantTx(ctx, r.pool, user.TenantID().String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		err := q.InsertUser(ctx, sqlcdb.InsertUserParams{
			ID:            uuid.MustParse(user.ID().String()),
			TenantID:      uuid.MustParse(user.TenantID().String()),
			Email:         user.Email(),
			PasswordHash:  user.PasswordHash(),
			FullName:      user.FullName(),
			IsActive:      user.IsActive(),
			EmailVerified: user.EmailVerified(),
			CreatedAt:     user.CreatedAt(),
			UpdatedAt:     user.UpdatedAt(),
		})
		if err != nil {
			return fmt.Errorf("insert user: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a user by ID within a tenant.
func (r *UserRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.UserID) (*domain.User, error) {
	var user *domain.User
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		row, rowErr := q.GetUserByID(ctx, uuid.MustParse(id.String()))
		if errors.Is(rowErr, pgx.ErrNoRows) {
			return fmt.Errorf("user not found: %w", domain.ErrNotFound)
		}
		if rowErr != nil {
			return fmt.Errorf("query user: %w", rowErr)
		}

		lastLoginIPStr := ""
		if row.LastLoginIpAddr != nil {
			lastLoginIPStr = *row.LastLoginIpAddr
		}

		var rehydrateErr error
		user, rehydrateErr = domain.RehydrateUser(
			domain.MustNewUserID(row.ID.String()),
			domain.MustNewTenantID(row.TenantID.String()),
			row.Email, row.PasswordHash, row.FullName,
			row.IsActive, row.EmailVerified,
			row.CreatedAt, row.UpdatedAt,
			row.LastLoginAt, lastLoginIPStr,
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
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		row, rowErr := q.GetUserByTenantAndEmail(ctx, sqlcdb.GetUserByTenantAndEmailParams{
			TenantID: uuid.MustParse(tenantID.String()),
			Email:    email,
		})
		if errors.Is(rowErr, pgx.ErrNoRows) {
			return fmt.Errorf("user not found: %w", domain.ErrNotFound)
		}
		if rowErr != nil {
			return fmt.Errorf("query user by email: %w", rowErr)
		}

		lastLoginIPStr := ""
		if row.LastLoginIpAddr != nil {
			lastLoginIPStr = *row.LastLoginIpAddr
		}

		var rehydrateErr error
		user, rehydrateErr = domain.RehydrateUser(
			domain.MustNewUserID(row.ID.String()),
			domain.MustNewTenantID(row.TenantID.String()),
			row.Email, row.PasswordHash, row.FullName,
			row.IsActive, row.EmailVerified,
			row.CreatedAt, row.UpdatedAt,
			row.LastLoginAt, lastLoginIPStr,
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
	return pgtenant.WithTenantTx(ctx, r.pool, user.TenantID().String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)

		var lastLoginIP *string
		if ip := user.LastLoginIPAddr(); ip != "" {
			lastLoginIP = &ip
		}

		result, err := q.UpdateUser(ctx, sqlcdb.UpdateUserParams{
			FullName:        user.FullName(),
			IsActive:        user.IsActive(),
			EmailVerified:   user.EmailVerified(),
			LastLoginAt:     user.LastLoginAt(),
			LastLoginIpAddr: lastLoginIP,
			UpdatedAt:       user.UpdatedAt(),
			ID:              uuid.MustParse(user.ID().String()),
		})
		if err != nil {
			return fmt.Errorf("update user: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("user not found: %w", domain.ErrNotFound)
		}
		return nil
	})
}

// Delete removes a user.
func (r *UserRepository) Delete(ctx context.Context, tenantID domain.TenantID, id domain.UserID) error {
	return pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		result, err := q.DeleteUser(ctx, uuid.MustParse(id.String()))
		if err != nil {
			return fmt.Errorf("delete user: %w", err)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("user not found: %w", domain.ErrNotFound)
		}
		return nil
	})
}

// GetRoles retrieves all roles assigned to a user.
func (r *UserRepository) GetRoles(ctx context.Context, tenantID domain.TenantID, userID domain.UserID) ([]*domain.Role, error) {
	var roles []*domain.Role
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		rows, err := q.GetUserRoles(ctx, uuid.MustParse(userID.String()))
		if err != nil {
			return fmt.Errorf("query user roles: %w", err)
		}
		for _, row := range rows {
			desc := ""
			if row.Description != nil {
				desc = *row.Description
			}
			roles = append(roles, domain.NewRole(
				domain.MustNewRoleID(row.ID.String()),
				domain.MustNewTenantID(row.TenantID.String()),
				row.Name,
				desc,
				row.IsSystem,
			))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// SetRoles updates the roles assigned to a user.
func (r *UserRepository) SetRoles(ctx context.Context, tenantID domain.TenantID, userID domain.UserID, roleIDs []domain.RoleID) error {
	return pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		q := sqlcdb.New(tx)
		uid := uuid.MustParse(userID.String())

		if err := q.DeleteUserRoles(ctx, uid); err != nil {
			return fmt.Errorf("delete user roles: %w", err)
		}

		for _, roleID := range roleIDs {
			if err := q.InsertUserRole(ctx, sqlcdb.InsertUserRoleParams{
				UserID: uid,
				RoleID: uuid.MustParse(roleID.String()),
			}); err != nil {
				return fmt.Errorf("insert user role: %w", err)
			}
		}
		return nil
	})
}
