package notification

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Briyantama/HRIS-Stery/services/notification/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/notification/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/httpx"
)

const (
	maxBody        = 1 << 20
	defaultPerPage = 20
	maxPerPage     = 100
	maxMarkIDs     = 100
	roleHRAdmin    = "hr_admin"
)

// HTTPHandler serves /api/v1/notifications and /api/v1/notification-channels.
type HTTPHandler struct {
	send     commands.SendNotificationHandler
	markRead commands.MarkNotificationReadHandler
	config   commands.UpdateChannelConfigHandler
	list     queries.ListNotificationsHandler
	status   queries.GetDeliveryStatusHandler
	verifier authn.Verifier
}

func NewHTTPHandler(
	send commands.SendNotificationHandler,
	markRead commands.MarkNotificationReadHandler,
	config commands.UpdateChannelConfigHandler,
	list queries.ListNotificationsHandler,
	status queries.GetDeliveryStatusHandler,
	verifier authn.Verifier,
) *HTTPHandler {
	return &HTTPHandler{send: send, markRead: markRead, config: config, list: list, status: status, verifier: verifier}
}

// Register mounts the routes: list and mark-read for any authenticated user (recipient = token user),
// send, delivery status and channel config for hr_admin.
func (h *HTTPHandler) Register(mux *http.ServeMux) {
	authed := authn.Authenticate(h.verifier)
	admin := authn.WithRole(roleHRAdmin)
	mux.Handle("GET /api/v1/notifications", authed(http.HandlerFunc(h.listNotifications)))
	mux.Handle("PATCH /api/v1/notifications", authed(http.HandlerFunc(h.markNotificationsRead)))
	mux.Handle("POST /api/v1/notifications", authed(admin(http.HandlerFunc(h.sendNotification))))
	mux.Handle("GET /api/v1/notifications/{notification_id}", authed(admin(http.HandlerFunc(h.deliveryStatus))))
	mux.Handle("PUT /api/v1/notification-channels/{channel}", authed(admin(http.HandlerFunc(h.updateChannel))))
}

// caller returns the verified claims and the tenant id; it writes 401 when either is unusable.
func caller(w http.ResponseWriter, r *http.Request) (authn.Claims, domain.TenantID, bool) {
	c, ok := authn.FromContext(r.Context())
	if ok {
		if tenant, err := domain.NewTenantID(c.TenantID); err == nil {
			return c, tenant, true
		}
	}
	httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
	return authn.Claims{}, "", false
}

func internalError(w http.ResponseWriter, r *http.Request, msg string, err error, tenant domain.TenantID) {
	slog.ErrorContext(r.Context(), msg, "error", err, "tenant_id", tenant.String())
	httpx.WriteError(w, http.StatusInternalServerError, "internal error")
}

func parsePaging(r *http.Request) (page, perPage int, ok bool) {
	page, perPage = 1, defaultPerPage
	q := r.URL.Query()
	if s := q.Get("page"); s != "" {
		p, err := strconv.Atoi(s)
		if err != nil || p < 1 {
			return 0, 0, false
		}
		page = p
	}
	if s := q.Get("per_page"); s != "" {
		p, err := strconv.Atoi(s)
		if err != nil || p < 1 || p > maxPerPage {
			return 0, 0, false
		}
		perPage = p
	}
	return page, perPage, true
}

func (h *HTTPHandler) listNotifications(w http.ResponseWriter, r *http.Request) {
	claims, tenant, ok := caller(w, r)
	if !ok {
		return
	}
	page, perPage, ok := parsePaging(r)
	unread := false
	if s := r.URL.Query().Get("unread_only"); s != "" {
		v, err := strconv.ParseBool(s)
		ok = ok && err == nil
		unread = v
	}
	if !ok {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid query parameters")
		return
	}
	if _, err := domain.NewRecipientID(claims.UserID); err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	res, err := h.list.Handle(r.Context(), queries.ListNotificationsQuery{
		TenantID:    tenant.String(),
		RecipientID: claims.UserID,
		UnreadOnly:  unread,
		Limit:       perPage,
		Offset:      (page - 1) * perPage,
	})
	if err != nil {
		internalError(w, r, "list notifications failed", err, tenant)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, map[string]any{
		"items": res.Notifications, "page": page, "per_page": perPage, "total": res.Total,
	})
}

