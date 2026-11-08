package leave

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/leave/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/leave/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/httpx"
)

const (
	maxBody        = 1 << 20
	defaultPerPage = 20
	maxPerPage     = 100
	dateLayout     = "2006-01-02"
)

type creator interface {
	Handle(context.Context, commands.CreateLeaveRequestCommand) (*commands.CreateLeaveRequestResult, error)
}
type approver interface {
	Handle(context.Context, commands.ApproveLeaveRequestCommand) (*commands.ApproveLeaveRequestResult, error)
}
type rejecter interface {
	Handle(context.Context, commands.RejectLeaveRequestCommand) (*commands.RejectLeaveRequestResult, error)
}
type canceller interface {
	Handle(context.Context, commands.CancelLeaveRequestCommand) (*commands.CancelLeaveRequestResult, error)
}
type getter interface {
	Handle(context.Context, queries.GetLeaveRequestQuery) (*queries.LeaveRequestDTO, error)
}
type lister interface {
	Handle(context.Context, queries.ListLeaveRequestsQuery) (*queries.ListLeaveRequestsResult, error)
}
type balanceGetter interface {
	Handle(context.Context, queries.GetLeaveBalanceQuery) (*queries.LeaveBalanceResult, error)
}
type typeLister interface {
	Handle(context.Context, queries.ListLeaveTypesQuery) (*queries.ListLeaveTypesResult, error)
}

// HTTPHandler serves the /api/v1 leave routes.
type HTTPHandler struct {
	create   creator
	approve  approver
	reject   rejecter
	cancel   canceller
	get      getter
	list     lister
	balance  balanceGetter
	types    typeLister
	verifier authn.Verifier
}

// NewHTTPHandler builds the handler from the command and query handlers.
func NewHTTPHandler(c creator, a approver, rj rejecter, cn canceller, g getter, l lister, b balanceGetter, t typeLister, v authn.Verifier) *HTTPHandler {
	return &HTTPHandler{create: c, approve: a, reject: rj, cancel: cn, get: g, list: l, balance: b, types: t, verifier: v}
}

// Register adds the leave routes to mux. Every route requires a valid bearer token.
func (h *HTTPHandler) Register(mux *http.ServeMux) {
	auth := authn.Authenticate(h.verifier)
	approvers := authn.WithRole("manager", "hr_admin")
	mux.Handle("POST /api/v1/leaves", auth(http.HandlerFunc(h.apply)))
	mux.Handle("GET /api/v1/leaves", auth(http.HandlerFunc(h.listLeaves)))
	mux.Handle("GET /api/v1/leaves/{id}", auth(http.HandlerFunc(h.getLeave)))
	mux.Handle("POST /api/v1/leaves/{id}/approve", auth(approvers(http.HandlerFunc(h.approveLeave))))
	mux.Handle("POST /api/v1/leaves/{id}/reject", auth(approvers(http.HandlerFunc(h.rejectLeave))))
	mux.Handle("POST /api/v1/leaves/{id}/cancel", auth(http.HandlerFunc(h.cancelLeave)))
	mux.Handle("GET /api/v1/leave-balance/{employee_id}", auth(http.HandlerFunc(h.getBalance)))
	mux.Handle("GET /api/v1/leave-types", auth(http.HandlerFunc(h.listTypes)))
}

// privileged reports whether the caller may act on other employees' leave.
func privileged(c authn.Claims) bool {
	for _, r := range c.Roles {
		if strings.EqualFold(r, "manager") || strings.EqualFold(r, "hr_admin") {
			return true
		}
	}
	return false
}

func isAdmin(c authn.Claims) bool {
	for _, r := range c.Roles {
		if strings.EqualFold(r, "hr_admin") {
			return true
		}
	}
	return false
}

