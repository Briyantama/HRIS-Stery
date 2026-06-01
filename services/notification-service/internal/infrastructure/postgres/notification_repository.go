package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sharedpostgres "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) domain.NotificationRepository {
	return &NotificationRepository{
		pool: pool,
	}
}

func (r *NotificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	return sharedpostgres.WithTenantTx(ctx, r.pool, sharedpostgres.TenantID(notif.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		variablesJSON, _ := json.Marshal(notif.Variables())

		query := `
			INSERT INTO notification.notifications
			(id, tenant_id, recipient_id, channel, template_key, variables, subject, body, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`

		_, err := tx.Exec(ctx, query,
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

		return fmt.Errorf("create notification: %w", err)
	})
}

func (r *NotificationRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.NotificationID) (*domain.Notification, error) {
	var notif *domain.Notification
	var err error

	innerErr := sharedpostgres.WithTenantTx(ctx, r.pool, sharedpostgres.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		notif, err = r.getByIDFromTx(ctx, tx, tenantID, id)
		return err
	})

	return notif, innerErr
}

func (r *NotificationRepository) getByIDFromTx(ctx context.Context, tx pgx.Tx, tenantID domain.TenantID, id domain.NotificationID) (*domain.Notification, error) {
	query := `
		SELECT id, tenant_id, recipient_id, channel, template_key, variables, subject, body,
		       status, delivery_id, delivery_error, scheduled_at, sent_at, read_at, created_at, updated_at
		FROM notification.notifications
		WHERE id = $1 AND tenant_id = $2
	`

	var (
		notifID       string
		recID         string
		channelStr    string
		templateKey   string
		variablesJSON []byte
		subject       string
		body          string
		statusStr     string
		deliveryID    *string
		deliveryError *string
		scheduledAt   *time.Time
		sentAt        *time.Time
		readAt        *time.Time
		createdAt     time.Time
		updatedAt     time.Time
	)

	err := tx.QueryRow(ctx, query, id.String(), tenantID.String()).Scan(
		&notifID, &tenantID, &recID, &channelStr, &templateKey, &variablesJSON, &subject, &body,
		&statusStr, &deliveryID, &deliveryError, &scheduledAt, &sentAt, &readAt, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("get notification: %w", err)
	}

	// Parse variables
	variables := make(map[string]string)
	if variablesJSON != nil {
		_ = json.Unmarshal(variablesJSON, &variables)
	}

	// Parse enums
	notificationID, _ := domain.NewNotificationID(notifID)
	recipientID, _ := domain.NewRecipientID(recID)
	channel, _ := domain.ChannelTypeFromString(channelStr)
	status, _ := domain.NotificationStatusFromString(statusStr)

	return domain.RehydrateNotification(
		notificationID,
		tenantID,
		recipientID,
		channel,
		templateKey,
		variables,
		subject,
		body,
		status,
		*deliveryID,
		*deliveryError,
		scheduledAt,
		sentAt,
		readAt,
		createdAt,
		updatedAt,
	), nil
}

func (r *NotificationRepository) ListByRecipient(ctx context.Context, tenantID domain.TenantID, recipientID domain.RecipientID, filters domain.ListFilters) ([]*domain.Notification, error) {
	var notifs []*domain.Notification
	var err error

	innerErr := sharedpostgres.WithTenantTx(ctx, r.pool, sharedpostgres.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		notifs, err = r.listByRecipientFromTx(ctx, tx, tenantID, recipientID, filters)
		return err
	})

	return notifs, innerErr
}

func (r *NotificationRepository) listByRecipientFromTx(ctx context.Context, tx pgx.Tx, tenantID domain.TenantID, recipientID domain.RecipientID, filters domain.ListFilters) ([]*domain.Notification, error) {
	query := `
		SELECT id, tenant_id, recipient_id, channel, template_key, variables, subject, body,
		       status, delivery_id, delivery_error, scheduled_at, sent_at, read_at, created_at, updated_at
		FROM notification.notifications
		WHERE tenant_id = $1 AND recipient_id = $2
	`

	args := []interface{}{tenantID.String(), recipientID.String()}
	argNum := 3

	// Add status filter if provided
	if filters.Status != domain.NotificationStatusUnknown {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filters.Status.String())
		argNum++
	}

	// Add unread filter if specified
	if filters.UnreadOnly {
		query += fmt.Sprintf(" AND read_at IS NULL AND status = $%d", argNum)
		args = append(args, domain.NotificationStatusSent.String())
		argNum++
	}

	query += " ORDER BY created_at DESC"

	// Add pagination
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filters.Limit)
		argNum++
	}
	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filters.Offset)
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifs []*domain.Notification
	for rows.Next() {
		var (
			notifID       string
			recID         string
			channelStr    string
			templateKey   string
			variablesJSON []byte
			subject       string
			body          string
			statusStr     string
			deliveryID    *string
			deliveryError *string
			scheduledAt   *time.Time
			sentAt        *time.Time
			readAt        *time.Time
			createdAt     time.Time
			updatedAt     time.Time
		)

		err := rows.Scan(
			&notifID, &tenantID, &recID, &channelStr, &templateKey, &variablesJSON, &subject, &body,
			&statusStr, &deliveryID, &deliveryError, &scheduledAt, &sentAt, &readAt, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}

		variables := make(map[string]string)
		if variablesJSON != nil {
			_ = json.Unmarshal(variablesJSON, &variables)
		}

		notificationID, _ := domain.NewNotificationID(notifID)
		recipID, _ := domain.NewRecipientID(recID)
		channel, _ := domain.ChannelTypeFromString(channelStr)
		status, _ := domain.NotificationStatusFromString(statusStr)

		notif := domain.RehydrateNotification(
			notificationID,
			tenantID,
			recipID,
			channel,
			templateKey,
			variables,
			subject,
			body,
			status,
			*deliveryID,
			*deliveryError,
			scheduledAt,
			sentAt,
			readAt,
			createdAt,
			updatedAt,
		)
		notifs = append(notifs, notif)
	}

	return notifs, rows.Err()
}

func (r *NotificationRepository) Update(ctx context.Context, notif *domain.Notification) error {
	return sharedpostgres.WithTenantTx(ctx, r.pool, sharedpostgres.TenantID(notif.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		query := `
			UPDATE notification.notifications
			SET status = $1, delivery_id = $2, delivery_error = $3,
			    sent_at = $4, read_at = $5, updated_at = $6
			WHERE id = $7 AND tenant_id = $8
		`

		_, err := tx.Exec(ctx, query,
			notif.Status().String(),
			notif.DeliveryID(),
			notif.DeliveryError(),
			notif.SentAt(),
			notif.ReadAt(),
			time.Now().UTC(),
			notif.ID().String(),
			notif.TenantID().String(),
		)

		return fmt.Errorf("update notification: %w", err)
	})
}
