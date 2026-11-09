package attendance

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/httpx"
	"uuid"

	"github.com/Briyantama/HRIS-Stery/services/attendance/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/attendance/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/attendance/domain"
)

const (
	maxBody        = 1 << 20
	defaultPerPage = 20
	maxPerPage     = 100
	roleHRAdmin    = "hr_admin"
	roleManager    = "manager"
)

// Consumer-side views of the command and query handlers.
type (
	checkInHandler interface {
		Handle(context.Context, commands.CheckInCommand) (*commands.CheckInResult, error)
	}
	checkOutHandler interface {
		Handle(context.Context, commands.CheckOutCommand) (*commands.CheckOutResult, error)
	}
	overrideHandler interface {
		Handle(context.Context, commands.OverrideAttendanceCommand) (*commands.OverrideAttendanceResult, error)
	}
	listHandler interface {
		Handle(context.Context, queries.ListAttendanceQuery) (*queries.ListAttendanceResult, error)
	}
	summaryHandler interface {
		Handle(context.Context, queries.GetDailySummaryQuery) (*queries.DailySummaryResult, error)
	}
)

// HTTPHandler serves the attendance routes.
type HTTPHandler struct {
	checkIn  checkInHandler
	checkOut checkOutHandler
	override overrideHandler
	list     listHandler
	summary  summaryHandler
	verifier authn.Verifier
}

// NewHTTPHandler creates the attendance HTTP handler.
func NewHTTPHandler(ci checkInHandler, co checkOutHandler, ov overrideHandler, ls listHandler, sm summaryHandler, v authn.Verifier) *HTTPHandler {
	return &HTTPHandler{checkIn: ci, checkOut: co, override: ov, list: ls, summary: sm, verifier: v}
}

// Register mounts the routes. All of them require a valid token.
func (h *HTTPHandler) Register(mux *http.ServeMux) {
	auth := authn.Authenticate(h.verifier)
	managers := authn.WithRole(roleManager, roleHRAdmin)
	admins := authn.WithRole(roleHRAdmin)
	mux.Handle("POST /api/v1/attendance/check-in", auth(http.HandlerFunc(h.handleCheckIn)))
	mux.Handle("POST /api/v1/attendance/check-out", auth(http.HandlerFunc(h.handleCheckOut)))
	mux.Handle("GET /api/v1/attendance/summary/{employee_id}", auth(managers(http.HandlerFunc(h.handleSummary)))) // tenant-wide counts
	mux.Handle("GET /api/v1/attendance", auth(managers(http.HandlerFunc(h.handleList))))
	mux.Handle("POST /api/v1/attendance/override", auth(admins(http.HandlerFunc(h.handleOverride))))
}

type punchBody struct {
	EmployeeID string   `json:"employee_id"` // honored for hr_admin only
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	Notes      string   `json:"notes"`
}

// readPunch decodes an optional body: check-in and check-out work with no body at all.
func readPunch(w http.ResponseWriter, r *http.Request) (punchBody, bool) {
	var b punchBody
	if r.ContentLength == 0 {
		return b, true
	}
	if !httpx.DecodeJSON(w, r, &b, maxBody) {
		return b, false
	}
	if !validLocation(b.Latitude, -90, 90) || !validLocation(b.Longitude, -180, 180) {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "location is out of range")
		return b, false
	}
	return b, true
}

func validLocation(v *float64, lo, hi float64) bool { return v == nil || (*v >= lo && *v <= hi) }

// employeeFor returns the employee the punch is for: the caller, or the body's employee for hr_admin.
func employeeFor(c authn.Claims, bodyEmployee string) string {
	if bodyEmployee != "" {
		for _, role := range c.Roles {
			if strings.EqualFold(role, roleHRAdmin) {
				return bodyEmployee
			}
		}
	}
	return c.UserID
}

