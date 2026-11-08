package document

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Briyantama/HRIS-Stery/services/document/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/document/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/document/domain"
	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/httpx"
)

const (
	maxBody        = 1 << 20 // the body is only metadata; the file itself goes straight to the bucket
	defaultPerPage = 20
	maxPerPage     = 100
)

type uploader interface {
	Handle(context.Context, commands.UploadDocumentCommand) (*commands.UploadDocumentResult, error)
}
type lister interface {
	Handle(context.Context, queries.ListDocumentsQuery) (*queries.ListDocumentsResult, error)
}
type metadataGetter interface {
	Handle(context.Context, queries.GetDocumentMetadataQuery) (*queries.GetDocumentMetadataResult, error)
}
type urlGetter interface {
	Handle(context.Context, queries.GetDownloadURLQuery) (*queries.DownloadURLResult, error)
}
type deleter interface {
	Handle(context.Context, commands.DeleteDocumentCommand) (*commands.DeleteDocumentResult, error)
}
type rollbacker interface {
	Handle(context.Context, commands.RollbackDocumentVersionCommand) (*commands.CreateDocumentVersionResult, error)
}

// HTTPHandler serves the /api/v1/documents routes.
type HTTPHandler struct {
	upload   uploader
	list     lister
	metadata metadataGetter
	url      urlGetter
	del      deleter
	rollback rollbacker
	verifier authn.Verifier
}

// NewHTTPHandler builds the handler from the command and query handlers.
func NewHTTPHandler(u uploader, l lister, m metadataGetter, g urlGetter, d deleter, r rollbacker, v authn.Verifier) *HTTPHandler {
	return &HTTPHandler{upload: u, list: l, metadata: m, url: g, del: d, rollback: r, verifier: v}
}

// Register adds the document routes to mux. Every route requires a valid bearer token; delete and
// rollback additionally require hr_admin.
func (h *HTTPHandler) Register(mux *http.ServeMux) {
	auth := authn.Authenticate(h.verifier)
	admin := authn.WithRole("hr_admin")
	mux.Handle("POST /api/v1/documents", auth(http.HandlerFunc(h.uploadDocument)))
	mux.Handle("GET /api/v1/documents", auth(http.HandlerFunc(h.listDocuments)))
	mux.Handle("GET /api/v1/documents/{file_id}/download", auth(http.HandlerFunc(h.download)))
	mux.Handle("GET /api/v1/documents/{file_id}/metadata", auth(http.HandlerFunc(h.getMetadata)))
	mux.Handle("POST /api/v1/documents/{file_id}/presigned-url", auth(http.HandlerFunc(h.presign)))
	mux.Handle("DELETE /api/v1/documents/{file_id}", auth(admin(http.HandlerFunc(h.deleteDocument))))
	mux.Handle("POST /api/v1/documents/{file_id}/rollback", auth(admin(http.HandlerFunc(h.rollbackVersion))))
}

// fail maps a use-case error to a status and a fixed message; anything unexpected is logged and
// answered with a generic 500.
func fail(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrDocumentNotFound), errors.Is(err, domain.ErrVersionNotFound), errors.Is(err, domain.ErrInvalidDocumentID):
		httpx.WriteError(w, http.StatusNotFound, "document not found")
	case errors.Is(err, domain.ErrDocumentDeleted):
		httpx.WriteError(w, http.StatusConflict, "document is deleted")
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrInvalidTenantID):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid document data")
	default:
		slog.ErrorContext(r.Context(), "document request failed", "path", r.URL.Path, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
	}
}

