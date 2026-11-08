package leave_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Briyantama/HRIS-Stery/services/leave"
	"github.com/Briyantama/HRIS-Stery/services/leave/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/leave/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
	"github.com/Briyantama/stery-kit/go/authn"
)

type verifier map[string]authn.Claims

func (v verifier) Verify(_ context.Context, token string) (authn.Claims, error) {
	if c, ok := v[token]; ok {
		return c, nil
	}
	return authn.Claims{}, errors.New("bad token")
}

// stub implements every handler interface of the leave service.
type stub struct {
	err     error
	gotCmd  commands.CreateLeaveRequestCommand
	ownerID string
}

func (s *stub) createH(ctx context.Context, c commands.CreateLeaveRequestCommand) (*commands.CreateLeaveRequestResult, error) {
	s.gotCmd = c
	return &commands.CreateLeaveRequestResult{LeaveRequestID: "lr-1"}, s.err
}

type creator struct{ s *stub }

func (c creator) Handle(ctx context.Context, cmd commands.CreateLeaveRequestCommand) (*commands.CreateLeaveRequestResult, error) {
	return c.s.createH(ctx, cmd)
}

type approver struct{ s *stub }

func (a approver) Handle(context.Context, commands.ApproveLeaveRequestCommand) (*commands.ApproveLeaveRequestResult, error) {
	return &commands.ApproveLeaveRequestResult{LeaveRequestID: "lr-1"}, a.s.err
}

type rejecter struct{ s *stub }

func (a rejecter) Handle(context.Context, commands.RejectLeaveRequestCommand) (*commands.RejectLeaveRequestResult, error) {
	return &commands.RejectLeaveRequestResult{LeaveRequestID: "lr-1"}, a.s.err
}

type canceller struct{ s *stub }

func (a canceller) Handle(context.Context, commands.CancelLeaveRequestCommand) (*commands.CancelLeaveRequestResult, error) {
	return &commands.CancelLeaveRequestResult{LeaveRequestID: "lr-1"}, a.s.err
}

type getter struct{ s *stub }

func (g getter) Handle(_ context.Context, q queries.GetLeaveRequestQuery) (*queries.LeaveRequestDTO, error) {
	if q.LeaveRequestID == "missing" {
		return nil, domain.ErrNotFound
	}
	return &queries.LeaveRequestDTO{ID: "lr-1", EmployeeID: g.s.ownerID, Status: "PENDING"}, nil
}

type lister struct{ s *stub }

func (l lister) Handle(context.Context, queries.ListLeaveRequestsQuery) (*queries.ListLeaveRequestsResult, error) {
	return &queries.ListLeaveRequestsResult{Records: []*queries.LeaveRequestDTO{{ID: "lr-1"}}, Total: 1}, l.s.err
}

type balancer struct{ s *stub }

func (b balancer) Handle(context.Context, queries.GetLeaveBalanceQuery) (*queries.LeaveBalanceResult, error) {
	return &queries.LeaveBalanceResult{EmployeeID: "emp-1"}, b.s.err
}

type typer struct{ s *stub }

func (t typer) Handle(context.Context, queries.ListLeaveTypesQuery) (*queries.ListLeaveTypesResult, error) {
	return &queries.ListLeaveTypesResult{LeaveTypes: []*queries.LeaveTypeDTO{{ID: "t-1"}}}, t.s.err
}