func (h *HTTPHandler) markNotificationsRead(w http.ResponseWriter, r *http.Request) {
	claims, tenant, ok := caller(w, r)
	if !ok {
		return
	}
	var body struct {
		NotificationIDs []string `json:"notification_ids"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	if len(body.NotificationIDs) == 0 || len(body.NotificationIDs) > maxMarkIDs {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "notification_ids must hold 1 to 100 ids")
		return
	}
	res, err := h.markRead.Handle(r.Context(), commands.MarkNotificationReadCommand{
		TenantID:        tenant.String(),
		RecipientID:     claims.UserID,
		NotificationIDs: body.NotificationIDs,
	})
	if err != nil {
		internalError(w, r, "mark notifications read failed", err, tenant)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, map[string]int{"updated_count": res.UpdatedCount})
}

func (h *HTTPHandler) sendNotification(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := caller(w, r)
	if !ok {
		return
	}
	var body struct {
		TenantID    string            `json:"tenant_id"` // accepted for old clients, ignored: the tenant comes from the token
		RecipientID string            `json:"recipient_id"`
		Channels    []string          `json:"channels"`
		TemplateKey string            `json:"template_key"`
		Variables   map[string]string `json:"variables"`
		Subject     string            `json:"subject"`
		Body        string            `json:"body"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}
	if _, err := domain.NewRecipientID(body.RecipientID); err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "recipient_id must be a uuid")
		return
	}
	if len(body.Channels) == 0 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "channels is required")
		return
	}
	channels := make([]domain.ChannelType, len(body.Channels))
	for i, s := range body.Channels {
		ch, ok := parseChannel(s)
		if !ok {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "unsupported channel")
			return
		}
		channels[i] = ch
	}

	res, err := h.send.Handle(r.Context(), commands.SendNotificationCommand{
		TenantID:    tenant.String(),
		RecipientID: body.RecipientID,
		TemplateKey: body.TemplateKey,
		Variables:   body.Variables,
		Channels:    channels,
		Subject:     body.Subject,
		Body:        body.Body,
	})
	if errors.Is(err, domain.ErrInvalidTemplateKey) {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "template_key is required")
		return
	}
	if err != nil {
		internalError(w, r, "send notification failed", err, tenant)
		return
	}

	type receipt struct {
		NotificationID string `json:"notification_id"`
		Channel        string `json:"channel"`
		Status         string `json:"status"`
	}
	receipts := make([]receipt, len(res.NotificationIDs))
	for i, id := range res.NotificationIDs {
		receipts[i] = receipt{NotificationID: id, Channel: channels[i].String(), Status: domain.NotificationStatusPending.String()}
	}
	if len(receipts) > 0 {
		w.Header().Set("Location", "/api/v1/notifications/"+receipts[0].NotificationID)
	}
	httpx.WriteSuccess(w, http.StatusCreated, map[string]any{"receipts": receipts})
}

func (h *HTTPHandler) deliveryStatus(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := caller(w, r)
	if !ok {
		return
	}
	id := r.PathValue("notification_id")
	if _, err := domain.NewNotificationID(id); err != nil {
		httpx.WriteError(w, http.StatusNotFound, "notification not found")
		return
	}
	res, err := h.status.Handle(r.Context(), queries.GetDeliveryStatusQuery{TenantID: tenant.String(), NotificationID: id})
	if errors.Is(err, domain.ErrNotificationNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "notification not found")
		return
	}
	if err != nil {
		internalError(w, r, "get delivery status failed", err, tenant)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, res)
}

func (h *HTTPHandler) updateChannel(w http.ResponseWriter, r *http.Request) {
	_, tenant, ok := caller(w, r)
	if !ok {
		return
	}
	channel, ok := parseChannel(r.PathValue("channel"))
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "channel not found")
		return
	}
	var body struct {
		SMTPHost     string `json:"smtp_host"`
		SMTPPort     int    `json:"smtp_port"`
		SMTPUser     string `json:"smtp_user"`
		SMTPPassword string `json:"smtp_password"`
		FromAddress  string `json:"from_address"`
		FromName     string `json:"from_name"`
	}
	if !httpx.DecodeJSON(w, r, &body, maxBody) {
		return
	}

	cmd := commands.UpdateChannelConfigCommand{TenantID: tenant.String(), Channels: []domain.ChannelType{channel}}
	if channel == domain.ChannelTypeEmail {
		if body.SMTPPort == 0 {
			body.SMTPPort = 587
		}
		if body.SMTPPort < 1 || body.SMTPPort > 65535 {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "smtp_port is out of range")
			return
		}
		cmd.EmailConfig = &domain.EmailConfig{
			SMTPHost: body.SMTPHost, SMTPPort: body.SMTPPort, SMTPUser: body.SMTPUser,
			SMTPPassword: body.SMTPPassword, FromAddress: body.FromAddress, FromName: body.FromName,
		}
	}
	res, err := h.config.Handle(r.Context(), cmd)
	if errors.Is(err, domain.ErrInvalidEmailConfig) {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "smtp_host and from_address are required")
		return
	}
	if err != nil {
		internalError(w, r, "update channel config failed", err, tenant)
		return
	}
	httpx.WriteSuccess(w, http.StatusOK, map[string]any{"enabled_channels": res.EnabledChannels})
}

// parseChannel accepts EMAIL and IN_APP in any case, with "-" or "_" (so the path segment can be in-app).
func parseChannel(s string) (domain.ChannelType, bool) {
	ch, err := domain.ChannelTypeFromString(strings.ToUpper(strings.ReplaceAll(s, "-", "_")))
	return ch, err == nil
}
