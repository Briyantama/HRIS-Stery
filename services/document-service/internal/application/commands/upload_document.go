package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/document-service/internal/domain"
	"github.com/hris-stery/hris-stery/services/document-service/internal/infrastructure/storage"
	"go.uber.org/zap"
)

// UploadDocumentCommand captures upload request
type UploadDocumentCommand struct {
	TenantID       string
	UploadedBy     string
	EntityType     string // EMPLOYEE, LEAVE, GENERAL
	EntityID       string
	FileName       string
	MimeType       string
	SizeBytes      int64
	Classification string // PUBLIC, CONFIDENTIAL, SECRET
	RetentionDays  int32
}

// UploadDocumentResult is the result of upload
type UploadDocumentResult struct {
	DocumentID      string
	UploadSessionID string
	PresignedURL    string
	ExpiresAt       string // ISO 8601 timestamp
}

// UploadDocumentHandler implements the upload use case
type UploadDocumentHandler struct {
	docRepo        domain.DocumentRepository
	storageService *storage.DocumentStorageService
	logger         *zap.Logger
}

func NewUploadDocumentHandler(
	docRepo domain.DocumentRepository,
	storageService *storage.DocumentStorageService,
	logger *zap.Logger,
) *UploadDocumentHandler {
	return &UploadDocumentHandler{
		docRepo:        docRepo,
		storageService: storageService,
		logger:         logger,
	}
}

// Handle executes the upload command
func (h *UploadDocumentHandler) Handle(ctx context.Context, cmd UploadDocumentCommand) (*UploadDocumentResult, error) {
	// Map strings to domain types
	tenantID := domain.MustNewTenantID(cmd.TenantID)
	entityType := domain.EntityType(cmd.EntityType)
	classification := domain.FileClassification(cmd.Classification)

	// Generate MinIO storage location
	bucketName := h.storageService.GetBucketName(tenantID)

	// Create a temporary document ID to generate the key
	tempDocID := domain.NewDocumentID()
	objectKey := h.storageService.GetObjectKey(entityType, cmd.EntityID, tempDocID.String(), 1)

	// Create domain aggregate
	doc, err := domain.NewDocumentFile(
		tenantID,
		cmd.UploadedBy,
		entityType,
		cmd.EntityID,
		cmd.FileName,
		cmd.MimeType,
		cmd.SizeBytes,
		bucketName,
		objectKey,
		classification,
		cmd.RetentionDays,
	)
	if err != nil {
		h.logger.Warn("invalid upload parameters", zap.Error(err), zap.String("tenant_id", cmd.TenantID))
		return nil, fmt.Errorf("invalid upload parameters: %w", err)
	}

	// Update the object key with the actual document ID
	actualObjectKey := h.storageService.GetObjectKey(entityType, cmd.EntityID, doc.ID().String(), 1)

	// Create domain aggregate with actual key
	doc, err = domain.NewDocumentFile(
		tenantID,
		cmd.UploadedBy,
		entityType,
		cmd.EntityID,
		cmd.FileName,
		cmd.MimeType,
		cmd.SizeBytes,
		bucketName,
		actualObjectKey,
		classification,
		cmd.RetentionDays,
	)
	if err != nil {
		h.logger.Warn("invalid upload parameters", zap.Error(err), zap.String("tenant_id", cmd.TenantID))
		return nil, fmt.Errorf("invalid upload parameters: %w", err)
	}

	// Persist to repository
	if err = h.docRepo.CreateDocument(ctx, doc); err != nil {
		h.logger.Error("failed to create document", zap.Error(err), zap.String("tenant_id", cmd.TenantID))
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Generate presigned upload URL
	presigner := h.storageService.GetPresignService()
	presignedURL, err := presigner.GenerateUploadURL(ctx, tenantID, bucketName, actualObjectKey, 3600)
	if err != nil {
		h.logger.Error("failed to generate presigned URL", zap.Error(err))
		presignedURL = "" // Continue without presigned URL
	}

	h.logger.Info("document uploaded successfully",
		zap.String("document_id", doc.ID().String()),
		zap.String("tenant_id", cmd.TenantID),
		zap.String("file_name", cmd.FileName),
	)

	return &UploadDocumentResult{
		DocumentID:      doc.ID().String(),
		UploadSessionID: fmt.Sprintf("session-%s", doc.ID().String()),
		PresignedURL:    presignedURL,
		ExpiresAt:       "", // Set by presigner
	}, nil
}
