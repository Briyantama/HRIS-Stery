package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
	"github.com/Briyantama/stery-kit/go/objstore"
	"github.com/google/uuid"
)

// UploadURLTTL is how long the presigned upload URL stays valid.
const UploadURLTTL = time.Hour

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

// UploadDocumentResult is the result of upload: the client PUTs the file to PresignedURL with Headers.
type UploadDocumentResult struct {
	DocumentID      string            `json:"file_id"`
	UploadSessionID string            `json:"upload_session_id"`
	PresignedURL    string            `json:"presigned_url"`
	Headers         map[string]string `json:"headers"`
	ExpiresAt       time.Time         `json:"expires_at"`
}

// UploadDocumentHandler implements the upload use case
type UploadDocumentHandler struct {
	docRepo domain.DocumentRepository
	store   objstore.Store
	bucket  string
}

func NewUploadDocumentHandler(docRepo domain.DocumentRepository, store objstore.Store, bucket string) *UploadDocumentHandler {
	return &UploadDocumentHandler{docRepo: docRepo, store: store, bucket: bucket}
}

// Handle records the document, then signs the upload. If signing fails the metadata row is left
// without an object; it stays UPLOADED and can be deleted like any other document.
func (h *UploadDocumentHandler) Handle(ctx context.Context, cmd UploadDocumentCommand) (*UploadDocumentResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)
	if _, err := uuid.Parse(cmd.EntityID); err != nil {
		return nil, fmt.Errorf("%w: entity_id must be a uuid", domain.ErrInvalidInput)
	}
	entityType := domain.EntityType(cmd.EntityType)
	docID := domain.NewDocumentID()
	objectKey := domain.ObjectKey(tenantID, entityType, cmd.EntityID, docID, 1)

	doc, err := domain.NewDocumentFile(
		docID, tenantID, cmd.UploadedBy, entityType, cmd.EntityID, cmd.FileName, cmd.MimeType,
		cmd.SizeBytes, h.bucket, objectKey, domain.FileClassification(cmd.Classification), cmd.RetentionDays,
	)
	if err != nil {
		return nil, err
	}
	if err = h.docRepo.CreateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}
	req, err := h.store.PresignPut(ctx, objectKey, cmd.MimeType, cmd.SizeBytes, UploadURLTTL)
	if err != nil {
		return nil, fmt.Errorf("presign upload: %w", err)
	}
	return &UploadDocumentResult{
		DocumentID:      docID.String(),
		UploadSessionID: "session-" + docID.String(),
		PresignedURL:    req.URL,
		Headers:         req.Headers,
		ExpiresAt:       req.ExpiresAt,
	}, nil
}