// clientErrors maps domain errors to a status. The text sent is the sentinel's fixed message.
var clientErrors = []struct {
	err    error
	status int
}{
	{domain.ErrNotFound, http.StatusNotFound},
	{domain.ErrInsufficientLeaveBalance, http.StatusConflict},
	{domain.ErrCannotApproveAlreadyRejected, http.StatusConflict},
	{domain.ErrCannotRejectAlreadyApproved, http.StatusConflict},
	{domain.ErrCannotCancelRejected, http.StatusConflict},
	{domain.ErrCannotCancelPending, http.StatusConflict},
	{domain.ErrInvalidStatusTransition, http.StatusConflict},
	{domain.ErrLeaveOverlapExists, http.StatusConflict},
	{domain.ErrAlreadyExists, http.StatusConflict},
	{domain.ErrInvalidLeaveRequestID, http.StatusUnprocessableEntity},
	{domain.ErrInvalidLeaveTypeID, http.StatusUnprocessableEntity},
	{domain.ErrInvalidEmployeeID, http.StatusUnprocessableEntity},
	{domain.ErrInvalidTenantID, http.StatusUnprocessableEntity},
	{domain.ErrInvalidDate, http.StatusUnprocessableEntity},
	{domain.ErrInvalidDaysCount, http.StatusUnprocessableEntity},
	{domain.ErrInvalidLeaveStatus, http.StatusUnprocessableEntity},
	{domain.ErrMissingRequiredField, http.StatusUnprocessableEntity},
}

func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	for _, ce := range clientErrors {
		if errors.Is(err, ce.err) {
			msg := ce.err.Error()
			if ce.status == http.StatusNotFound {
				msg = "not found"
			}
			httpx.WriteError(w, ce.status, msg)
			return
		}
	}
	slog.ErrorContext(r.Context(), "leave request failed", "method", r.Method, "path", r.URL.Path, "error", err)
	httpx.WriteError(w, http.StatusInternalServerError, "internal error")
}

func claims(w http.ResponseWriter, r *http.Request) (authn.Claims, bool) {
	c, ok := authn.FromContext(r.Context())
	if !ok || c.TenantID == "" || c.UserID == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return c, false
	}
	return c, true
}

