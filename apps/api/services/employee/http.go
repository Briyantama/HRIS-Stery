package employee

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/employee/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/employee/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/httpx"
)

const (
	maxBody        = 1 << 20
	defaultPerPage = 20
	maxPerPage     = 100
	roleHRAdmin    = "hr_admin"
)

// Consumer-side views of the use case handlers; the real ones live in application/.
type (
	createEmployeeCommand interface {
		Handle(context.Context, commands.CreateEmployeeCommand) (*commands.CreateEmployeeResult, error)
	}
	terminateEmployeeCommand interface {
		Handle(context.Context, commands.TerminateEmployeeCommand) (*commands.TerminateEmployeeResult, error)
	}
	createDepartmentCommand interface {
		Handle(context.Context, commands.CreateDepartmentCommand) (*commands.CreateDepartmentResult, error)
	}
	createPositionCommand interface {
		Handle(context.Context, commands.CreatePositionCommand) (*commands.CreatePositionResult, error)
	}
	getEmployeeQuery interface {
		Handle(context.Context, queries.GetEmployeeQuery) (*queries.GetEmployeeResult, error)
	}
	listEmployeesQuery interface {
		Handle(context.Context, queries.ListEmployeesQuery) (*queries.ListEmployeesResult, error)
	}
	listDepartmentsQuery interface {
		Handle(context.Context, queries.ListDepartmentsQuery) (*queries.ListDepartmentsResult, error)
	}
	listPositionsQuery interface {
		Handle(context.Context, queries.ListPositionsQuery) (*queries.ListPositionsResult, error)
	}
)

// HTTPHandler serves the employee, department and position routes.
type HTTPHandler struct {
	createEmployee    createEmployeeCommand
	terminateEmployee terminateEmployeeCommand
	createDepartment  createDepartmentCommand
	createPosition    createPositionCommand
	getEmployee       getEmployeeQuery
	listEmployees     listEmployeesQuery
	listDepartments   listDepartmentsQuery
	listPositions     listPositionsQuery
	verifier          authn.Verifier
}

// NewHTTPHandler wires the use cases behind the HTTP routes. verifier authenticates every route.
func NewHTTPHandler(
	createEmployee createEmployeeCommand,
	terminateEmployee terminateEmployeeCommand,
	createDepartment createDepartmentCommand,
	createPosition createPositionCommand,
	getEmployee getEmployeeQuery,
	listEmployees listEmployeesQuery,
	listDepartments listDepartmentsQuery,
	listPositions listPositionsQuery,
	verifier authn.Verifier,
) *HTTPHandler {
	return &HTTPHandler{
		createEmployee:    createEmployee,
		terminateEmployee: terminateEmployee,
		createDepartment:  createDepartment,
		createPosition:    createPosition,
		getEmployee:       getEmployee,
		listEmployees:     listEmployees,
		listDepartments:   listDepartments,
		listPositions:     listPositions,
		verifier:          verifier,
	}
}

// Register adds the routes to mux. Reads need any authenticated user; writes need hr_admin.
func (h *HTTPHandler) Register(mux *http.ServeMux) {
	authed := func(f http.HandlerFunc) http.Handler { return authn.Authenticate(h.verifier)(f) }
	admin := func(f http.HandlerFunc) http.Handler {
		return authn.Authenticate(h.verifier)(authn.WithRole(roleHRAdmin)(f))
	}
	mux.Handle("GET /api/v1/employees", authed(h.handleListEmployees))
	mux.Handle("GET /api/v1/employees/{id}", authed(h.handleGetEmployee))
	mux.Handle("POST /api/v1/employees", admin(h.handleCreateEmployee))
	mux.Handle("POST /api/v1/employees/{id}/terminate", admin(h.handleTerminateEmployee))
	mux.Handle("GET /api/v1/departments", authed(h.handleListDepartments))
	mux.Handle("POST /api/v1/departments", admin(h.handleCreateDepartment))
	mux.Handle("GET /api/v1/positions", authed(h.handleListPositions))
	mux.Handle("POST /api/v1/positions", admin(h.handleCreatePosition))
}

// ---- employees ----

type createEmployeeBody struct {
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	DepartmentID string `json:"department_id"`
	PositionID   string `json:"position_id"`
	ManagerID    string `json:"manager_id"`
}

type employeeCreated struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	DepartmentID string `json:"department_id"`
	PositionID   string `json:"position_id"`
}

