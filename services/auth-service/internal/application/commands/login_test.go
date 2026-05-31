package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
)

// MockTenantRepository for testing
type MockTenantRepository struct {
	tenants map[string]*domain.Tenant
}

func NewMockTenantRepository() *MockTenantRepository {
	return &MockTenantRepository{tenants: make(map[string]*domain.Tenant)}
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	m.tenants[tenant.ID().String()] = tenant
	return nil
}

func (m *MockTenantRepository) GetByID(ctx context.Context, id domain.TenantID) (*domain.Tenant, error) {
	if t, ok := m.tenants[id.String()]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("tenant not found")
}

func (m *MockTenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	for _, t := range m.tenants {
		if t.Slug().String() == slug {
			return t, nil
		}
	}
	return nil, fmt.Errorf("tenant not found")
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	m.tenants[tenant.ID().String()] = tenant
	return nil
}

func (m *MockTenantRepository) Delete(ctx context.Context, id domain.TenantID) error {
	delete(m.tenants, id.String())
	return nil
}

// MockUserRepository for testing
type MockUserRepository struct {
	users map[string]*domain.User
	roles map[string][]domain.RoleID
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domain.User),
		roles: make(map[string][]domain.RoleID),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.users[user.ID().String()] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.UserID) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID().Equals(id) && u.TenantID().Equals(tenantID) {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) GetByTenantAndEmail(ctx context.Context, tenantID domain.TenantID, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.TenantID().Equals(tenantID) && u.Email() == email {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.users[user.ID().String()] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, tenantID domain.TenantID, id domain.UserID) error {
	delete(m.users, id.String())
	return nil
}

func (m *MockUserRepository) GetRoles(ctx context.Context, tenantID domain.TenantID, userID domain.UserID) ([]*domain.Role, error) {
	return []*domain.Role{}, nil
}

func (m *MockUserRepository) SetRoles(ctx context.Context, tenantID domain.TenantID, userID domain.UserID, roleIDs []domain.RoleID) error {
	m.roles[userID.String()] = roleIDs
	return nil
}

// MockTokenService for testing
type MockTokenService struct{}

func (m *MockTokenService) GenerateTokenPair(ctx context.Context, userID domain.UserID, tenantID domain.TenantID, email string, roles []string) (string, string, int32, error) {
	return "mock-access-token", "mock-refresh-token", 900, nil
}

func (m *MockTokenService) ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockTokenService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, int32, error) {
	return "new-access-token", "new-refresh-token", 900, nil
}

func (m *MockTokenService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockTokenService) IsRefreshTokenRevoked(ctx context.Context, tokenHash string) (bool, error) {
	return false, nil
}

// MockEventPublisher for testing
type MockEventPublisher struct {
	events []domain.DomainEvent
}

func (m *MockEventPublisher) PublishAsync(ctx context.Context, event domain.DomainEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventPublisher) PublishSync(ctx context.Context, event domain.DomainEvent) error {
	m.events = append(m.events, event)
	return nil
}

// Tests
func TestLoginSuccess(t *testing.T) {
	ctx := context.Background()

	// Setup
	tenantID := domain.MustNewTenantID(uuid.New().String())
	tenantSlug := domain.MustNewTenantSlug("acme-corp")
	tenant := domain.NewTenant(tenantID, tenantSlug, "ACME Corp")

	userID := domain.GenerateUserID()
	password := "ValidPass123!"
	user, _ := domain.NewUser(userID, tenantID, "john@example.com", password, "John Doe")
	user.VerifyEmail()

	// Initialize mocks
	tenantRepo := NewMockTenantRepository()
	tenantRepo.Create(ctx, tenant)

	userRepo := NewMockUserRepository()
	userRepo.Create(ctx, user)

	tokenSvc := &MockTokenService{}
	eventPub := &MockEventPublisher{}

	handler := NewLoginHandler(tenantRepo, userRepo, tokenSvc, eventPub)

	// Execute
	result, err := handler.Handle(ctx, LoginCommand{
		Email:      "john@example.com",
		Password:   password,
		TenantSlug: "acme-corp",
		IPAddr:     "192.168.1.1",
	})

	// Assert
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if result == nil {
		t.Fatalf("Result should not be nil")
	}
	if result.AccessToken != "mock-access-token" {
		t.Errorf("AccessToken mismatch")
	}
	if result.RefreshToken != "mock-refresh-token" {
		t.Errorf("RefreshToken mismatch")
	}
	if len(eventPub.events) < 1 {
		t.Errorf("Expected at least 1 event published, got %d", len(eventPub.events))
	}
}

func TestLoginFailInvalidTenant(t *testing.T) {
	ctx := context.Background()
	tenantRepo := NewMockTenantRepository()
	userRepo := NewMockUserRepository()
	tokenSvc := &MockTokenService{}
	eventPub := &MockEventPublisher{}

	handler := NewLoginHandler(tenantRepo, userRepo, tokenSvc, eventPub)

	result, err := handler.Handle(ctx, LoginCommand{
		Email:      "john@example.com",
		Password:   "ValidPass123!",
		TenantSlug: "nonexistent",
		IPAddr:     "192.168.1.1",
	})

	if err == nil {
		t.Fatalf("Login should fail for invalid tenant")
	}
	if result != nil {
		t.Fatalf("Result should be nil on error")
	}
	if len(eventPub.events) < 1 {
		t.Errorf("Expected login_failed event")
	}
}

func TestLoginFailInvalidPassword(t *testing.T) {
	ctx := context.Background()

	tenantID := domain.MustNewTenantID(uuid.New().String())
	tenantSlug := domain.MustNewTenantSlug("acme-corp")
	tenant := domain.NewTenant(tenantID, tenantSlug, "ACME Corp")

	userID := domain.GenerateUserID()
	user, _ := domain.NewUser(userID, tenantID, "john@example.com", "ValidPass123!", "John Doe")

	tenantRepo := NewMockTenantRepository()
	tenantRepo.Create(ctx, tenant)

	userRepo := NewMockUserRepository()
	userRepo.Create(ctx, user)

	tokenSvc := &MockTokenService{}
	eventPub := &MockEventPublisher{}

	handler := NewLoginHandler(tenantRepo, userRepo, tokenSvc, eventPub)

	result, err := handler.Handle(ctx, LoginCommand{
		Email:      "john@example.com",
		Password:   "WrongPassword123!",
		TenantSlug: "acme-corp",
		IPAddr:     "192.168.1.1",
	})

	if err == nil {
		t.Fatalf("Login should fail with wrong password")
	}
	if result != nil {
		t.Fatalf("Result should be nil on error")
	}
}
