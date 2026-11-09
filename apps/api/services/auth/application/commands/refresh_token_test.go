package commands

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
	"uuid"
)

type rotatingTokenMock struct {
	claims    *TokenClaims
	err       error
	issued    bool
	issuedFor []string // roles passed to GenerateTokenPair
}

func (m *rotatingTokenMock) GenerateTokenPair(_ context.Context, _ domain.UserID, _ domain.TenantID, _ string, roles []string) (string, string, int32, error) {
	m.issued = true
	m.issuedFor = roles
	return "access-2", "refresh-2", 900, nil
}

func (m *rotatingTokenMock) ValidateAccessToken(context.Context, string) (*TokenClaims, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *rotatingTokenMock) ConsumeRefreshToken(context.Context, string) (*TokenClaims, error) {
	return m.claims, m.err
}

func (m *rotatingTokenMock) RevokeRefreshToken(context.Context, string) error {
	return fmt.Errorf("not implemented")
}

// currentRolesRepo returns the roles the user has now, not the ones in the old token.
type currentRolesRepo struct {
	*MockUserRepository
	roles []*domain.Role
}

func (r currentRolesRepo) GetRoles(context.Context, domain.TenantID, domain.UserID) ([]*domain.Role, error) {
	return r.roles, nil
}

type refreshFixture struct {
	tenant  *domain.Tenant
	user    *domain.User
	tokens  *rotatingTokenMock
	handler *RefreshTokenHandler
}

func newRefreshFixture(t *testing.T) *refreshFixture {
	t.Helper()
	ctx := context.Background()
	tenantID := domain.MustNewTenantID(uuid.NewV7().String())
	tenant := domain.NewTenant(tenantID, domain.MustNewTenantSlug("acme-corp"), "ACME")
	user, err := domain.NewUser(domain.GenerateUserID(), tenantID, "john@example.com", "ValidPass123!", "John")
	if err != nil {
		t.Fatal(err)
	}
	tenants := NewMockTenantRepository()
	users := NewMockUserRepository()
	_ = tenants.Create(ctx, tenant)
	_ = users.Create(ctx, user)

	tokens := &rotatingTokenMock{claims: &TokenClaims{UserID: user.ID(), TenantID: tenantID, Email: user.Email(), Roles: []string{"hr_admin"}}}
	repo := currentRolesRepo{users, []*domain.Role{domain.NewRole(domain.MustNewRoleID(uuid.NewV7().String()), tenantID, "employee", "", true)}}
	return &refreshFixture{tenant, user, tokens, NewRefreshTokenHandler(tenants, repo, tokens)}
}

func TestRefreshTokenRotatesPairWithCurrentRoles(t *testing.T) {
	f := newRefreshFixture(t)

	result, err := f.handler.Handle(context.Background(), RefreshTokenCommand{RefreshToken: "refresh-1"})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if result.AccessToken != "access-2" || result.RefreshToken != "refresh-2" {
		t.Fatalf("unexpected token pair: %+v", result)
	}
	if len(f.tokens.issuedFor) != 1 || f.tokens.issuedFor[0] != "employee" {
		t.Fatalf("new tokens must carry current roles, not the old claims: %v", f.tokens.issuedFor)
	}
}

func TestRefreshTokenRejectsInvalid(t *testing.T) {
	f := newRefreshFixture(t)
	f.tokens.claims, f.tokens.err = nil, fmt.Errorf("replayed: %w", domain.ErrUnauthenticated)

	_, err := f.handler.Handle(context.Background(), RefreshTokenCommand{RefreshToken: "bad"})
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if f.tokens.issued {
		t.Fatal("no tokens may be issued when consume fails")
	}
}

func TestRefreshTokenRefusesInactiveTenantOrUser(t *testing.T) {
	cases := map[string]func(*refreshFixture){
		"tenant deactivated": func(f *refreshFixture) { f.tenant.Deactivate() },
		"user deactivated":   func(f *refreshFixture) { f.user.Deactivate() },
	}
	for name, deactivate := range cases {
		t.Run(name, func(t *testing.T) {
			f := newRefreshFixture(t)
			deactivate(f)

			_, err := f.handler.Handle(context.Background(), RefreshTokenCommand{RefreshToken: "refresh-1"})
			if !errors.Is(err, domain.ErrUnauthenticated) {
				t.Fatalf("expected unauthenticated, got %v", err)
			}
			if f.tokens.issued {
				t.Fatal("no tokens may be issued")
			}
		})
	}
}

func TestRefreshTokenRefusesDeletedUser(t *testing.T) {
	f := newRefreshFixture(t)
	f.tokens.claims.UserID = domain.GenerateUserID()

	_, err := f.handler.Handle(context.Background(), RefreshTokenCommand{RefreshToken: "refresh-1"})
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}
