package channels

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
)

// InAppAdapter persists notifications to the database for in-app delivery
type InAppAdapter struct{}

func NewInAppAdapter() domain.NotificationChannelAdapter {
	return &InAppAdapter{}
}

func (a *InAppAdapter) Send(ctx context.Context, notif *domain.Notification, rendered domain.RenderResult) (string, error) {
	// In-app notifications are already persisted in the database
	// This adapter just confirms delivery without sending to external service
	// The notification is retrieved via the REST API or gRPC when the user checks their inbox

	deliveryID := fmt.Sprintf("in-app-%s", notif.ID().String())
	return deliveryID, nil
}
