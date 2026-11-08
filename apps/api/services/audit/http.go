package audit

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/httpx"
	"github.com/google/uuid"

	"github.com/Briyantama/HRIS-Stery/services/audit/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/audit/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/audit/domain"
)

const (
	maxBody        = 1 << 20
	defaultPerPage = 20
	maxPerPage     = 100
	maxFieldLen    = 50 // audit.entries.action / resource_type are VARCHAR(50)
)

type recorder interface {
	Handle(ctx context.Context, cmd commands.RecordAuditCommand) (*commands.RecordAuditResult, error)
}

type trailQuerier interface {
	Handle(ctx context.Context, f queries.AuditQueryFilters) (*queries.QueryAuditTrailResult, error)
}

type entryGetter interface {
	Handle(ctx context.Context, tenantID, entryID string) (*queries.AuditEntryDTO, error)
}

// HTTPHandler serves the audit trail routes. Every route requires the hr_admin role.
type HTTPHandler struct {
	record   recorder
	query    trailQuerier
	get      entryGetter
	verifier authn.Verifier
}

func NewHTTPHandler(record recorder, query trailQuerier, get entryGetter, verifier authn.Verifier) *HTTPHandler {
	return &HTTPHandler{record: record, query: query, get: get, verifier: verifier}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	guard := func(f http.HandlerFunc) http.Handler {
		return authn.Authenticate(h.verifier)(authn.WithRole("hr_admin")(f))
	}
	mux.Handle("POST /api/v1/audit/entries", guard(h.recordEntry))
	mux.Handle("GET /api/v1/audit/entries", guard(h.listEntries))
	mux.Handle("GET /api/v1/audit/entries/{id}", guard(h.getEntry))
}

type entryResponse struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	ActorID      string            `json:"actor_id"`
	Action       string            `json:"action"`
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id"`
	Description  string            `json:"description"`
	Success      bool              `json:"success"`
	ErrorMessage string            `json:"error_message"`
	Changes      map[string]string `json:"changes"`
	CreatedAt    time.Time         `json:"created_at"`
}

func toResponse(d queries.AuditEntryDTO) entryResponse {
	return entryResponse{
		ID: d.ID, TenantID: d.TenantID, ActorID: d.ActorID,
		Action: string(d.Action), ResourceType: string(d.ResourceType), ResourceID: d.ResourceID,
		Description: d.Description, Success: d.Success, ErrorMessage: d.ErrorMessage,
		Changes: d.Changes, CreatedAt: d.CreatedAt,
	}
}

func (h *HTTPHandler) recordEntry(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.FromContext(r.Context())
	if !ok || claims.UserID == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Tenant and actor come from the token; a body that carries them is rejected as an unknown field.
	var body struct {
		Action       string            `json:"action"`
		ResourceType string            `json:"resource_type"`
		ResourceID   string            `json:"resource_id"`
		Description  string            `json:"description"`
		Success      bool              `json:"success"`
		ErrorMessage string            `json:"error_message"`
		Changes      map[string]string `json:"changes"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	if body.Action == "" || len(body.Action) > maxFieldLen ||
		body.ResourceType == "" || len(body.ResourceType) > maxFieldLen {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "action and resource_type are required")
		return
	}
	if body.ResourceID != "" {
		if _, err := uuid.Parse(body.ResourceID); err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "resource_id must be a uuid")
			return
		}
	}

	res, err := h.record.Handle(r.Context(), commands.RecordAuditCommand{
		TenantID:     claims.TenantID,
		ActorID:      claims.UserID,
		Action:       domain.AuditAction(body.Action),
		ResourceType: domain.ResourceType(body.ResourceType),
		ResourceID:   body.ResourceID,
		Description:  body.Description,
		Success:      body.Success,
		ErrorMessage: body.ErrorMessage,
		Changes:      body.Changes,
	})
	if err != nil {
		h.fail(w, r, claims.TenantID, err)
		return
	}
	entry, err := h.get.Handle(r.Context(), claims.TenantID, res.EntryID)
	if err != nil {
		h.fail(w, r, claims.TenantID, err)
		return
	}
	w.Header().Set("Location", "/api/v1/audit/entries/"+entry.ID)
	httpx.WriteSuccess(w, http.StatusCreated, toResponse(*entry))
}

func (h *HTTPHandler) listEntries(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	q := r.URL.Query()
	if s := q.Get("resource_id"); s != "" {
		if _, err := uuid.Parse(s); err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "resource_id must be a uuid")
			return
		}
	}
	page, perPage, ok := pagination(q.Get("page"), q.Get("per_page"))
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid pagination")
		return
	}

	res, err := h.query.Handle(r.Context(), queries.AuditQueryFilters{
		TenantID:     claims.TenantID,
		ActorID:      q.Get("actor_id"),
		Action:       domain.AuditAction(q.Get("action")),
		ResourceType: domain.ResourceType(q.Get("resource_type")),
		ResourceID:   q.Get("resource_id"),
		Limit:        perPage,
		Offset:       (page - 1) * perPage,
	})
	if err != nil {
		h.fail(w, r, claims.TenantID, err)
		return
	}
	items := make([]entryResponse, len(res.Entries))
	for i, e := range res.Entries {
		items[i] = toResponse(e)
	}
	httpx.WriteSuccess(w, http.StatusOK, map[string]any{
		"items": items, "page": page, "per_page": perPage, "total": res.TotalCount,
	})
}

func (h *HTTPHandler) getEntry(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	entry, err := h.get.Handle(r.Context(), claims.TenantID, r.PathValue("id"))
	if err != nil {
		h.fail(w, r, claims.TenantID, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, toResponse(*entry))
}

// pagination parses ?page and ?per_page: defaults 1 and 20, per_page capped at 100.
func pagination(pageStr, perPageStr string) (page, perPage int, ok bool) {
	page, perPage = 1, defaultPerPage
	var err error
	if pageStr != "" {
		if page, err = strconv.Atoi(pageStr); err != nil || page < 1 {
			return 0, 0, false
		}
	}
	if perPageStr != "" {
		if perPage, err = strconv.Atoi(perPageStr); err != nil || perPage < 1 {
			return 0, 0, false
		}
	}
	return page, min(perPage, maxPerPage), true
}

func (h *HTTPHandler) fail(w http.ResponseWriter, r *http.Request, tenantID string, err error) {
	switch {
	case errors.Is(err, domain.ErrAuditEntryNotFound):
		httpx.WriteError(w, http.StatusNotFound, "audit entry not found")
	case errors.Is(err, domain.ErrInvalidTenantID):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid tenant")
	case errors.Is(err, domain.ErrInvalidEntry):
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		slog.ErrorContext(r.Context(), "audit request failed", "tenant_id", tenantID, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
	}
}
