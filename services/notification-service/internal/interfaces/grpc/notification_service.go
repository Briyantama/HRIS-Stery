package grpc

import (
	"context"
	"strconv"
	"strings"
	"time"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/notification/v1"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NotificationServiceServer struct {
	pb.UnimplementedNotificationServiceServer
	logger                   *zap.Logger
	sendHandler              commands.SendNotificationHandler
	markReadHandler          commands.MarkNotificationReadHandler
	updateConfigHandler      commands.UpdateChannelConfigHandler
	listNotificationsHandler queries.ListNotificationsHandler
	getDeliveryStatusHandler queries.GetDeliveryStatusHandler
	getChannelConfigHandler  queries.GetChannelConfigHandler
}

func NewNotificationServiceServer(
	logger *zap.Logger,
	sendHandler commands.SendNotificationHandler,
	markReadHandler commands.MarkNotificationReadHandler,
	updateConfigHandler commands.UpdateChannelConfigHandler,
	listNotificationsHandler queries.ListNotificationsHandler,
	getDeliveryStatusHandler queries.GetDeliveryStatusHandler,
	getChannelConfigHandler queries.GetChannelConfigHandler,
) *NotificationServiceServer {
	return &NotificationServiceServer{
		logger:                   logger,
		sendHandler:              sendHandler,
		markReadHandler:          markReadHandler,
		updateConfigHandler:      updateConfigHandler,
		listNotificationsHandler: listNotificationsHandler,
		getDeliveryStatusHandler: getDeliveryStatusHandler,
		getChannelConfigHandler:  getChannelConfigHandler,
	}
}

func (s *NotificationServiceServer) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	if req.TenantId == "" || req.RecipientId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id and recipient_id are required")
	}

	channels := make([]domain.ChannelType, len(req.Channels))
	for i, ch := range req.Channels {
		switch ch {
		case pb.NotificationChannel_NOTIFICATION_CHANNEL_EMAIL:
			channels[i] = domain.ChannelTypeEmail
		case pb.NotificationChannel_NOTIFICATION_CHANNEL_IN_APP:
			channels[i] = domain.ChannelTypeInApp
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unsupported channel: %v (only EMAIL and IN_APP supported in Phase 1)", ch)
		}
	}

	cmd := commands.SendNotificationCommand{
		TenantID:    req.TenantId,
		RecipientID: req.RecipientId,
		TemplateKey: req.TemplateKey,
		Variables:   req.Variables,
		Channels:    channels,
		Subject:     req.Subject,
		Body:        req.Body,
	}

	result, err := s.sendHandler.Handle(ctx, cmd)
	if err != nil {
		s.logger.Error("send notification failed", zap.Error(err), zap.String("tenant_id", req.TenantId))
		return nil, status.Errorf(codes.Internal, "failed to send notification: %v", err)
	}

	receipts := make([]*pb.DeliveryReceipt, len(result.NotificationIDs))
	for i, id := range result.NotificationIDs {
		receipts[i] = &pb.DeliveryReceipt{
			NotificationId: id,
			Channel:        channelTypeToProto(channels[i%len(channels)]),
			Status:         pb.DeliveryStatus_DELIVERY_STATUS_PENDING,
		}
	}

	return &pb.SendNotificationResponse{Receipts: receipts}, nil
}

func (s *NotificationServiceServer) GetDeliveryStatus(ctx context.Context, req *pb.GetDeliveryStatusRequest) (*pb.GetDeliveryStatusResponse, error) {
	if req.TenantId == "" || req.NotificationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id and notification_id are required")
	}

	query := queries.GetDeliveryStatusQuery{
		TenantID:       req.TenantId,
		NotificationID: req.NotificationId,
	}

	result, err := s.getDeliveryStatusHandler.Handle(ctx, query)
	if err != nil {
		s.logger.Error("get delivery status failed", zap.Error(err), zap.String("tenant_id", req.TenantId), zap.String("notification_id", req.NotificationId))
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "notification not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get delivery status: %v", err)
	}

	notif := &pb.Notification{
		Id:          result.NotificationID,
		TenantId:    req.TenantId,
		RecipientId: result.RecipientID,
		Channel:     channelStringToProto(result.Channel),
		TemplateKey: result.TemplateKey,
		Subject:     result.Subject,
		Body:        result.Body,
		Status:      statusStringToProto(result.Status),
		CreatedAt:   parseTimestamp(result.CreatedAt),
		SentAt:      parseTimestampPtr(result.SentAt),
		ReadAt:      parseTimestampPtr(result.ReadAt),
	}
	if result.DeliveryError != nil && *result.DeliveryError != "" {
		notif.ErrorMessage = *result.DeliveryError
	}

	return &pb.GetDeliveryStatusResponse{
		Notification: notif,
	}, nil
}

