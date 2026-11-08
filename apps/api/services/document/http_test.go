package document_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Briyantama/HRIS-Stery/services/document"
	"github.com/Briyantama/HRIS-Stery/services/document/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/document/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/document/domain"
	"github.com/Briyantama/stery-kit/go/authn"
)

type verifier map[string]authn.Claims

func (v verifier) Verify(_ context.Context, token string) (authn.Claims, error) {
	if c, ok := v[token]; ok {
		return c, nil
	}
	return authn.Claims{}, errors.New("bad token")
}

// stub implements every handler interface of the document service; err is returned by all of them.
type stub struct {
	err    error
	gotUp  commands.UploadDocumentCommand
	gotURL queries.GetDownloadURLQuery
}

type uploadH struct{ s *stub }
type listH struct{ s *stub }
type metaH struct{ s *stub }
type urlH struct{ s *stub }
type delH struct{ s *stub }
type rollH struct{ s *stub }

func (h uploadH) Handle(_ context.Context, c commands.UploadDocumentCommand) (*commands.UploadDocumentResult, error) {
	h.s.gotUp = c
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &commands.UploadDocumentResult{DocumentID: "f-1"}, nil
}
func (h listH) Handle(context.Context, queries.ListDocumentsQuery) (*queries.ListDocumentsResult, error) {
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &queries.ListDocumentsResult{Documents: []queries.DocumentDTO{{FileID: "f-1"}}, TotalCount: 1}, nil
}
func (h metaH) Handle(context.Context, queries.GetDocumentMetadataQuery) (*queries.GetDocumentMetadataResult, error) {
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &queries.GetDocumentMetadataResult{Document: queries.DocumentDTO{FileID: "f-1"}}, nil
}
func (h urlH) Handle(_ context.Context, q queries.GetDownloadURLQuery) (*queries.DownloadURLResult, error) {
	h.s.gotURL = q
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &queries.DownloadURLResult{URL: "https://example.test/f"}, nil
}
func (h delH) Handle(context.Context, commands.DeleteDocumentCommand) (*commands.DeleteDocumentResult, error) {
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &commands.DeleteDocumentResult{Success: true}, nil
}
func (h rollH) Handle(context.Context, commands.RollbackDocumentVersionCommand) (*commands.CreateDocumentVersionResult, error) {
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &commands.CreateDocumentVersionResult{VersionNumber: 3}, nil
}

const (
	userTok  = "user"
	adminTok = "admin"
)

func newMux(s *stub) *http.ServeMux {
	v := verifier{
		userTok:  {UserID: "u-1", TenantID: "tenant-token", Roles: []string{"employee"}},
		adminTok: {UserID: "a-1", TenantID: "tenant-token", Roles: []string{"hr_admin"}},
	}
	mux := http.NewServeMux()
	document.NewHTTPHandler(uploadH{s}, listH{s}, metaH{s}, urlH{s}, delH{s}, rollH{s}, v).Register(mux)
	return mux
}

func do(mux *http.ServeMux, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestAuthAndRoles(t *testing.T) {
	mux := newMux(&stub{})
	if rec := do(mux, "GET", "/api/v1/documents?search=x", "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: %d", rec.Code)
	}
	if rec := do(mux, "DELETE", "/api/v1/documents/f-1", userTok, ""); rec.Code != http.StatusForbidden {
		t.Fatalf("delete as employee: %d", rec.Code)
	}
	if rec := do(mux, "POST", "/api/v1/documents/f-1/rollback", userTok, `{"version_number":1}`); rec.Code != http.StatusForbidden {
		t.Fatalf("rollback as employee: %d", rec.Code)
	}
}

