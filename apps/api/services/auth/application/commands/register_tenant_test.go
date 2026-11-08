package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
)

type fakeRoleRepo struct{}

func (fakeRoleRepo) Create(context.Context, *domain.Role) error { return nil }
func (fakeRoleRepo) GetByID(context.Context, domain.TenantID, domain.RoleID) (*domain.Role, error) {
	return nil, nil
}
func (fakeRoleRepo) GetByTenantAndName(context.Context, domain.TenantID, string) (*domain.Role, error) {
	return nil, nil
}
func (fakeRoleRepo) ListByTenant(context.Context, domain.TenantID) ([]*domain.Role, error) {
	return nil, nil
}
func (fakeRoleRepo) Update(context.Context, *domain.Role) error { return nil }

type fakePermRepo struct {
	assignErr error
	created   int
}

func (f *fakePermRepo) Create(context.Context, string, string) error { f.created++; return nil }
func (f *fakePermRepo) GetByName(context.Context, string) (domain.Permission, error) {
	return "", nil
}
func (f *fakePermRepo) List(context.Context) ([]domain.Permission, error) { return nil, nil }
func (f *fakePermRepo) AssignToRole(context.Context, domain.RoleID, domain.Permission) error {
	return f.assignErr
}
func (f *fakePermRepo) RemoveFromRole(context.Context, domain.RoleID, domain.Permission) error {
	return nil
}
func (f *fakePermRepo) GetForRole(context.Context, domain.RoleID) ([]domain.Permission, error) {
	return nil, nil
}
func (f *fakePermRepo) GetForRoles(context.Context, []domain.RoleID) ([]domain.Permission, error) {
	return nil, nil
}

func registerCmd() RegisterTenantCommand {
	return RegisterTenantCommand{
		CompanyName:   "Acme",
		TenantSlug:    "acme",
		AdminEmail:    "admin@acme.test",
		AdminPassword: "Str0ng!Passw0rd",
		AdminName:     "Admin",
	}
}

func TestRegisterTenant_AssignToRoleErrorPropagates(t *testing.T) {
	boom := errors.New("assign failed")
	perms := &fakePermRepo{assignErr: boom}
	h := NewRegisterTenantHandler(NewMockTenantRepository(), NewMockUserRepository(), fakeRoleRepo{}, perms, &MockTokenService{}, &MockEventPublisher{})

	if _, err := h.Handle(context.Background(), registerCmd()); !errors.Is(err, boom) {
		t.Fatalf("want wrapped assign error, got %v", err)
	}
}

func TestRegisterTenant_DoesNotCreatePermissions(t *testing.T) {
	perms := &fakePermRepo{}
	h := NewRegisterTenantHandler(NewMockTenantRepository(), NewMockUserRepository(), fakeRoleRepo{}, perms, &MockTokenService{}, &MockEventPublisher{})

	if _, err := h.Handle(context.Background(), registerCmd()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perms.created != 0 {
		t.Fatalf("permissions are seed data, Create called %d times", perms.created)
	}
}