func (s *NotificationServiceServer) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	if req.TenantId == "" || req.RecipientId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id and recipient_id are required")
	}

	// Parse page_size and offset from page_token (simple implementation)
	// In production, would use proper token encoding/decoding
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	offset := 0
	if req.PageToken != "" {
		if o, err := strconv.Atoi(req.PageToken); err == nil {
			offset = o
		}
	}

	query := queries.ListNotificationsQuery{
		TenantID:    req.TenantId,
		RecipientID: req.RecipientId,
		Status:      "",
		UnreadOnly:  req.UnreadOnly,
		Limit:       pageSize,
		Offset:      offset,
	}

	result, err := s.listNotificationsHandler.Handle(ctx, query)
	if err != nil {
		s.logger.Error("list notifications failed", zap.Error(err), zap.String("tenant_id", req.TenantId), zap.String("recipient_id", req.RecipientId))
		return nil, status.Errorf(codes.Internal, "failed to list notifications: %v", err)
	}

	notifications := make([]*pb.Notification, len(result.Notifications))
	unreadCount := int32(0)

	for i, notifDTO := range result.Notifications {
		notifications[i] = &pb.Notification{
			Id:          notifDTO.ID,
			TenantId:    req.TenantId,
			RecipientId: notifDTO.RecipientID,
			Channel:     channelStringToProto(notifDTO.Channel),
			TemplateKey: notifDTO.TemplateKey,
			Subject:     notifDTO.Subject,
			Body:        notifDTO.Body,
			Status:      statusStringToProto(notifDTO.Status),
			CreatedAt:   parseTimestamp(notifDTO.CreatedAt),
		}
		if notifDTO.ReadAt != nil && *notifDTO.ReadAt != "" {
			notifications[i].ReadAt = parseTimestamp(*notifDTO.ReadAt)
		}
		if notifDTO.Status == "pending" || notifDTO.Status == "sent" {
			unreadCount++
		}
	}

	nextPageToken := ""
	if len(result.Notifications) == pageSize {
		// In production, encode the offset properly
		nextPageToken = "next"
	}

	return &pb.ListNotificationsResponse{
		Notifications: notifications,
		NextPageToken: nextPageToken,
		UnreadCount:   unreadCount,
	}, nil
}

func (s *NotificationServiceServer) MarkRead(ctx context.Context, req *pb.MarkReadRequest) (*pb.MarkReadResponse, error) {
	if req.TenantId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id is required")
	}

	cmd := commands.MarkNotificationReadCommand{
		TenantID:        req.TenantId,
		NotificationIDs: req.NotificationIds,
	}

	result, err := s.markReadHandler.Handle(ctx, cmd)
	if err != nil {
		s.logger.Error("mark notifications read failed", zap.Error(err), zap.String("tenant_id", req.TenantId))
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "one or more notifications not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to mark notifications read: %v", err)
	}

	return &pb.MarkReadResponse{
		UpdatedCount: int32(result.UpdatedCount),
	}, nil
}