func (h *HTTPHandler) handleCreateEmployee(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsOf(w, r)
	if !ok {
		return
	}
	var body createEmployeeBody
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	res, err := h.createEmployee.Handle(r.Context(), commands.CreateEmployeeCommand{
		TenantID:     claims.TenantID,
		Email:        body.Email,
		FullName:     body.FullName,
		Phone:        body.Phone,
		DepartmentID: body.DepartmentID,
		PositionID:   body.PositionID,
		ManagerID:    body.ManagerID,
		ActorID:      claims.UserID,
	})
	if err != nil {
		fail(w, r, err, "employee already exists")
		return
	}
	w.Header().Set("Location", "/api/v1/employees/"+res.EmployeeID)
	httpx.WriteSuccess(w, http.StatusCreated, employeeCreated{
		ID: res.EmployeeID, Email: res.Email, FullName: res.FullName,
		DepartmentID: res.DepartmentID, PositionID: res.PositionID,
	})
}

type employeeDetail struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	FullName        string  `json:"full_name"`
	Phone           string  `json:"phone"`
	DepartmentID    string  `json:"department_id"`
	PositionID      string  `json:"position_id"`
	ManagerID       *string `json:"manager_id"`
	Status          string  `json:"status"`
	ContractType    string  `json:"contract_type"`
	JoinDate        string  `json:"join_date"`
	TerminationDate *string `json:"termination_date"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func (h *HTTPHandler) handleGetEmployee(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsOf(w, r)
	if !ok {
		return
	}
	res, err := h.getEmployee.Handle(r.Context(), queries.GetEmployeeQuery{
		TenantID:   claims.TenantID,
		EmployeeID: r.PathValue("id"),
	})
	if err != nil {
		fail(w, r, err, "")
		return
	}
	d := employeeDetail{
		ID: res.ID, Email: res.Email, FullName: res.FullName, Phone: res.Phone,
		DepartmentID: res.DepartmentID, PositionID: res.PositionID, ManagerID: res.ManagerID,
		Status: res.Status, ContractType: res.ContractType,
		JoinDate: unixToRFC3339(res.JoinDate), CreatedAt: unixToRFC3339(res.CreatedAt), UpdatedAt: unixToRFC3339(res.UpdatedAt),
	}
	if res.TerminationDate != nil {
		t := unixToRFC3339(*res.TerminationDate)
		d.TerminationDate = &t
	}
	httpx.WriteSuccess(w, http.StatusOK, d)
}

type employeeSummary struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	DepartmentID string `json:"department_id"`
	PositionID   string `json:"position_id"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