func newServer(s *stub) http.Handler {
	v := verifier{
		"employee": {UserID: "emp-1", TenantID: "tenant-claim", Roles: []string{"employee"}},
		"manager":  {UserID: "mgr-1", TenantID: "tenant-claim", Roles: []string{"manager"}},
	}
	h := leave.NewHTTPHandler(creator{s}, approver{s}, rejecter{s}, canceller{s}, getter{s}, lister{s}, balancer{s}, typer{s}, v)
	mux := http.NewServeMux()
	h.Register(mux)
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

const validBody = `{"leave_type_id":"t-1","start_date":"2026-12-01","end_date":"2026-12-03","reason":"trip"}`

func TestUnauthenticated401(t *testing.T) {
	rec := do(newServer(&stub{}), "GET", "/api/v1/leaves", "", "")
	if rec.Code != 401 || !strings.Contains(rec.Body.String(), `"success":false`) {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestWrongRole403(t *testing.T) {
	if rec := do(newServer(&stub{}), "POST", "/api/v1/leaves/lr-1/approve", "employee", ""); rec.Code != 403 {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestSuccessEnvelope(t *testing.T) {
	rec := do(newServer(&stub{}), "GET", "/api/v1/leave-types", "employee", "")
	var env struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   *string         `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil || rec.Code != 200 || !env.Success || env.Error != nil || !strings.HasPrefix(string(env.Data), "[") {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestIdentityComesFromToken(t *testing.T) {
	s := &stub{}
	// A tenant_id or employee_id in the body is rejected as an unknown field, so it can never reach the command.
	rec := do(newServer(s), "POST", "/api/v1/leaves", "employee", `{"tenant_id":"evil","leave_type_id":"t-1","start_date":"2026-12-01","end_date":"2026-12-03"}`)
	if rec.Code != 400 || s.gotCmd.TenantID != "" {
		t.Fatalf("body tenant_id not rejected: %d %s", rec.Code, rec.Body)
	}
	rec = do(newServer(s), "POST", "/api/v1/leaves", "employee", validBody)
	if rec.Code != 201 || s.gotCmd.TenantID != "tenant-claim" || s.gotCmd.EmployeeID != "emp-1" || s.gotCmd.DaysCount != 3 {
		t.Fatalf("got %d %+v", rec.Code, s.gotCmd)
	}
	if rec.Header().Get("Location") != "/api/v1/leaves/lr-1" {
		t.Fatalf("location %q", rec.Header().Get("Location"))
	}
}

func TestStatusTable(t *testing.T) {
	tests := []struct {
		name, method, path, token, body string
		err                             error
		owner                           string
		want                            int
	}{
		{"list 200", "GET", "/api/v1/leaves?page=2&per_page=500", "employee", "", nil, "", 200},
		{"apply 201", "POST", "/api/v1/leaves", "employee", validBody, nil, "", 201},
		{"apply malformed 400", "POST", "/api/v1/leaves", "employee", `{`, nil, "", 400},
		{"apply bad dates 422", "POST", "/api/v1/leaves", "employee", `{"leave_type_id":"t-1","start_date":"2026-12-05","end_date":"2026-12-01"}`, nil, "", 422},
		{"apply domain 422", "POST", "/api/v1/leaves", "employee", validBody, domain.ErrInvalidLeaveTypeID, "", 422},
		{"apply no balance 409", "POST", "/api/v1/leaves", "employee", validBody, domain.ErrInsufficientLeaveBalance, "", 409},
		{"apply unknown employee 404", "POST", "/api/v1/leaves", "employee", validBody, domain.ErrNotFound, "", 404},
		{"get missing 404", "GET", "/api/v1/leaves/missing", "employee", "", nil, "emp-1", 404},
		{"get other's 403", "GET", "/api/v1/leaves/lr-1", "employee", "", nil, "emp-9", 403},
		{"get other's as manager 200", "GET", "/api/v1/leaves/lr-1", "manager", "", nil, "emp-9", 200},
		{"approve 200", "POST", "/api/v1/leaves/lr-1/approve", "manager", "", nil, "", 200},
		{"approve twice 409", "POST", "/api/v1/leaves/lr-1/approve", "manager", "", domain.ErrCannotApproveAlreadyRejected, "", 409},
		{"reject 200", "POST", "/api/v1/leaves/lr-1/reject", "manager", `{"reason":"busy"}`, nil, "", 200},
		{"cancel own 200", "POST", "/api/v1/leaves/lr-1/cancel", "employee", "", nil, "emp-1", 200},
		{"cancel other's 403", "POST", "/api/v1/leaves/lr-1/cancel", "employee", "", nil, "emp-9", 403},
		{"balance own 200", "GET", "/api/v1/leave-balance/emp-1", "employee", "", nil, "", 200},
		{"balance other's 403", "GET", "/api/v1/leave-balance/emp-2", "employee", "", nil, "", 403},
		{"types 500", "GET", "/api/v1/leave-types", "employee", "", errors.New("pq: secret detail"), "", 500},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(newServer(&stub{err: tc.err, ownerID: tc.owner}), tc.method, tc.path, tc.token, tc.body)
			if rec.Code != tc.want {
				t.Fatalf("got %d, want %d: %s", rec.Code, tc.want, rec.Body)
			}
			if tc.want == 500 && !strings.Contains(rec.Body.String(), `"internal error"`) {
				t.Fatalf("500 leaked detail: %s", rec.Body)
			}
		})
	}
}

func TestListShape(t *testing.T) {
	rec := do(newServer(&stub{}), "GET", "/api/v1/leaves?per_page=500", "employee", "")
	var env struct {
		Data struct {
			Items   []any `json:"items"`
			Page    int   `json:"page"`
			PerPage int   `json:"per_page"`
			Total   int   `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil || env.Data.PerPage != 100 || env.Data.Page != 1 || env.Data.Total != 1 || len(env.Data.Items) != 1 {
		t.Fatalf("got %s", rec.Body)
	}
}