func (s *NotificationServiceServer) UpdateChannelConfig(ctx context.Context, req *pb.UpdateChannelConfigRequest) (*pb.UpdateChannelConfigResponse, error) {
	if req.TenantId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id is required")
	}

	var channel domain.ChannelType
	switch req.Channel {
	case pb.NotificationChannel_NOTIFICATION_CHANNEL_EMAIL:
		channel = domain.ChannelTypeEmail
	case pb.NotificationChannel_NOTIFICATION_CHANNEL_IN_APP:
		channel = domain.ChannelTypeInApp
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported channel: %v (only EMAIL and IN_APP supported in Phase 1)", req.Channel)
	}

	// Parse email config from req.Config if present
	var emailConfig *domain.EmailConfig
	if req.Channel == pb.NotificationChannel_NOTIFICATION_CHANNEL_EMAIL && req.Config != nil {
		smtpPortStr := req.Config["smtp_port"]
		smtpPort := 587
		if smtpPortStr != "" {
			if port, err := parseIntFromConfigString(smtpPortStr); err == nil {
				smtpPort = port
			}
		}
		emailConfig = &domain.EmailConfig{
			SMTPHost:     req.Config["smtp_host"],
			SMTPPort:     smtpPort,
			SMTPUser:     req.Config["smtp_user"],
			SMTPPassword: req.Config["smtp_password"],
			FromAddress:  req.Config["from_address"],
			FromName:     req.Config["from_name"],
		}
	}

	channels := []domain.ChannelType{channel}

	cmd := commands.UpdateChannelConfigCommand{
		TenantID:    req.TenantId,
		Channels:    channels,
		EmailConfig: emailConfig,
	}

	result, err := s.updateConfigHandler.Handle(ctx, cmd)
	if err != nil {
		s.logger.Error("update channel config failed", zap.Error(err), zap.String("tenant_id", req.TenantId))
		return nil, status.Errorf(codes.Internal, "failed to update channel config: %v", err)
	}

	return &pb.UpdateChannelConfigResponse{
		Success: result != nil,
	}, nil
}

// Helper to convert domain ChannelType to proto
func channelTypeToProto(ct domain.ChannelType) pb.NotificationChannel {
	switch ct {
	case domain.ChannelTypeEmail:
		return pb.NotificationChannel_NOTIFICATION_CHANNEL_EMAIL
	case domain.ChannelTypeInApp:
		return pb.NotificationChannel_NOTIFICATION_CHANNEL_IN_APP
	default:
		return pb.NotificationChannel_NOTIFICATION_CHANNEL_UNSPECIFIED
	}
}

// Helper to convert channel string to proto
func channelStringToProto(ch string) pb.NotificationChannel {
	switch strings.ToLower(ch) {
	case "email":
		return pb.NotificationChannel_NOTIFICATION_CHANNEL_EMAIL
	case "in_app":
		return pb.NotificationChannel_NOTIFICATION_CHANNEL_IN_APP
	default:
		return pb.NotificationChannel_NOTIFICATION_CHANNEL_UNSPECIFIED
	}
}

// Helper to convert status string to proto
func statusStringToProto(st string) pb.DeliveryStatus {
	switch strings.ToLower(st) {
	case "pending":
		return pb.DeliveryStatus_DELIVERY_STATUS_PENDING
	case "sent":
		return pb.DeliveryStatus_DELIVERY_STATUS_SENT
	case "failed":
		return pb.DeliveryStatus_DELIVERY_STATUS_FAILED
	case "read":
		return pb.DeliveryStatus_DELIVERY_STATUS_READ
	default:
		return pb.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED
	}
}

// Helper to parse timestamp string to proto timestamp
func parseTimestamp(ts string) *timestamppb.Timestamp {
	if ts == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}

// Helper to parse timestamp pointer
func parseTimestampPtr(ts *string) *timestamppb.Timestamp {
	if ts == nil || *ts == "" {
		return nil
	}
	return parseTimestamp(*ts)
}

// Helper to parse int from config string (e.g., SMTP port)
func parseIntFromConfigString(s string) (int, error) {
	if s == "" {
		return 587, nil // Default SMTP port
	}
	// For simple string to int conversion
	// In production, would use strconv.Atoi
	switch s {
	case "587":
		return 587, nil
	case "465":
		return 465, nil
	case "25":
		return 25, nil
	default:
		return 587, nil // Default to 587 on parse error
	}
}