func TestEnvelopeAndPagination(t *testing.T) {
	rec := do(newMux(&stub{}), "GET", "/api/v1/documents?search=x&per_page=500", userTok, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var env struct {
		Success bool
		Data    struct {
			Items   []map[string]any
			Page    int
			PerPage int `json:"per_page"`
			Total   int
		}
		Error *string
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.Success || env.Error != nil || len(env.Data.Items) != 1 || env.Data.Page != 1 || env.Data.PerPage != 100 || env.Data.Total != 1 {
		t.Fatalf("bad envelope: %s", rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing Cache-Control: no-store")
	}
}

func TestTenantAndActorComeFromToken(t *testing.T) {
	s := &stub{}
	body := `{"entity_type":"EMPLOYEE","entity_id":"e","file_name":"a.pdf","mime_type":"application/pdf","size_bytes":1,"classification":"PUBLIC"}`
	rec := do(newMux(s), "POST", "/api/v1/documents", userTok, body)
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/v1/documents/f-1" {
		t.Fatalf("upload: %d %q", rec.Code, rec.Header().Get("Location"))
	}
	if s.gotUp.TenantID != "tenant-token" || s.gotUp.UploadedBy != "u-1" {
		t.Fatalf("ids not from token: %+v", s.gotUp)
	}
	// A body-supplied tenant_id is an unknown field: rejected, never trusted.
	rec = do(newMux(s), "POST", "/api/v1/documents", userTok, `{"tenant_id":"evil","entity_type":"EMPLOYEE"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("body tenant_id: %d", rec.Code)
	}
}

func TestStatusCodes(t *testing.T) {
	upload := `{"entity_type":"EMPLOYEE","entity_id":"e","file_name":"a.pdf","mime_type":"application/pdf","size_bytes":1,"classification":"PUBLIC"}`
	rows := []struct {
		name, method, path, token, body string
		err                             error
		want                            int
	}{
		{"create", "POST", "/api/v1/documents", userTok, upload, nil, 201},
		{"list", "GET", "/api/v1/documents?search=x", userTok, "", nil, 200},
		{"metadata", "GET", "/api/v1/documents/f-1/metadata", userTok, "", nil, 200},
		{"download", "GET", "/api/v1/documents/f-1/download", userTok, "", nil, 200},
		{"presigned", "POST", "/api/v1/documents/f-1/presigned-url", userTok, `{"expiration_seconds":60}`, nil, 200},
		{"delete", "DELETE", "/api/v1/documents/f-1", adminTok, "", nil, 200},
		{"rollback", "POST", "/api/v1/documents/f-1/rollback", adminTok, `{"version_number":1}`, nil, 200},
		{"malformed json", "POST", "/api/v1/documents", userTok, `{`, nil, 400},
		{"bad page", "GET", "/api/v1/documents?page=0", userTok, "", nil, 400},
		{"rollback without version", "POST", "/api/v1/documents/f-1/rollback", adminTok, `{}`, nil, 422},
		{"invalid input", "POST", "/api/v1/documents", userTok, upload, domain.ErrInvalidInput, 422},
		{"not found", "GET", "/api/v1/documents/f-1/metadata", userTok, "", domain.ErrDocumentNotFound, 404},
		{"version not found", "POST", "/api/v1/documents/f-1/rollback", adminTok, `{"version_number":9}`, domain.ErrVersionNotFound, 404},
		{"already deleted", "DELETE", "/api/v1/documents/f-1", adminTok, "", domain.ErrDocumentDeleted, 409},
		{"unexpected", "GET", "/api/v1/documents/f-1/download", userTok, "", errors.New("pq: secret detail"), 500},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			rec := do(newMux(&stub{err: r.err}), r.method, r.path, r.token, r.body)
			if rec.Code != r.want {
				t.Fatalf("got %d, want %d: %s", rec.Code, r.want, rec.Body.String())
			}
			if r.want == 500 && !strings.Contains(rec.Body.String(), `"internal error"`) {
				t.Fatalf("500 leaked detail: %s", rec.Body.String())
			}
		})
	}
}

func TestDownloadCarriesViewer(t *testing.T) {
	for tok, want := range map[string]domain.Viewer{
		userTok:  {UserID: "u-1"},
		adminTok: {UserID: "a-1", HR: true},
	} {
		s := &stub{}
		if rec := do(newMux(s), "GET", "/api/v1/documents/f-1/download", tok, ""); rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", tok, rec.Code)
		}
		if s.gotURL.Viewer != want {
			t.Errorf("%s: viewer %+v, want %+v", tok, s.gotURL.Viewer, want)
		}
	}
}

func TestDownloadForbiddenDocumentIsNotFound(t *testing.T) {
	rec := do(newMux(&stub{err: domain.ErrDocumentNotFound}), "GET", "/api/v1/documents/f-1/download", userTok, "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}
