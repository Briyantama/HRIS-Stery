package notification_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Briyantama/HRIS-Stery/services/notification"
	"github.com/Briyantama/HRIS-Stery/services/notification/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/notification/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
	"github.com/Briyantama/stery-kit/go/authn"
)

const (
	tenantA   = "11111111-1111-4111-8111-111111111111"
	userID    = "22222222-2222-4222-8222-222222222222"
	notifID   = "33333333-3333-4333-8333-333333333333"
	adminTok  = "admin"
	memberTok = "member"
)

type stubVerifier struct{}

func (stubVerifier) Verify(_ context.Context, token string) (authn.Claims, error) {
	switch token {
	case adminTok:
		return authn.Claims{UserID: userID, TenantID: tenantA, Roles: []string{"hr_admin"}}, nil
	case memberTok:
		return authn.Claims{UserID: userID, TenantID: tenantA, Roles: []string{"employee"}}, nil
	}
	return authn.Claims{}, errors.New("bad token")
}

type stub struct {
	err      error
	sentCmd  commands.SendNotificationCommand
	markCmd  commands.MarkNotificationReadCommand
	listQ    queries.ListNotificationsQuery
	listJSON queries.NotificationDTO
}

func (s *stub) sendH() commands.SendNotificationHandler     { return sendStub{s} }
func (s *stub) markH() commands.MarkNotificationReadHandler { return markStub{s} }
func (s *stub) cfgH() commands.UpdateChannelConfigHandler   { return cfgStub{s} }
func (s *stub) listH() queries.ListNotificationsHandler     { return listStub{s} }
func (s *stub) statusH() queries.GetDeliveryStatusHandler   { return statusStub{s} }

type sendStub struct{ s *stub }

func (h sendStub) Handle(_ context.Context, c commands.SendNotificationCommand) (*commands.SendNotificationResult, error) {
	h.s.sentCmd = c
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &commands.SendNotificationResult{NotificationIDs: []string{notifID}}, nil
}

type markStub struct{ s *stub }

func (h markStub) Handle(_ context.Context, c commands.MarkNotificationReadCommand) (*commands.MarkNotificationReadResult, error) {
	h.s.markCmd = c
	return &commands.MarkNotificationReadResult{UpdatedCount: 1}, h.s.err
}

type cfgStub struct{ s *stub }

func (h cfgStub) Handle(_ context.Context, c commands.UpdateChannelConfigCommand) (*commands.UpdateChannelConfigResult, error) {
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &commands.UpdateChannelConfigResult{TenantID: c.TenantID, EnabledChannels: []string{"EMAIL"}}, nil
}

type listStub struct{ s *stub }

func (h listStub) Handle(_ context.Context, q queries.ListNotificationsQuery) (*queries.ListNotificationsResult, error) {
	h.s.listQ = q
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &queries.ListNotificationsResult{Notifications: []*queries.NotificationDTO{{ID: notifID}}, Total: 41}, nil
}

type statusStub struct{ s *stub }

func (h statusStub) Handle(_ context.Context, q queries.GetDeliveryStatusQuery) (*queries.DeliveryStatusDTO, error) {
	if h.s.err != nil {
		return nil, h.s.err
	}
	return &queries.DeliveryStatusDTO{NotificationID: q.NotificationID, Status: "SENT"}, nil
}

func newServer(s *stub) http.Handler {
	mux := http.NewServeMux()
	notification.NewHTTPHandler(s.sendH(), s.markH(), s.cfgH(), s.listH(), s.statusH(), stubVerifier{}).Register(mux)
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

type env struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *string         `json:"error"`
}