type page[T any] struct {
	Items   []T `json:"items"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

func (h *HTTPHandler) handleListEmployees(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsOf(w, r)
	if !ok {
		return
	}
	pageNo, perPage, err := httpx.PageParams(r, defaultPerPage, maxPerPage)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid pagination")
		return
	}
	q := queries.ListEmployeesQuery{
		TenantID:   claims.TenantID,
		SearchText: r.URL.Query().Get("search"),
		Limit:      int32(perPage),
		Offset:     int32((pageNo - 1) * perPage),
	}
	if v := r.URL.Query().Get("status"); v != "" {
		q.Status = &v
	}
	if v := r.URL.Query().Get("department_id"); v != "" {
		q.DepartmentID = &v
	}
	res, err := h.listEmployees.Handle(r.Context(), q)
	if err != nil {
		fail(w, r, err, "")
		return
	}
	items := make([]employeeSummary, len(res.Employees))
	for i, e := range res.Employees {
		items[i] = employeeSummary{
			ID: e.ID, Email: e.Email, FullName: e.FullName, DepartmentID: e.DepartmentID,
			PositionID: e.PositionID, Status: e.Status, CreatedAt: unixToRFC3339(e.CreatedAt),
		}
	}
	httpx.WriteSuccess(w, http.StatusOK, page[employeeSummary]{Items: items, Page: pageNo, PerPage: perPage, Total: res.Total})
}

type terminateBody struct {
	TerminationDate string `json:"termination_date"` // RFC 3339 or YYYY-MM-DD; defaults to now
}

type employeeTerminated struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	Status          string `json:"status"`
	TerminationDate string `json:"termination_date"`
}

func (h *HTTPHandler) handleTerminateEmployee(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsOf(w, r)
	if !ok {
		return
	}
	var body terminateBody
	if r.ContentLength != 0 && !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	var when time.Time // zero means now
	if body.TerminationDate != "" {
		t, err := time.Parse(time.RFC3339, body.TerminationDate)
		if err != nil {
			t, err = time.Parse(time.DateOnly, body.TerminationDate)
		}
		if err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid termination_date")
			return
		}
		when = t.UTC()
	}
	res, err := h.terminateEmployee.Handle(r.Context(), commands.TerminateEmployeeCommand{
		TenantID:        claims.TenantID,
		EmployeeID:      r.PathValue("id"),
		TerminationDate: when,
		ActorID:         claims.UserID,
	})
	if err != nil {
		fail(w, r, err, "employee already terminated")
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, employeeTerminated{
		ID: res.EmployeeID, Email: res.Email, Status: string(domain.StatusTerminated),
		TerminationDate: res.TerminationDate.UTC().Format(time.RFC3339),
	})
}

// ---- departments ----

type createDepartmentBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type departmentCreated struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *HTTPHandler) handleCreateDepartment(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsOf(w, r)
	if !ok {
		return
	}
	var body createDepartmentBody
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	res, err := h.createDepartment.Handle(r.Context(), commands.CreateDepartmentCommand{
		TenantID:    claims.TenantID,
		Name:        body.Name,
		Description: body.Description,
		ActorID:     claims.UserID,
	})
	if err != nil {
		fail(w, r, err, "department already exists")
		return
	}
	w.Header().Set("Location", "/api/v1/departments/"+res.DepartmentID)
	httpx.WriteSuccess(w, http.StatusCreated, departmentCreated{ID: res.DepartmentID, Name: res.Name, Description: res.Description})
}

type departmentItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	HeadID      *string `json:"head_id"`
	CreatedAt   string  `json:"created_at"`
}

func (h *HTTPHandler) handleListDepartments(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsOf(w, r)
	if !ok {
		return
	}
	res, err := h.listDepartments.Handle(r.Context(), queries.ListDepartmentsQuery{TenantID: claims.TenantID})
	if err != nil {
		fail(w, r, err, "")
		return
	}
	items := make([]departmentItem, len(res.Departments))
	for i, d := range res.Departments {
		items[i] = departmentItem{ID: d.ID, Name: d.Name, Description: d.Description, HeadID: d.HeadID, CreatedAt: unixToRFC3339(d.CreatedAt)}
	}
	httpx.WriteSuccess(w, http.StatusOK, items)
}

// ---- positions ----

type createPositionBody struct {
	Title string `json:"title"`
	Level string `json:"level"`
}

type positionCreated struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Level string `json:"level"`
}

func (h *HTTPHandler) handleCreatePosition(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsOf(w, r)
	if !ok {
		return
	}
	var body createPositionBody
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	res, err := h.createPosition.Handle(r.Context(), commands.CreatePositionCommand{
		TenantID: claims.TenantID,
		Title:    body.Title,
		Level:    body.Level,
		ActorID:  claims.UserID,
	})
	if err != nil {
		fail(w, r, err, "position already exists")
		return
	}
	w.Header().Set("Location", "/api/v1/positions/"+res.PositionID)
	httpx.WriteSuccess(w, http.StatusCreated, positionCreated{ID: res.PositionID, Title: res.Title, Level: res.Level})
}

type positionItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Level       string `json:"level"`
	CreatedAt   string `json:"created_at"`
}

func (h *HTTPHandler) handleListPositions(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsOf(w, r)
	if !ok {
		return
	}
	res, err := h.listPositions.Handle(r.Context(), queries.ListPositionsQuery{TenantID: claims.TenantID})
	if err != nil {
		fail(w, r, err, "")
		return
	}
	items := make([]positionItem, len(res.Positions))
	for i, p := range res.Positions {
		items[i] = positionItem{ID: p.ID, Title: p.Title, Description: p.Description, Level: p.Level, CreatedAt: unixToRFC3339(p.CreatedAt)}
	}
	httpx.WriteSuccess(w, http.StatusOK, items)
}

// ---- helpers ----

// claimsOf returns the caller identity; every route sits behind authn.Authenticate, so a missing
// claim is a wiring bug and answers 401 rather than acting without a tenant.
func claimsOf(w http.ResponseWriter, r *http.Request) (authn.Claims, bool) {
	claims, ok := authn.FromContext(r.Context())
	if !ok || claims.TenantID == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return authn.Claims{}, false
	}
	return claims, true
}

// fail maps a use case error to a status. conflictMsg is the fixed 409 text for the route.
func fail(w http.ResponseWriter, r *http.Request, err error, conflictMsg string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not found")
	case errors.Is(err, domain.ErrConflict):
		httpx.WriteError(w, http.StatusConflict, conflictMsg)
	case errors.Is(err, domain.ErrInvalid):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "validation failed")
	default:
		slog.ErrorContext(r.Context(), "employee request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
	}
}

func unixToRFC3339(sec int64) string { return time.Unix(sec, 0).UTC().Format(time.RFC3339) }
