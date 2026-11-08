package attendance_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Briyantama/stery-kit/go/authn"

	"github.com/Briyantama/HRIS-Stery/services/attendance"
	"github.com/Briyantama/HRIS-Stery/services/attendance/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/attendance/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/attendance/domain"
)

const (
	tenant = "550e8400-e29b-41d4-a716-446655440000"
	userID = "660e8400-e29b-41d4-a716-446655440000"
	other  = "770e8400-e29b-41d4-a716-446655440000"
)

type verifier map[string]authn.Claims

func (v verifier) Verify(_ context.Context, token string) (authn.Claims, error) {
	if c, ok := v[token]; ok {
		return c, nil
	}
	return authn.Claims{}, errors.New("bad token")
}

// stub implements every handler interface; err is returned by all of them.
type stub struct {
	err      error
	checkIn  commands.CheckInCommand
	checkOut commands.CheckOutCommand
	override commands.OverrideAttendanceCommand
	list     queries.ListAttendanceQuery
	summary  queries.GetDailySummaryQuery
}

type ci struct{ s *stub }

func (c ci) Handle(_ context.Context, cmd commands.CheckInCommand) (*commands.CheckInResult, error) {
	c.s.checkIn = cmd
	return &commands.CheckInResult{AttendanceID: "a1", EmployeeID: cmd.EmployeeID, TenantID: cmd.TenantID}, c.s.err
}

type co struct{ s *stub }

func (c co) Handle(_ context.Context, cmd commands.CheckOutCommand) (*commands.CheckOutResult, error) {
	c.s.checkOut = cmd
	return &commands.CheckOutResult{AttendanceID: "a1"}, c.s.err
}

type ov struct{ s *stub }

func (c ov) Handle(_ context.Context, cmd commands.OverrideAttendanceCommand) (*commands.OverrideAttendanceResult, error) {
	c.s.override = cmd
	return &commands.OverrideAttendanceResult{AttendanceID: "a1", TenantID: cmd.TenantID}, c.s.err
}

type ls struct{ s *stub }

func (c ls) Handle(_ context.Context, q queries.ListAttendanceQuery) (*queries.ListAttendanceResult, error) {
	c.s.list = q
	return &queries.ListAttendanceResult{Records: []*queries.AttendanceDTO{{ID: "a1"}}, Total: 41}, c.s.err
}

type sm struct{ s *stub }

func (c sm) Handle(_ context.Context, q queries.GetDailySummaryQuery) (*queries.DailySummaryResult, error) {
	c.s.summary = q
	return &queries.DailySummaryResult{TenantID: q.TenantID}, c.s.err
}

func setup(s *stub) http.Handler {
	v := verifier{
		"admin":    {UserID: userID, TenantID: tenant, Roles: []string{"hr_admin"}},
		"manager":  {UserID: userID, TenantID: tenant, Roles: []string{"manager"}},
		"employee": {UserID: userID, TenantID: tenant, Roles: []string{"employee"}},
	}
	mux := http.NewServeMux()
	attendance.NewHTTPHandler(ci{s}, co{s}, ov{s}, ls{s}, sm{s}, v).Register(mux)
	return mux
}

func do(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if body == "" {
		r.ContentLength = 0
	}
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *string         `json:"error"`
}

func decode(t *testing.T, w *httptest.ResponseRecorder) envelope {
	t.Helper()
	var e envelope
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
		t.Fatalf("body %q is not an envelope: %v", w.Body.String(), err)
	}
	return e
}

func TestAuthRequiredAndRoles(t *testing.T) {
	h := setup(&stub{})
	routes := []struct{ method, path, body, role string }{
		{"POST", "/api/v1/attendance/check-in", "", ""},
		{"POST", "/api/v1/attendance/check-out", "", ""},
		{"GET", "/api/v1/attendance/summary/" + userID, "", ""},
		{"GET", "/api/v1/attendance", "", "manager"},
		{"POST", "/api/v1/attendance/override", `{}`, "hr_admin"},
	}
	for _, rt := range routes {
		if w := do(h, rt.method, rt.path, "", rt.body); w.Code != 401 {
			t.Errorf("%s %s without token: %d, want 401", rt.method, rt.path, w.Code)
		}
		if rt.role != "" {
			if w := do(h, rt.method, rt.path, "employee", rt.body); w.Code != 403 {
				t.Errorf("%s %s as employee: %d, want 403", rt.method, rt.path, w.Code)
			}
		}
	}
	if w := do(h, "POST", "/api/v1/attendance/override", "manager", `{}`); w.Code != 403 {
		t.Errorf("override as manager: %d, want 403", w.Code)
	}
	if w := do(h, "GET", "/api/v1/attendance", "manager", ""); w.Code != 200 {
		t.Errorf("list as manager: %d, want 200", w.Code)
	}
}

