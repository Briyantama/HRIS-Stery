package auth

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/httpx"

	"github.com/Briyantama/HRIS-Stery/services/auth/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/auth/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
)

const maxBody = 1 << 20

// Consumer-side views of the command and query handlers.
type (
	loginer interface {
		Handle(ctx context.Context, cmd commands.LoginCommand) (*commands.LoginResult, error)
	}
	refresher interface {
		Handle(ctx context.Context, cmd commands.RefreshTokenCommand) (*commands.RefreshTokenResult, error)
	}
	revoker interface {
		Handle(ctx context.Context, cmd commands.RevokeTokenCommand) error
	}
	registrar interface {
		Handle(ctx context.Context, cmd commands.RegisterTenantCommand) (*commands.RegisterTenantResult, error)
	}
	permissionLister interface {
		Handle(ctx context.Context, q queries.GetPermissionsQuery) ([]domain.Permission, error)
	}
)

// HTTPHandler serves the auth routes under /api/v1/auth.
type HTTPHandler struct {
	login       loginer
	refresh     refresher
	revoke      revoker
	register    registrar
	permissions permissionLister
	verifier    authn.Verifier
}

// NewHTTPHandler wires the handlers; verifier guards every route that is not public.
func NewHTTPHandler(login loginer, refresh refresher, revoke revoker, register registrar, permissions permissionLister, verifier authn.Verifier) *HTTPHandler {
	return &HTTPHandler{login: login, refresh: refresh, revoke: revoke, register: register, permissions: permissions, verifier: verifier}
}

// Register mounts the routes. login, refresh and register are public: the credential is in the body.
func (h *HTTPHandler) Register(mux *http.ServeMux) {
	authed := authn.Authenticate(h.verifier)
	mux.HandleFunc("POST /api/v1/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.handleRefresh)
	mux.HandleFunc("POST /api/v1/auth/register", h.handleRegister)
	mux.Handle("POST /api/v1/auth/logout", authed(http.HandlerFunc(h.handleLogout)))
	mux.Handle("GET /api/v1/auth/permissions/{user_id}", authed(http.HandlerFunc(h.handlePermissions)))
}

type tokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

type loginResponse struct {
	tokenPairResponse
	Claims claimsResponse `json:"claims"`
}

type claimsResponse struct {
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

func (h *HTTPHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		TenantSlug string `json:"tenant_slug"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	if body.Email == "" || body.Password == "" || body.TenantSlug == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "email, password and tenant_slug are required")
		return
	}
	res, err := h.login.Handle(r.Context(), commands.LoginCommand{
		Email: body.Email, Password: body.Password, TenantSlug: body.TenantSlug, IPAddr: clientIP(r),
	})
	if err != nil {
		fail(w, r, err, "invalid credentials")
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, loginResponse{
		tokenPairResponse: tokenPairResponse{AccessToken: res.AccessToken, RefreshToken: res.RefreshToken, ExpiresIn: res.AccessTokenTTL},
		Claims:            claimsResponse{UserID: res.UserID.String(), TenantID: res.TenantID.String(), Email: res.Email, Roles: res.Roles},
	})
}

func (h *HTTPHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	if body.RefreshToken == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "refresh_token is required")
		return
	}
	res, err := h.refresh.Handle(r.Context(), commands.RefreshTokenCommand{RefreshToken: body.RefreshToken})
	if err != nil {
		fail(w, r, err, "invalid refresh token")
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, tokenPairResponse{AccessToken: res.AccessToken, RefreshToken: res.RefreshToken, ExpiresIn: res.AccessTokenTTL})
}

func (h *HTTPHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	if body.RefreshToken == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "refresh_token is required")
		return
	}
	if err := h.revoke.Handle(r.Context(), commands.RevokeTokenCommand{RefreshToken: body.RefreshToken}); err != nil {
		fail(w, r, err, "invalid refresh token")
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, map[string]bool{"revoked": true})
}

func (h *HTTPHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CompanyName   string `json:"company_name"`
		TenantSlug    string `json:"tenant_slug"`
		AdminEmail    string `json:"admin_email"`
		AdminPassword string `json:"admin_password"`
		AdminName     string `json:"admin_name"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	if body.CompanyName == "" || body.TenantSlug == "" || body.AdminEmail == "" || body.AdminPassword == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "company_name, tenant_slug, admin_email and admin_password are required")
		return
	}
	res, err := h.register.Handle(r.Context(), commands.RegisterTenantCommand{
		CompanyName: body.CompanyName, TenantSlug: body.TenantSlug, AdminEmail: body.AdminEmail,
		AdminPassword: body.AdminPassword, AdminName: body.AdminName,
	})
	if err != nil {
		fail(w, r, err, "unauthorized")
		return
	}
	httpx.WriteSuccess(w, http.StatusCreated, map[string]string{
		"tenant_id": res.TenantID.String(), "user_id": res.UserID.String(), "email": res.Email,
	})
}

// handlePermissions lists a user's permissions. Callers may read their own; hr_admin may read anyone's
// in their tenant. The tenant always comes from the token, so another tenant's user is simply not found.
func (h *HTTPHandler) handlePermissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	target := r.PathValue("user_id")
	if !strings.EqualFold(target, claims.UserID) && !hasRole(claims.Roles, "hr_admin") {
		httpx.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	userID, err := domain.NewUserID(target)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid user_id")
		return
	}
	tenantID, err := domain.NewTenantID(claims.TenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	perms, err := h.permissions.Handle(r.Context(), queries.GetPermissionsQuery{UserID: userID, TenantID: tenantID})
	if err != nil {
		fail(w, r, err, "unauthorized")
		return
	}
	out := make([]string, len(perms))
	for i, p := range perms {
		out[i] = p.String()
	}
	httpx.WriteSuccess(w, http.StatusOK, map[string][]string{"permissions": out})
}

// fail maps a handler error to a status and a fixed message. Anything unrecognised is a 500 whose
// cause is logged, never sent.
func fail(w http.ResponseWriter, r *http.Request, err error, unauthenticated string) {
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		httpx.WriteError(w, http.StatusUnauthorized, unauthenticated)
	case errors.Is(err, domain.ErrInvalidInput):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid input")
	case errors.Is(err, domain.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not found")
	case errors.Is(err, domain.ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "already exists")
	default:
		args := []any{"path", r.URL.Path, "error", err}
		if c, ok := authn.FromContext(r.Context()); ok {
			args = append(args, "tenant_id", c.TenantID)
		}
		slog.ErrorContext(r.Context(), "auth request failed", args...)
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
	}
}

func hasRole(roles []string, want string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, want) {
			return true
		}
	}
	return false
}

// clientIP is the connection address; httpx.TrustedProxies has already replaced it with the
// forwarded client when the peer is a trusted proxy.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