func TestStatusTable(t *testing.T) {
	sendBody := `{"recipient_id":"` + userID + `","channels":["EMAIL"],"template_key":"x"}`
	rows := []struct {
		name, method, path, token, body string
		err                             error
		want                            int
	}{
		{"list ok", "GET", "/api/v1/notifications", memberTok, "", nil, 200},
		{"list bad per_page", "GET", "/api/v1/notifications?per_page=500", memberTok, "", nil, 422},
		{"list internal", "GET", "/api/v1/notifications", memberTok, "", errors.New("db down"), 500},
		{"mark ok", "PATCH", "/api/v1/notifications", memberTok, `{"notification_ids":["` + notifID + `"]}`, nil, 200},
		{"mark empty", "PATCH", "/api/v1/notifications", memberTok, `{"notification_ids":[]}`, nil, 422},
		{"send ok", "POST", "/api/v1/notifications", adminTok, sendBody, nil, 201},
		{"send malformed json", "POST", "/api/v1/notifications", adminTok, `{`, nil, 400},
		{"send bad recipient", "POST", "/api/v1/notifications", adminTok, `{"recipient_id":"x","channels":["EMAIL"]}`, nil, 422},
		{"send bad channel", "POST", "/api/v1/notifications", adminTok, `{"recipient_id":"` + userID + `","channels":["SMS"]}`, nil, 422},
		{"send missing template", "POST", "/api/v1/notifications", adminTok, sendBody, domain.ErrInvalidTemplateKey, 422},
		{"send internal", "POST", "/api/v1/notifications", adminTok, sendBody, errors.New("db down"), 500},
		{"status ok", "GET", "/api/v1/notifications/" + notifID, adminTok, "", nil, 200},
		{"status missing", "GET", "/api/v1/notifications/" + notifID, adminTok, "", domain.ErrNotificationNotFound, 404},
		{"status malformed id", "GET", "/api/v1/notifications/nope", adminTok, "", nil, 404},
		{"channel ok", "PUT", "/api/v1/notification-channels/email", adminTok, `{"smtp_host":"h","from_address":"a@b.c"}`, nil, 200},
		{"channel invalid config", "PUT", "/api/v1/notification-channels/email", adminTok, `{}`, domain.ErrInvalidEmailConfig, 422},
		{"channel unknown", "PUT", "/api/v1/notification-channels/sms", adminTok, `{}`, nil, 404},
		{"no token", "GET", "/api/v1/notifications", "", "", nil, 401},
		{"bad token", "GET", "/api/v1/notifications", "zzz", "", nil, 401},
		{"wrong role send", "POST", "/api/v1/notifications", memberTok, sendBody, nil, 403},
		{"wrong role status", "GET", "/api/v1/notifications/" + notifID, memberTok, "", nil, 403},
		{"wrong role channel", "PUT", "/api/v1/notification-channels/email", memberTok, `{}`, nil, 403},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			rec := do(newServer(&stub{err: r.err}), r.method, r.path, r.token, r.body)
			if rec.Code != r.want {
				t.Fatalf("status = %d, want %d: %s", rec.Code, r.want, rec.Body)
			}
			var e env
			if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
				t.Fatalf("body is not an envelope: %v", err)
			}
			if e.Success != (r.want < 400) {
				t.Fatalf("success = %v for status %d", e.Success, r.want)
			}
			if r.want == 500 && (e.Error == nil || *e.Error != "internal error") {
				t.Fatalf("500 must carry fixed text, got %v", e.Error)
			}
		})
	}
}

func TestListUsesTokenRecipientAndPaginates(t *testing.T) {
	s := &stub{}
	rec := do(newServer(s), "GET", "/api/v1/notifications?page=3&per_page=10", memberTok, "")
	if rec.Code != 200 {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if s.listQ.TenantID != tenantA || s.listQ.RecipientID != userID || s.listQ.Limit != 10 || s.listQ.Offset != 20 {
		t.Fatalf("unexpected query: %+v", s.listQ)
	}
	var e env
	_ = json.Unmarshal(rec.Body.Bytes(), &e)
	var page struct {
		Items   []map[string]any `json:"items"`
		Page    int              `json:"page"`
		PerPage int              `json:"per_page"`
		Total   int              `json:"total"`
	}
	if err := json.Unmarshal(e.Data, &page); err != nil || len(page.Items) != 1 || page.Page != 3 || page.PerPage != 10 || page.Total != 41 {
		t.Fatalf("unexpected page: %s", e.Data)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
	}
}

func TestSendIgnoresBodyTenantAndSetsLocation(t *testing.T) {
	s := &stub{}
	body := `{"tenant_id":"99999999-9999-4999-8999-999999999999","recipient_id":"` + userID + `","channels":["email"],"template_key":"x"}`
	rec := do(newServer(s), "POST", "/api/v1/notifications", adminTok, body)
	if rec.Code != 201 {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if s.sentCmd.TenantID != tenantA {
		t.Fatalf("tenant = %q, want the token claim", s.sentCmd.TenantID)
	}
	if got := rec.Header().Get("Location"); got != "/api/v1/notifications/"+notifID {
		t.Fatalf("Location = %q", got)
	}
}

func TestMarkReadScopedToTokenUser(t *testing.T) {
	s := &stub{}
	do(newServer(s), "PATCH", "/api/v1/notifications", memberTok, `{"notification_ids":["`+notifID+`"]}`)
	if s.markCmd.TenantID != tenantA || s.markCmd.RecipientID != userID {
		t.Fatalf("unexpected command: %+v", s.markCmd)
	}
}
