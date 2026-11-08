package audit_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Briyantama/stery-kit/go/authn"

	"github.com/Briyantama/HRIS-Stery/services/audit"
	"github.com/Briyantama/HRIS-Stery/services/audit/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/audit/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/audit/domain"
)

const (
	tenant  = "550e8400-e29b-41d4-a716-446655440000"
	entryID = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
)

type stubVerifier struct{}

func (stubVerifier) Verify(_ context.Context, token string) (authn.Claims, error) {
	switch token {
	case "admin":
		return authn.Claims{UserID: "u-1", TenantID: tenant, Roles: []string{"hr_admin"}}, nil
	case "employee":
		return authn.Claims{UserID: "u-2", TenantID: tenant, Roles: []string{"employee"}}, nil
	}
	return authn.Claims{}, errors.New("bad token")
}

type stub struct {
	err       error
	gotRecord commands.RecordAuditCommand
	gotFilter queries.AuditQueryFilters
	gotTenant string
	recorded  bool
}

func (s *stub) dto() *queries.AuditEntryDTO {
	return &queries.AuditEntryDTO{ID: entryID, TenantID: tenant, ActorID: "u-1", Action: "X", ResourceType: "Y", CreatedAt: time.Now()}
}

type recordStub struct{ *stub }
type queryStub struct{ *stub }
type getStub struct{ *stub }

func (s recordStub) Handle(_ context.Context, c commands.RecordAuditCommand) (*commands.RecordAuditResult, error) {
	s.recorded, s.stub.gotRecord = true, c
	return &commands.RecordAuditResult{EntryID: entryID}, s.err
}
func (s queryStub) Handle(_ context.Context, f queries.AuditQueryFilters) (*queries.QueryAuditTrailResult, error) {
	s.gotFilter = f
	if s.err != nil {
		return nil, s.err
	}
	return &queries.QueryAuditTrailResult{Entries: []queries.AuditEntryDTO{*s.dto()}, TotalCount: 41}, nil
}
func (s getStub) Handle(_ context.Context, tenantID, _ string) (*queries.AuditEntryDTO, error) {
	s.gotTenant = tenantID
	if s.err != nil {
		return nil, s.err
	}
	return s.dto(), nil
}

func serve(s *stub, method, path, token, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	audit.NewHTTPHandler(recordStub{s}, queryStub{s}, getStub{s}, stubVerifier{}).Register(mux)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestAuditHTTP(t *testing.T) {
	const okBody = `{"action":"LOGIN","resource_type":"USER","resource_id":"` + entryID + `","success":true}`
	tests := []struct {
		name, method, path, token, body string
		err                             error
		want                            int
	}{
		{"401 no token", "GET", "/api/v1/audit/entries", "", "", nil, 401},
		{"403 wrong role", "GET", "/api/v1/audit/entries", "employee", "", nil, 403},
		{"200 list", "GET", "/api/v1/audit/entries?per_page=500", "admin", "", nil, 200},
		{"200 get", "GET", "/api/v1/audit/entries/" + entryID, "admin", "", nil, 200},
		{"201 record", "POST", "/api/v1/audit/entries", "admin", okBody, nil, 201},
		{"400 bad json", "POST", "/api/v1/audit/entries", "admin", `{`, nil, 400},
		{"400 body tenant_id", "POST", "/api/v1/audit/entries", "admin", `{"tenant_id":"x","action":"A","resource_type":"B"}`, nil, 400},
		{"400 bad page", "GET", "/api/v1/audit/entries?page=0", "admin", "", nil, 400},
		{"422 missing action", "POST", "/api/v1/audit/entries", "admin", `{"resource_type":"B"}`, nil, 422},
		{"422 bad resource_id", "POST", "/api/v1/audit/entries", "admin", `{"action":"A","resource_type":"B","resource_id":"nope"}`, nil, 422},
		{"404 missing", "GET", "/api/v1/audit/entries/" + entryID, "admin", "", domain.ErrAuditEntryNotFound, 404},
		{"500 hides cause", "GET", "/api/v1/audit/entries", "admin", "", errors.New("pq: secret"), 500},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(&stub{err: tc.err}, tc.method, tc.path, tc.token, tc.body)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body)
			}
			if tc.want == 500 && strings.Contains(rec.Body.String(), "secret") {
				t.Fatalf("500 leaks cause: %s", rec.Body)
			}
		})
	}
}

func TestAuditHTTP_EnvelopeAndClaims(t *testing.T) {
	s := &stub{}
	rec := serve(s, "GET", "/api/v1/audit/entries?actor_id=a&page=3&per_page=10", "admin", "")
	var env struct {
		Success bool `json:"success"`
		Data    struct {
			Items   []map[string]any `json:"items"`
			Page    int              `json:"page"`
			PerPage int              `json:"per_page"`
			Total   int              `json:"total"`
		} `json:"data"`
		Error *string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.Success || env.Error != nil || len(env.Data.Items) != 1 || env.Data.Page != 3 || env.Data.PerPage != 10 || env.Data.Total != 41 {
		t.Fatalf("unexpected envelope: %s", rec.Body)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("GET must set Cache-Control: no-store")
	}

	// The recorder must see the tenant and actor of the token.
	rec = serve(s, "POST", "/api/v1/audit/entries", "admin", `{"action":"A","resource_type":"B"}`)
	if rec.Code != 201 || rec.Header().Get("Location") != "/api/v1/audit/entries/"+entryID {
		t.Fatalf("record: %d %v", rec.Code, rec.Header())
	}
	if s.gotRecord.TenantID != tenant || s.gotRecord.ActorID != "u-1" {
		t.Fatalf("tenant/actor not taken from claims: %+v", s.gotRecord)
	}
}
