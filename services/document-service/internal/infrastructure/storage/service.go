package storage

import (
	"github.com/hris-stery/hris-stery/services/document-service/internal/domain"
	"go.uber.org/zap"
)

// DocumentStorageService coordinates document storage operations
type DocumentStorageService struct {
	minioClient *MinIOClient
	presigner   *PresignService
	logger      *zap.Logger
}

func NewDocumentStorageService(minioClient *MinIOClient, logger *zap.Logger) *DocumentStorageService {
	return &DocumentStorageService{
		minioClient: minioClient,
		presigner:   NewPresignService(minioClient, logger),
		logger:      logger,
	}
}

// GetMinIOClient returns the underlying MinIO client
func (s *DocumentStorageService) GetMinIOClient() *MinIOClient {
	return s.minioClient
}

// GetPresignService returns the presigner service
func (s *DocumentStorageService) GetPresignService() *PresignService {
	return s.presigner
}

// GetBucketName returns the bucket name for a tenant
func (s *DocumentStorageService) GetBucketName(tenantID domain.TenantID) string {
	return GenerateBucketName(tenantID)
}

// GetObjectKey returns the object key for a document
func (s *DocumentStorageService) GetObjectKey(entityType domain.EntityType, entityID, fileID string, versionNumber int32) string {
	return GenerateObjectKey(entityType, entityID, fileID, versionNumber)
}
