package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
	"github.com/Briyantama/stery-kit/go/pgtenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) domain.NotificationRepository {
	return &NotificationRepository{pool: pool}
}

const notificationColumns = `id, recipient_id, channel, template_key, variables, subject, body,
	status, delivery_id, delivery_error, scheduled_at, sent_at, read_at, created_at, updated_at`

func (r *NotificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	return pgtenant.WithTenantTx(ctx, r.pool, notif.TenantID().String(), "", func(tx pgx.Tx) error {
		variablesJSON, err := json.Marshal(notif.Variables())
		if err != nil {
			return fmt.Errorf("create notification: marshal variables: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO notification.notifications
			(id, tenant_id, recipient_id, channel, template_key, variables, subject, body, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			notif.ID().String(),
			notif.TenantID().String(),
			notif.RecipientID().String(),
			notif.Channel().String(),
			notif.TemplateKey(),
			variablesJSON,
			notif.Subject(),
			notif.Body(),
			notif.Status().String(),
			notif.CreatedAt(),
			notif.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("create notification: %w", err)
		}
		return nil
	})
}

func (r *NotificationRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.NotificationID) (*domain.Notification, error) {
	var notif *domain.Notification
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT `+notificationColumns+` FROM notification.notifications WHERE id = $1 AND tenant_id = $2`,
			id.String(), tenantID.String())
		if err != nil {
			return fmt.Errorf("get notification: %w", err)
		}
		defer rows.Close()
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return fmt.Errorf("get notification: %w", err)
			}
			return domain.ErrNotificationNotFound
		}
		notif, err = scanNotification(rows, tenantID)
		return err
	})
	return notif, err
}

func (r *NotificationRepository) ListByRecipient(ctx context.Context, tenantID domain.TenantID, recipientID domain.RecipientID, filters domain.ListFilters) ([]*domain.Notification, error) {
	var notifs []*domain.Notification
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		where, args := listWhere(tenantID, recipientID, filters)
		query := `SELECT ` + notificationColumns + ` FROM notification.notifications WHERE ` + where + ` ORDER BY created_at DESC`
		if filters.Limit > 0 {
			args = append(args, filters.Limit)
			query += fmt.Sprintf(" LIMIT $%d", len(args))
		}
		if filters.Offset > 0 {
			args = append(args, filters.Offset)
			query += fmt.Sprintf(" OFFSET $%d", len(args))
		}
		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("list notifications: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			n, err := scanNotification(rows, tenantID)
			if err != nil {
				return err
			}
			notifs = append(notifs, n)
		}
		return rows.Err()
	})
	return notifs, err
}

func (r *NotificationRepository) CountByRecipient(ctx context.Context, tenantID domain.TenantID, recipientID domain.RecipientID, filters domain.ListFilters) (int, error) {
	var total int
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		where, args := listWhere(tenantID, recipientID, filters)
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM notification.notifications WHERE `+where, args...).Scan(&total); err != nil {
			return fmt.Errorf("count notifications: %w", err)
		}
		return nil
	})
	return total, err
}

// listWhere builds the filter shared by the list and count queries (no limit/offset).
func listWhere(tenantID domain.TenantID, recipientID domain.RecipientID, f domain.ListFilters) (string, []any) {
	where := "tenant_id = $1 AND recipient_id = $2"
	args := []any{tenantID.String(), recipientID.String()}
	if f.Status != domain.NotificationStatusUnknown {
		args = append(args, f.Status.String())
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if f.UnreadOnly {
		args = append(args, domain.NotificationStatusSent.String())
		where += fmt.Sprintf(" AND read_at IS NULL AND status = $%d", len(args))
	}
	return where, args
}

func scanNotification(rows pgx.Rows, tenantID domain.TenantID) (*domain.Notification, error) {
	var (
		notifID, recID, channelStr, templateKey, statusStr string
		variablesJSON                                      []byte
		subject, body, deliveryID, deliveryError           *string
		scheduledAt, sentAt, readAt                        *time.Time
		createdAt, updatedAt                               time.Time
	)
	if err := rows.Scan(&notifID, &recID, &channelStr, &templateKey, &variablesJSON, &subject, &body,
		&statusStr, &deliveryID, &deliveryError, &scheduledAt, &sentAt, &readAt, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("scan notification: %w", err)
	}

	variables := map[string]string{}
	if variablesJSON != nil {
		if err := json.Unmarshal(variablesJSON, &variables); err != nil {
			return nil, fmt.Errorf("scan notification: decode variables: %w", err)
		}
	}
	notificationID, err := domain.NewNotificationID(notifID)
	if err != nil {
		return nil, err
	}
	recipientID, err := domain.NewRecipientID(recID)
	if err != nil {
		return nil, err
	}
	channel, err := domain.ChannelTypeFromString(channelStr)
	if err != nil {
		return nil, err
	}
	status, err := domain.NotificationStatusFromString(statusStr)
	if err != nil {
		return nil, err
	}
	return domain.RehydrateNotification(notificationID, tenantID, recipientID, channel, templateKey, variables,
		deref(subject), deref(body), status, deref(deliveryID), deref(deliveryError),
		scheduledAt, sentAt, readAt, createdAt, updatedAt), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *NotificationRepository) Update(ctx context.Context, notif *domain.Notification) error {
	return pgtenant.WithTenantTx(ctx, r.pool, notif.TenantID().String(), "", func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE notification.notifications
			SET status = $1, delivery_id = $2, delivery_error = $3, sent_at = $4, read_at = $5, updated_at = $6
			WHERE id = $7 AND tenant_id = $8`,
			notif.Status().String(),
			notif.DeliveryID(),
			notif.DeliveryError(),
			notif.SentAt(),
			notif.ReadAt(),
			time.Now().UTC(),
			notif.ID().String(),
			notif.TenantID().String(),
		)
		if err != nil {
			return fmt.Errorf("update notification: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrNotificationNotFound
		}
		return nil
	})
}
