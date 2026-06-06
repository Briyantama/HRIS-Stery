package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/document-service/internal/domain"
	"go.uber.org/zap"
)

// PresignService handles presigned URL generation
type PresignService struct {
	minioClient *MinIOClient
	logger      *zap.Logger
}

func NewPresignService(minioClient *MinIOClient, logger *zap.Logger) *PresignService {
	return &PresignService{
		minioClient: minioClient,
		logger:      logger,
	}
}

// GenerateUploadURL creates a presigned URL for uploading a document
func (p *PresignService) GenerateUploadURL(ctx context.Context, tenantID domain.TenantID, bucket, objectKey string, expirationSeconds int32) (string, error) {
	if expirationSeconds <= 0 {
		expirationSeconds = 3600 // Default 1 hour
	}
	if expirationSeconds > 604800 {
		expirationSeconds = 604800 // Max 7 days
	}

	duration := time.Duration(expirationSeconds) * time.Second
	url, err := p.minioClient.PresignedPutObject(ctx, bucket, objectKey, duration)
	if err != nil {
		p.logger.Error("failed to generate upload URL",
			zap.Error(err),
			zap.String("bucket", bucket),
			zap.String("object_key", objectKey),
		)
		return "", fmt.Errorf("generate upload url: %w", err)
	}

	return url, nil
}

// GenerateDownloadURL creates a presigned URL for downloading a document
func (p *PresignService) GenerateDownloadURL(ctx context.Context, tenantID domain.TenantID, bucket, objectKey string, expirationSeconds int32) (string, error) {
	if expirationSeconds <= 0 {
		expirationSeconds = 3600 // Default 1 hour
	}
	if expirationSeconds > 604800 {
		expirationSeconds = 604800 // Max 7 days
	}

	duration := time.Duration(expirationSeconds) * time.Second
	url, err := p.minioClient.PresignedGetObject(ctx, bucket, objectKey, duration)
	if err != nil {
		p.logger.Error("failed to generate download URL",
			zap.Error(err),
			zap.String("bucket", bucket),
			zap.String("object_key", objectKey),
		)
		return "", fmt.Errorf("generate download url: %w", err)
	}

	return url, nil
}

// GenerateBucketName creates a bucket name for a tenant
func GenerateBucketName(tenantID domain.TenantID) string {
	return fmt.Sprintf("%s-documents", tenantID.String())
}

// GenerateObjectKey creates an object key for a document version
func GenerateObjectKey(entityType domain.EntityType, entityID, fileID string, versionNumber int32) string {
	return fmt.Sprintf("%s/%s/%d/%s", entityType, entityID, versionNumber, fileID)
}
