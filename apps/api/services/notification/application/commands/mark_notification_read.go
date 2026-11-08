package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
)

type MarkNotificationReadCommand struct {
	TenantID        string
	RecipientID     string   // when set, only this recipient's notifications are marked
	NotificationIDs []string // Batch support
}

type MarkNotificationReadResult struct {
	UpdatedCount int
}

type MarkNotificationReadHandler interface {
	Handle(ctx context.Context, cmd MarkNotificationReadCommand) (*MarkNotificationReadResult, error)
}

type MarkNotificationReadHandlerImpl struct {
	notificationRepo domain.NotificationRepository
}

func NewMarkNotificationReadHandler(notificationRepo domain.NotificationRepository) MarkNotificationReadHandler {
	return &MarkNotificationReadHandlerImpl{
		notificationRepo: notificationRepo,
	}
}

func (h *MarkNotificationReadHandlerImpl) Handle(ctx context.Context, cmd MarkNotificationReadCommand) (*MarkNotificationReadResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)

	updatedCount := 0

	// Handle batch: if no notification IDs provided, mark all as read for tenant (future enhancement)
	if len(cmd.NotificationIDs) == 0 {
		// For now, return success with count 0
		return &MarkNotificationReadResult{UpdatedCount: updatedCount}, nil
	}

	// Process each notification
	for _, notifID := range cmd.NotificationIDs {
		notificationID, err := domain.NewNotificationID(notifID)
		if err != nil {
			// Skip invalid IDs, continue processing others
			continue
		}

		// Fetch notification
		notif, err := h.notificationRepo.GetByID(ctx, tenantID, notificationID)
		if errors.Is(err, domain.ErrNotificationNotFound) || (err == nil && notif == nil) {
			continue // not found (or not visible to this tenant): skip, keep processing others
		}
		if err != nil {
			return nil, fmt.Errorf("mark notification read: fetch: %w", err)
		}
		if cmd.RecipientID != "" && notif.RecipientID().String() != cmd.RecipientID {
			continue
		}

		// Transition to READ (only if not already read)
		if notif.Status() != domain.NotificationStatusRead {
			if err := notif.MarkRead(); err != nil {
				// Skip on transition error, continue processing others
				continue
			}

			// Persist
			if err := h.notificationRepo.Update(ctx, notif); err != nil {
				if errors.Is(err, domain.ErrNotificationNotFound) {
					continue
				}
				return nil, fmt.Errorf("mark notification read: persist: %w", err)
			}

			updatedCount++
		}
	}

	return &MarkNotificationReadResult{
		UpdatedCount: updatedCount,
	}, nil
}
