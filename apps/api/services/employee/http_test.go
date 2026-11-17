package employee_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/employee"
	"github.com/Briyantama/HRIS-Stery/services/employee/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/employee/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
	"github.com/Briyantama/stery-kit/go/authn"
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
	err       error
	created   commands.CreateEmployeeCommand
	terminate commands.TerminateEmployeeCommand
	list      queries.ListEmployeesQuery
}

type createEmp struct{ s *stub }

func (c createEmp) Handle(_ context.Context, cmd commands.CreateEmployeeCommand) (*commands.CreateEmployeeResult, error) {
	c.s.created = cmd
	return &commands.CreateEmployeeResult{EmployeeID: "emp-9", Email: cmd.Email}, c.s.err
}

type terminateEmp struct{ s *stub }

func (t terminateEmp) Handle(_ context.Context, cmd commands.TerminateEmployeeCommand) (*commands.TerminateEmployeeResult, error) {
	t.s.terminate = cmd
	return &commands.TerminateEmployeeResult{EmployeeID: cmd.EmployeeID, TerminationDate: time.Unix(0, 0)}, t.s.err
}

type createDept struct{ s *stub }

func (c createDept) Handle(context.Context, commands.CreateDepartmentCommand) (*commands.CreateDepartmentResult, error) {
	return &commands.CreateDepartmentResult{DepartmentID: "d-1"}, c.s.err
}

type createPos struct{ s *stub }

func (c createPos) Handle(context.Context, commands.CreatePositionCommand) (*commands.CreatePositionResult, error) {
	return &commands.CreatePositionResult{PositionID: "p-1"}, c.s.err
}

type getEmp struct{ s *stub }

func (g getEmp) Handle(context.Context, queries.GetEmployeeQuery) (*queries.GetEmployeeResult, error) {
	return &queries.GetEmployeeResult{ID: "emp-1"}, g.s.err
}

type listEmp struct{ s *stub }

func (l listEmp) Handle(_ context.Context, q queries.ListEmployeesQuery) (*queries.ListEmployeesResult, error) {
	l.s.list = q
	return &queries.ListEmployeesResult{Employees: []*queries.EmployeeSummary{{ID: "emp-1"}}, Total: 41}, l.s.err
}

type listDept struct{ s *stub }

func (l listDept) Handle(context.Context, queries.ListDepartmentsQuery) (*queries.ListDepartmentsResult, error) {
	return &queries.ListDepartmentsResult{}, l.s.err
}

type listPos struct{ s *stub }

func (l listPos) Handle(context.Context, queries.ListPositionsQuery) (*queries.ListPositionsResult, error) {
	return &queries.ListPositionsResult{}, l.s.err
}

