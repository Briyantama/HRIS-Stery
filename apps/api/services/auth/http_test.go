package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Briyantama/stery-kit/go/authn"

	"github.com/Briyantama/HRIS-Stery/services/auth"
	"github.com/Briyantama/HRIS-Stery/services/auth/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/auth/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
)

const (
	userID   = "11111111-1111-4111-8111-111111111111"
	otherID  = "22222222-2222-4222-8222-222222222222"
	tenantID = "33333333-3333-4333-8333-333333333333"
)

type stubVerifier struct{}

func (stubVerifier) Verify(_ context.Context, token string) (authn.Claims, error) {
	switch token {
	case "user":
		return authn.Claims{UserID: userID, TenantID: tenantID, Roles: []string{"employee"}}, nil
	case "admin":
		return authn.Claims{UserID: otherID, TenantID: tenantID, Roles: []string{"hr_admin"}}, nil
	}
	return authn.Claims{}, errors.New("bad token")
}

// stub implements all five handler views; err is returned by every call.
type stub struct {
	err   error
	query queries.GetPermissionsQuery
}

func (s *stub) Handle(context.Context, commands.LoginCommand) (*commands.LoginResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &commands.LoginResult{UserID: domain.MustNewUserID(userID), TenantID: domain.MustNewTenantID(tenantID), Email: "a@b.c", AccessToken: "at", RefreshToken: "rt", AccessTokenTTL: 900, Roles: []string{"hr_admin"}}, nil
}

type refreshStub struct{ *stub }

func (s refreshStub) Handle(context.Context, commands.RefreshTokenCommand) (*commands.RefreshTokenResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &commands.RefreshTokenResult{AccessToken: "at2", RefreshToken: "rt2", AccessTokenTTL: 900}, nil
}

type revokeStub struct{ *stub }

func (s revokeStub) Handle(context.Context, commands.RevokeTokenCommand) error { return s.err }

type registerStub struct{ *stub }

func (s registerStub) Handle(context.Context, commands.RegisterTenantCommand) (*commands.RegisterTenantResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &commands.RegisterTenantResult{TenantID: domain.MustNewTenantID(tenantID), UserID: domain.MustNewUserID(userID), Email: "a@b.c"}, nil
}

type permStub struct{ *stub }

func (s permStub) Handle(_ context.Context, q queries.GetPermissionsQuery) ([]domain.Permission, error) {
	s.stub.query = q
	if s.err != nil {
		return nil, s.err
	}
	return []domain.Permission{domain.PermissionEmployeeRead}, nil
}

func newServer(s *stub) http.Handler {
	mux := http.NewServeMux()
	auth.NewHTTPHandler(s, refreshStub{s}, revokeStub{s}, registerStub{s}, permStub{s}, stubVerifier{}).Register(mux)
	return mux
}

func do(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRoutesRequireAuth(t *testing.T) {
	h := newServer(&stub{})
	for _, route := range []string{"POST /api/v1/auth/logout", "GET /api/v1/auth/permissions/" + userID} {
		m, p, _ := strings.Cut(route, " ")
		if rec := do(h, m, p, "", `{}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s without token = %d, want 401", route, rec.Code)
		}
	}
}

func TestPermissionsRoleAndTenant(t *testing.T) {
	s := &stub{}
	h := newServer(s)

	if rec := do(h, "GET", "/api/v1/auth/permissions/"+otherID, "user", ""); rec.Code != http.StatusForbidden {
		t.Errorf("other user's permissions as employee = %d, want 403", rec.Code)
	}

	rec := do(h, "GET", "/api/v1/auth/permissions/"+userID+"?tenant_id=99999999-9999-4999-8999-999999999999", "user", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("own permissions = %d, want 200", rec.Code)
	}
	var env struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   *string         `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil || !env.Success || env.Error != nil || string(env.Data) != `{"permissions":["employee:read"]}` {
		t.Errorf("envelope = %s", rec.Body)
	}
	if s.query.TenantID.String() != tenantID {
		t.Errorf("tenant = %q, want the token's %q", s.query.TenantID, tenantID)
	}

	if rec := do(h, "GET", "/api/v1/auth/permissions/"+userID, "admin", ""); rec.Code != http.StatusOK {
		t.Errorf("hr_admin reading another user = %d, want 200", rec.Code)
	}
}

func TestStatusCodes(t *testing.T) {
	good := map[string]string{
		"login":    `{"email":"a@b.c","password":"x","tenant_slug":"acme"}`,
		"refresh":  `{"refresh_token":"rt"}`,
		"logout":   `{"refresh_token":"rt"}`,
		"register": `{"company_name":"Acme","tenant_slug":"acme","admin_email":"a@b.c","admin_password":"x"}`,
	}
	cases := []struct {
		name         string
		err          error
		method, path string
		token, body  string
		want         int
	}{
		{"login ok", nil, "POST", "/api/v1/auth/login", "", good["login"], 200},
		{"login malformed", nil, "POST", "/api/v1/auth/login", "", `{`, 400},
		{"login unknown field", nil, "POST", "/api/v1/auth/login", "", `{"tenant_id":"x"}`, 400},
		{"login missing field", nil, "POST", "/api/v1/auth/login", "", `{"email":"a@b.c"}`, 422},
		{"login bad credentials", fmt.Errorf("x: %w", domain.ErrUnauthenticated), "POST", "/api/v1/auth/login", "", good["login"], 401},
		{"login infra error", errors.New("db down"), "POST", "/api/v1/auth/login", "", good["login"], 500},
		{"refresh ok", nil, "POST", "/api/v1/auth/refresh", "", good["refresh"], 200},
		{"refresh revoked", fmt.Errorf("x: %w", domain.ErrUnauthenticated), "POST", "/api/v1/auth/refresh", "", good["refresh"], 401},
		{"refresh missing", nil, "POST", "/api/v1/auth/refresh", "", `{}`, 422},
		{"logout ok", nil, "POST", "/api/v1/auth/logout", "user", good["logout"], 200},
		{"logout redis down", errors.New("redis down"), "POST", "/api/v1/auth/logout", "user", good["logout"], 500},
		{"register ok", nil, "POST", "/api/v1/auth/register", "", good["register"], 201},
		{"register duplicate slug", fmt.Errorf("x: %w", domain.ErrConflict), "POST", "/api/v1/auth/register", "", good["register"], 409},
		{"register weak password", fmt.Errorf("x: %w", domain.ErrInvalidInput), "POST", "/api/v1/auth/register", "", good["register"], 422},
		{"permissions unknown user", fmt.Errorf("x: %w", domain.ErrNotFound), "GET", "/api/v1/auth/permissions/" + userID, "user", "", 404},
		{"permissions bad id", nil, "GET", "/api/v1/auth/permissions/not-a-uuid", "admin", "", 422},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := do(newServer(&stub{err: c.err}), c.method, c.path, c.token, c.body)
			if rec.Code != c.want {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, c.want, rec.Body)
			}
			if c.want == 500 && !strings.Contains(rec.Body.String(), `"internal error"`) {
				t.Errorf("500 leaked detail: %s", rec.Body)
			}
		})
	}
}