// today is the current UTC calendar day at midnight, the value stored in the date column.
func today(now time.Time) time.Time {
	y, m, d := now.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (h *HTTPHandler) handleCheckIn(w http.ResponseWriter, r *http.Request) {
	c, _ := authn.FromContext(r.Context())
	b, ok := readPunch(w, r)
	if !ok {
		return
	}
	now := time.Now().UTC()
	res, err := h.checkIn.Handle(r.Context(), commands.CheckInCommand{
		TenantID: c.TenantID, EmployeeID: employeeFor(c, b.EmployeeID), Date: today(now), CheckInAt: now,
		Latitude: b.Latitude, Longitude: b.Longitude, Notes: b.Notes, ActorID: c.UserID,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res)
}

func (h *HTTPHandler) handleCheckOut(w http.ResponseWriter, r *http.Request) {
	c, _ := authn.FromContext(r.Context())
	b, ok := readPunch(w, r)
	if !ok {
		return
	}
	now := time.Now().UTC()
	res, err := h.checkOut.Handle(r.Context(), commands.CheckOutCommand{
		TenantID: c.TenantID, EmployeeID: employeeFor(c, b.EmployeeID), Date: today(now), CheckOutAt: now,
		Latitude: b.Latitude, Longitude: b.Longitude, Notes: b.Notes, ActorID: c.UserID,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res)
}

type overrideBody struct {
	EmployeeID string     `json:"employee_id"`
	Date       string     `json:"date"` // YYYY-MM-DD
	CheckInAt  *time.Time `json:"check_in_at"`
	CheckOutAt *time.Time `json:"check_out_at"`
	Status     string     `json:"status"`
	Reason     string     `json:"reason"`
}

func (h *HTTPHandler) handleOverride(w http.ResponseWriter, r *http.Request) {
	c, _ := authn.FromContext(r.Context())
	var b overrideBody
	if !httpx.DecodeJSON(w, r, &b, maxBody) {
		return
	}
	date, err := time.Parse(time.DateOnly, b.Date)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "date must be YYYY-MM-DD")
		return
	}
	res, err := h.override.Handle(r.Context(), commands.OverrideAttendanceCommand{
		TenantID: c.TenantID, EmployeeID: b.EmployeeID, Date: date, CheckInAt: b.CheckInAt, CheckOutAt: b.CheckOutAt,
		Status: b.Status, Reason: b.Reason, ActorID: c.UserID,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res)
}

// handleSummary returns the tenant's attendance summary for ?date= (default today). The employee id
// in the path is validated only: the former service summarized the whole tenant day too.
func (h *HTTPHandler) handleSummary(w http.ResponseWriter, r *http.Request) {
	c, _ := authn.FromContext(r.Context())
	if _, err := uuid.Parse(r.PathValue("employee_id")); err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid employee id")
		return
	}
	date := today(time.Now())
	if s := r.URL.Query().Get("date"); s != "" {
		d, err := time.Parse(time.DateOnly, s)
		if err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "date must be YYYY-MM-DD")
			return
		}
		date = d
	}
	res, err := h.summary.Handle(r.Context(), queries.GetDailySummaryQuery{TenantID: c.TenantID, Date: date})
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res)
}

type listPage struct {
	Items   []*queries.AttendanceDTO `json:"items"`
	Page    int                      `json:"page"`
	PerPage int                      `json:"per_page"`
	Total   int                      `json:"total"`
}

func (h *HTTPHandler) handleList(w http.ResponseWriter, r *http.Request) {
	c, _ := authn.FromContext(r.Context())
	q := r.URL.Query()
	page, perPage, ok := pageParams(w, q.Get("page"), q.Get("per_page"))
	if !ok {
		return
	}
	query := queries.ListAttendanceQuery{TenantID: c.TenantID, Offset: (page - 1) * perPage, Limit: perPage}
	if s := q.Get("employee_id"); s != "" {
		query.EmployeeID = &s
	}
	if s := q.Get("status"); s != "" {
		query.Status = &s
	}
	for name, dst := range map[string]**time.Time{"date_from": &query.DateFrom, "date_to": &query.DateTo} {
		if s := q.Get(name); s != "" {
			d, err := time.Parse(time.DateOnly, s)
			if err != nil {
				httpx.WriteError(w, http.StatusUnprocessableEntity, name+" must be YYYY-MM-DD")
				return
			}
			*dst = &d
		}
	}
	res, err := h.list.Handle(r.Context(), query)
	if err != nil {
		writeError(w, r, err)
		return
	}
	items := res.Records
	if items == nil {
		items = []*queries.AttendanceDTO{}
	}
	httpx.WriteSuccess(w, http.StatusOK, listPage{Items: items, Page: page, PerPage: perPage, Total: res.Total})
}

// pageParams reads page (default 1) and per_page (default 20, max 100).
func pageParams(w http.ResponseWriter, pageS, perPageS string) (page, perPage int, ok bool) {
	page, perPage = 1, defaultPerPage
	var err error
	if pageS != "" {
		if page, err = strconv.Atoi(pageS); err != nil || page < 1 {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "page must be a positive integer")
			return 0, 0, false
		}
	}
	if perPageS != "" {
		if perPage, err = strconv.Atoi(perPageS); err != nil || perPage < 1 || perPage > maxPerPage {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "per_page must be between 1 and 100")
			return 0, 0, false
		}
	}
	return page, perPage, true
}

// writeError maps domain errors to fixed messages; anything else is logged and answered 500.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "attendance record not found")
	case errors.Is(err, domain.ErrCannotCheckInTwiceSameDay):
		httpx.WriteError(w, http.StatusConflict, "already checked in today")
	case errors.Is(err, domain.ErrAlreadyCheckedOut):
		httpx.WriteError(w, http.StatusConflict, "already checked out today")
	case errors.Is(err, domain.ErrCannotCheckOutBeforeCheckIn):
		httpx.WriteError(w, http.StatusConflict, "check in before checking out")
	case errors.Is(err, domain.ErrInvalidTimeRange):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "check-in must be before check-out")
	case errors.Is(err, domain.ErrInvalidStatus):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid status")
	case errors.Is(err, domain.ErrInvalidEmployeeID):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid employee id")
	case errors.Is(err, domain.ErrInvalidTenantID), errors.Is(err, domain.ErrInvalidDate), errors.Is(err, domain.ErrMissingRequiredField):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid request")
	default:
		slog.ErrorContext(r.Context(), "attendance request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
	}
}