func newServer(s *stub) http.Handler {
	v := verifier{
		"employee": {UserID: "u-1", TenantID: "tenant-claim", Roles: []string{"employee"}},
		"admin":    {UserID: "u-admin", TenantID: "tenant-claim", Roles: []string{"hr_admin"}},
	}
	h := employee.NewHTTPHandler(createEmp{s}, terminateEmp{s}, createDept{s}, createPos{s}, getEmp{s}, listEmp{s}, listDept{s}, listPos{s}, v)
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

const createBody = `{"email":"a@b.c","full_name":"A B","department_id":"d-1","position_id":"p-1"}`

func TestAuthGates(t *testing.T) {
	srv := newServer(&stub{})
	cases := []struct {
		method, path, token, body string
		want                      int
	}{
		{"GET", "/api/v1/employees", "", "", 401},
		{"GET", "/api/v1/employees", "nope", "", 401},
		{"POST", "/api/v1/employees", "employee", createBody, 403},
		{"POST", "/api/v1/employees/emp-1/terminate", "employee", "", 403},
		{"POST", "/api/v1/departments", "employee", `{"name":"x"}`, 403},
		{"POST", "/api/v1/positions", "employee", `{"title":"x"}`, 403},
		{"GET", "/api/v1/departments", "employee", "", 200},
	}
	for _, c := range cases {
		if rec := do(srv, c.method, c.path, c.token, c.body); rec.Code != c.want {
			t.Errorf("%s %s as %q: got %d want %d (%s)", c.method, c.path, c.token, rec.Code, c.want, rec.Body)
		}
	}
}

func TestCreateUsesTokenIdentity(t *testing.T) {
	s := &stub{}
	rec := do(newServer(s), "POST", "/api/v1/employees", "admin", createBody)
	if rec.Code != 201 || rec.Header().Get("Location") != "/api/v1/employees/emp-9" {
		t.Fatalf("got %d %q %s", rec.Code, rec.Header().Get("Location"), rec.Body)
	}
	if s.created.TenantID != "tenant-claim" || s.created.ActorID != "u-admin" {
		t.Fatalf("identity not from token: %+v", s.created)
	}
	if rec := do(newServer(s), "POST", "/api/v1/employees", "admin", `{"tenant_id":"other","email":"a@b.c"}`); rec.Code != 400 {
		t.Fatalf("body tenant_id: got %d", rec.Code)
	}
}

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{domain.ErrNotFound, 404},
		{domain.ErrConflict, 409},
		{domain.ErrInvalid, 422},
		{errors.New("db down"), 500},
	}
	for _, c := range cases {
		rec := do(newServer(&stub{err: c.err}), "GET", "/api/v1/employees/emp-1", "employee", "")
		if rec.Code != c.want || !strings.Contains(rec.Body.String(), `"success":false`) {
			t.Errorf("%v: got %d %s", c.err, rec.Code, rec.Body)
		}
		if c.want == 500 && strings.Contains(rec.Body.String(), "db down") {
			t.Errorf("internal error leaked: %s", rec.Body)
		}
	}
}

func TestListPagination(t *testing.T) {
	s := &stub{}
	rec := do(newServer(s), "GET", "/api/v1/employees?page=3&per_page=10&status=ACTIVE", "employee", "")
	var env struct {
		Success bool `json:"success"`
		Data    struct {
			Items   []json.RawMessage `json:"items"`
			Page    int               `json:"page"`
			PerPage int               `json:"per_page"`
			Total   int               `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil || rec.Code != 200 || !env.Success {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
	if env.Data.Page != 3 || env.Data.PerPage != 10 || env.Data.Total != 41 || len(env.Data.Items) != 1 {
		t.Fatalf("page shape: %+v", env.Data)
	}
	if s.list.Limit != 10 || s.list.Offset != 20 || s.list.Status == nil || *s.list.Status != "ACTIVE" {
		t.Fatalf("query: %+v", s.list)
	}
	for _, q := range []string{"page=0", "page=x", "per_page=0", "per_page=101"} {
		if rec := do(newServer(s), "GET", "/api/v1/employees?"+q, "employee", ""); rec.Code != 422 {
			t.Errorf("%s: got %d", q, rec.Code)
		}
	}
}

func TestTerminateDate(t *testing.T) {
	s := &stub{}
	srv := newServer(s)
	if rec := do(srv, "POST", "/api/v1/employees/emp-1/terminate", "admin", ""); rec.Code != 200 || !s.terminate.TerminationDate.IsZero() {
		t.Fatalf("no body: got %d date %v", rec.Code, s.terminate.TerminationDate)
	}
	if rec := do(srv, "POST", "/api/v1/employees/emp-1/terminate", "admin", `{"termination_date":"2026-12-31"}`); rec.Code != 200 ||
		!s.terminate.TerminationDate.Equal(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("date-only: got %d date %v", rec.Code, s.terminate.TerminationDate)
	}
	if rec := do(srv, "POST", "/api/v1/employees/emp-1/terminate", "admin", `{"termination_date":"31/12/2026"}`); rec.Code != 422 {
		t.Fatalf("bad date: got %d", rec.Code)
	}
}