func TestSuccessEnvelopeAndClaimWins(t *testing.T) {
	s := &stub{}
	h := setup(s)
	w := do(h, "POST", "/api/v1/attendance/check-in", "employee", `{"employee_id":"`+other+`","notes":"x"}`)
	if w.Code != 200 {
		t.Fatalf("check-in: %d %s", w.Code, w.Body)
	}
	if e := decode(t, w); !e.Success || e.Error != nil || len(e.Data) == 0 {
		t.Errorf("bad envelope: %s", w.Body)
	}
	if s.checkIn.TenantID != tenant || s.checkIn.ActorID != userID || s.checkIn.EmployeeID != userID {
		t.Errorf("claims must win for a non-admin: %+v", s.checkIn)
	}

	// a body tenant_id is an unknown field: rejected, never used
	if w := do(h, "POST", "/api/v1/attendance/check-in", "employee", `{"tenant_id":"`+other+`"}`); w.Code != 400 {
		t.Errorf("body tenant_id: %d, want 400", w.Code)
	}

	// hr_admin may punch for another employee, still in its own tenant
	do(h, "POST", "/api/v1/attendance/check-in", "admin", `{"employee_id":"`+other+`"}`)
	if s.checkIn.EmployeeID != other || s.checkIn.TenantID != tenant {
		t.Errorf("hr_admin punch: %+v", s.checkIn)
	}

	// empty body is fine for check-in
	if w := do(h, "POST", "/api/v1/attendance/check-in", "employee", ""); w.Code != 200 {
		t.Errorf("check-in with no body: %d", w.Code)
	}
}

func TestListPagination(t *testing.T) {
	s := &stub{}
	h := setup(s)
	w := do(h, "GET", "/api/v1/attendance?page=3&per_page=10&status=LATE", "manager", "")
	if w.Code != 200 {
		t.Fatalf("list: %d %s", w.Code, w.Body)
	}
	var page struct {
		Items   []json.RawMessage `json:"items"`
		Page    int               `json:"page"`
		PerPage int               `json:"per_page"`
		Total   int               `json:"total"`
	}
	if err := json.Unmarshal(decode(t, w).Data, &page); err != nil {
		t.Fatal(err)
	}
	if page.Page != 3 || page.PerPage != 10 || page.Total != 41 || len(page.Items) != 1 {
		t.Errorf("page = %+v", page)
	}
	if s.list.Offset != 20 || s.list.Limit != 10 || *s.list.Status != "LATE" || s.list.TenantID != tenant {
		t.Errorf("query = %+v", s.list)
	}
	for _, q := range []string{"page=0", "per_page=101", "page=x", "date_from=nope"} {
		if w := do(h, "GET", "/api/v1/attendance?"+q, "manager", ""); w.Code != 422 {
			t.Errorf("?%s: %d, want 422", q, w.Code)
		}
	}
}

func TestStatusCodes(t *testing.T) {
	override := `{"employee_id":"` + userID + `","date":"2026-11-08","status":"PRESENT"}`
	tests := []struct {
		name         string
		err          error
		method, path string
		token, body  string
		want         int
		wantMsg      string
	}{
		{"ok", nil, "POST", "/api/v1/attendance/check-out", "employee", "", 200, ""},
		{"malformed json", nil, "POST", "/api/v1/attendance/check-in", "employee", `{`, 400, ""},
		{"not found", domain.ErrNotFound, "POST", "/api/v1/attendance/check-out", "employee", "", 404, "attendance record not found"},
		{"check in twice", domain.ErrCannotCheckInTwiceSameDay, "POST", "/api/v1/attendance/check-in", "employee", "", 409, ""},
		{"already checked out", domain.ErrAlreadyCheckedOut, "POST", "/api/v1/attendance/check-out", "employee", "", 409, ""},
		{"invalid status", domain.ErrInvalidStatus, "POST", "/api/v1/attendance/override", "admin", override, 422, "invalid status"},
		{"invalid time range", domain.ErrInvalidTimeRange, "POST", "/api/v1/attendance/override", "admin", override, 422, ""},
		{"bad date", nil, "POST", "/api/v1/attendance/override", "admin", `{"date":"x"}`, 422, ""},
		{"bad location", nil, "POST", "/api/v1/attendance/check-in", "employee", `{"latitude":123}`, 422, ""},
		{"bad summary id", nil, "GET", "/api/v1/attendance/summary/not-a-uuid", "employee", "", 422, ""},
		{"internal error hides detail", errors.New("pq: secret detail"), "GET", "/api/v1/attendance/summary/" + userID, "employee", "", 500, "internal error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := do(setup(&stub{err: tt.err}), tt.method, tt.path, tt.token, tt.body)
			if w.Code != tt.want {
				t.Fatalf("status %d, want %d: %s", w.Code, tt.want, w.Body)
			}
			e := decode(t, w)
			if (tt.want == 200) != e.Success {
				t.Errorf("success = %v for %d", e.Success, w.Code)
			}
			if tt.wantMsg != "" && (e.Error == nil || *e.Error != tt.wantMsg) {
				t.Errorf("error = %v, want %q", e.Error, tt.wantMsg)
			}
		})
	}
}