func (h *HTTPHandler) apply(w http.ResponseWriter, r *http.Request) {
	c, ok := claims(w, r)
	if !ok {
		return
	}
	var body struct {
		LeaveTypeID string `json:"leave_type_id"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		Reason      string `json:"reason"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	start, serr := time.Parse(dateLayout, body.StartDate)
	end, eerr := time.Parse(dateLayout, body.EndDate)
	if body.LeaveTypeID == "" || serr != nil || eerr != nil || end.Before(start) {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "leave_type_id, start_date and end_date are required, and end_date must not precede start_date")
		return
	}
	// ponytail: calendar days, inclusive; weekends and holidays are not excluded. Add a calendar when the business rule exists.
	days := int(end.Sub(start).Hours()/24) + 1
	res, err := h.create.Handle(r.Context(), commands.CreateLeaveRequestCommand{
		TenantID:    c.TenantID,
		EmployeeID:  c.UserID,
		LeaveTypeID: body.LeaveTypeID,
		StartDate:   start,
		EndDate:     end,
		DaysCount:   days,
		Reason:      body.Reason,
		ActorID:     c.UserID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	dto, err := h.get.Handle(r.Context(), queries.GetLeaveRequestQuery{TenantID: c.TenantID, LeaveRequestID: res.LeaveRequestID})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/leaves/"+dto.ID)
	httpx.WriteSuccess(w, http.StatusCreated, dto)
}

func (h *HTTPHandler) listLeaves(w http.ResponseWriter, r *http.Request) {
	c, ok := claims(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	page, perPage, ok := pageParams(w, q.Get("page"), q.Get("per_page"))
	if !ok {
		return
	}
	query := queries.ListLeaveRequestsQuery{TenantID: c.TenantID, Offset: (page - 1) * perPage, Limit: perPage}
	if v := q.Get("status"); v != "" {
		query.Status = &v
	}
	if v := q.Get("year"); v != "" {
		y, err := strconv.Atoi(v)
		if err != nil || y < 1 {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid year")
			return
		}
		query.Year = &y
	}
	if privileged(c) {
		if v := q.Get("employee_id"); v != "" {
			query.EmployeeID = &v
		}
		if v := q.Get("approver_id"); v != "" {
			query.ApprovedByID = &v
		}
	} else {
		query.EmployeeID = &c.UserID
	}
	res, err := h.list.Handle(r.Context(), query)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, map[string]any{"items": res.Records, "page": page, "per_page": perPage, "total": res.Total})
}

func pageParams(w http.ResponseWriter, rawPage, rawPer string) (page, perPage int, ok bool) {
	page, perPage = 1, defaultPerPage
	var err error
	if rawPage != "" {
		if page, err = strconv.Atoi(rawPage); err != nil || page < 1 {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid page")
			return 0, 0, false
		}
	}
	if rawPer != "" {
		if perPage, err = strconv.Atoi(rawPer); err != nil || perPage < 1 {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid per_page")
			return 0, 0, false
		}
		perPage = min(perPage, maxPerPage)
	}
	return page, perPage, true
}

func (h *HTTPHandler) getLeave(w http.ResponseWriter, r *http.Request) {
	c, ok := claims(w, r)
	if !ok {
		return
	}
	dto, err := h.get.Handle(r.Context(), queries.GetLeaveRequestQuery{TenantID: c.TenantID, LeaveRequestID: r.PathValue("id")})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if !privileged(c) && dto.EmployeeID != c.UserID {
		httpx.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, dto)
}

func (h *HTTPHandler) approveLeave(w http.ResponseWriter, r *http.Request) {
	c, ok := claims(w, r)
	if !ok {
		return
	}
	res, err := h.approve.Handle(r.Context(), commands.ApproveLeaveRequestCommand{
		TenantID: c.TenantID, LeaveRequestID: r.PathValue("id"), ApprovedByID: c.UserID, ActorID: c.UserID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	h.respondUpdated(w, r, c.TenantID, res.LeaveRequestID)
}

func (h *HTTPHandler) rejectLeave(w http.ResponseWriter, r *http.Request) {
	c, ok := claims(w, r)
	if !ok {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	res, err := h.reject.Handle(r.Context(), commands.RejectLeaveRequestCommand{
		TenantID: c.TenantID, LeaveRequestID: r.PathValue("id"), Reason: body.Reason, ActorID: c.UserID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	h.respondUpdated(w, r, c.TenantID, res.LeaveRequestID)
}

func (h *HTTPHandler) cancelLeave(w http.ResponseWriter, r *http.Request) {
	c, ok := claims(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	dto, err := h.get.Handle(r.Context(), queries.GetLeaveRequestQuery{TenantID: c.TenantID, LeaveRequestID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	// Employees cancel their own leave; hr_admin may cancel anyone's.
	if !isAdmin(c) && dto.EmployeeID != c.UserID {
		httpx.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	res, err := h.cancel.Handle(r.Context(), commands.CancelLeaveRequestCommand{
		TenantID: c.TenantID, LeaveRequestID: id, Reason: "Cancelled by " + c.UserID, ActorID: c.UserID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	h.respondUpdated(w, r, c.TenantID, res.LeaveRequestID)
}

func (h *HTTPHandler) respondUpdated(w http.ResponseWriter, r *http.Request, tenantID, id string) {
	dto, err := h.get.Handle(r.Context(), queries.GetLeaveRequestQuery{TenantID: tenantID, LeaveRequestID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, dto)
}

func (h *HTTPHandler) getBalance(w http.ResponseWriter, r *http.Request) {
	c, ok := claims(w, r)
	if !ok {
		return
	}
	employeeID := r.PathValue("employee_id")
	if !privileged(c) && employeeID != c.UserID {
		httpx.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	year := time.Now().Year()
	if v := r.URL.Query().Get("year"); v != "" {
		y, err := strconv.Atoi(v)
		if err != nil || y < 1 {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid year")
			return
		}
		year = y
	}
	res, err := h.balance.Handle(r.Context(), queries.GetLeaveBalanceQuery{TenantID: c.TenantID, EmployeeID: employeeID, Year: year})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res)
}

func (h *HTTPHandler) listTypes(w http.ResponseWriter, r *http.Request) {
	c, ok := claims(w, r)
	if !ok {
		return
	}
	res, err := h.types.Handle(r.Context(), queries.ListLeaveTypesQuery{TenantID: c.TenantID})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res.LeaveTypes)
}