type uploadBody struct {
	EntityType     string `json:"entity_type"`
	EntityID       string `json:"entity_id"`
	FileName       string `json:"file_name"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int64  `json:"size_bytes"`
	Classification string `json:"classification"`
	RetentionDays  int32  `json:"retention_days"`
}

func (h *HTTPHandler) uploadDocument(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	var body uploadBody
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	res, err := h.upload.Handle(r.Context(), commands.UploadDocumentCommand{
		TenantID:       claims.TenantID,
		UploadedBy:     claims.UserID,
		EntityType:     body.EntityType,
		EntityID:       body.EntityID,
		FileName:       body.FileName,
		MimeType:       body.MimeType,
		SizeBytes:      body.SizeBytes,
		Classification: body.Classification,
		RetentionDays:  body.RetentionDays,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/documents/"+res.DocumentID)
	httpx.WriteSuccess(w, http.StatusCreated, res)
}

type listData struct {
	Items   []queries.DocumentDTO `json:"items"`
	Page    int                   `json:"page"`
	PerPage int                   `json:"per_page"`
	Total   int32                 `json:"total"`
}

func (h *HTTPHandler) listDocuments(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	q := r.URL.Query()
	page, ok1 := intParam(q.Get("page"), 1)
	perPage, ok2 := intParam(q.Get("per_page"), defaultPerPage)
	if !ok1 || !ok2 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid pagination")
		return
	}
	perPage = min(perPage, maxPerPage)
	res, err := h.list.Handle(r.Context(), queries.ListDocumentsQuery{
		TenantID:   claims.TenantID,
		EntityType: q.Get("entity_type"),
		EntityID:   q.Get("entity_id"),
		Search:     q.Get("search"),
		PageSize:   int32(perPage),
		Offset:     int32((page - 1) * perPage),
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, listData{Items: res.Documents, Page: page, PerPage: perPage, Total: res.TotalCount})
}

// intParam parses an optional integer >= 1; an absent value gives def.
func intParam(s string, def int) (int, bool) {
	if s == "" {
		return def, true
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 1<<20 {
		return 0, false
	}
	return n, true
}

func (h *HTTPHandler) download(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	var version int32
	if s := r.URL.Query().Get("version"); s != "" {
		n, ok := intParam(s, 0)
		if !ok {
			httpx.WriteError(w, http.StatusBadRequest, "invalid version")
			return
		}
		version = int32(n)
	}
	res, err := h.url.Handle(r.Context(), queries.GetDownloadURLQuery{
		TenantID: claims.TenantID, DocumentID: r.PathValue("file_id"), VersionNumber: version,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res)
}

func (h *HTTPHandler) getMetadata(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	include := true
	if s := r.URL.Query().Get("include_versions"); s != "" {
		b, err := strconv.ParseBool(s)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid include_versions")
			return
		}
		include = b
	}
	res, err := h.metadata.Handle(r.Context(), queries.GetDocumentMetadataQuery{
		TenantID: claims.TenantID, DocumentID: r.PathValue("file_id"), IncludeVersions: include,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res)
}

type presignBody struct {
	ExpirationSeconds int32 `json:"expiration_seconds"`
	VersionNumber     int32 `json:"version_number"`
}

func (h *HTTPHandler) presign(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	var body presignBody
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	if body.VersionNumber < 0 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid document data")
		return
	}
	res, err := h.url.Handle(r.Context(), queries.GetDownloadURLQuery{
		TenantID: claims.TenantID, DocumentID: r.PathValue("file_id"), VersionNumber: body.VersionNumber, ExpirySeconds: body.ExpirationSeconds,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res)
}

func (h *HTTPHandler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	res, err := h.del.Handle(r.Context(), commands.DeleteDocumentCommand{
		TenantID: claims.TenantID, DocumentID: r.PathValue("file_id"), DeletedBy: claims.UserID, Reason: r.URL.Query().Get("reason"),
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	slog.InfoContext(r.Context(), "document deleted", "tenant_id", claims.TenantID, "file_id", r.PathValue("file_id"), "actor_id", claims.UserID)
	httpx.WriteSuccess(w, http.StatusOK, res)
}

type rollbackBody struct {
	VersionNumber int32  `json:"version_number"`
	Reason        string `json:"reason"`
}

func (h *HTTPHandler) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	var body rollbackBody
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	if body.VersionNumber < 1 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid document data")
		return
	}
	res, err := h.rollback.Handle(r.Context(), commands.RollbackDocumentVersionCommand{
		TenantID: claims.TenantID, DocumentID: r.PathValue("file_id"), VersionNumber: body.VersionNumber, ActorID: claims.UserID,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	slog.InfoContext(r.Context(), "document rolled back", "tenant_id", claims.TenantID, "file_id", r.PathValue("file_id"),
		"actor_id", claims.UserID, "to_version", body.VersionNumber, "new_version", res.VersionNumber, "reason", body.Reason)
	httpx.WriteSuccess(w, http.StatusOK, res)
}
