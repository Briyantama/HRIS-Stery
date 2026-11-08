package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
	"github.com/Briyantama/stery-kit/go/objstore"
)

const (
	defaultURLSeconds = 3600
	maxURLSeconds     = 604800 // 7 days, the SigV4 presign ceiling
)

// GetDownloadURLQuery asks for a time-limited download URL. VersionNumber 0 means the current version.
type GetDownloadURLQuery struct {
	TenantID      string
	DocumentID    string
	VersionNumber int32
	ExpirySeconds int32 // 0 = one hour
}

// DownloadURLResult is the presigned URL and when it stops working.
type DownloadURLResult struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetDownloadURLHandler signs a GET for one version of a document.
type GetDownloadURLHandler struct {
	docRepo     domain.DocumentRepository
	versionRepo domain.DocumentVersionRepository
	store       objstore.Store
}

func NewGetDownloadURLHandler(docRepo domain.DocumentRepository, versionRepo domain.DocumentVersionRepository, store objstore.Store) *GetDownloadURLHandler {
	return &GetDownloadURLHandler{docRepo: docRepo, versionRepo: versionRepo, store: store}
}

// Handle executes the query; a soft-deleted document reads as not found.
func (h *GetDownloadURLHandler) Handle(ctx context.Context, q GetDownloadURLQuery) (*DownloadURLResult, error) {
	if q.ExpirySeconds < 0 {
		return nil, fmt.Errorf("%w: expiration_seconds cannot be negative", domain.ErrInvalidInput)
	}
	secs := q.ExpirySeconds
	if secs == 0 {
		secs = defaultURLSeconds
	}
	if secs > maxURLSeconds {
		secs = maxURLSeconds
	}
	tenantID := domain.MustNewTenantID(q.TenantID)
	docID, err := domain.NewDocumentIDFromString(q.DocumentID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}
	doc, err := h.docRepo.GetDocumentByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if doc.IsDeleted() {
		return nil, domain.ErrDocumentNotFound
	}
	number := q.VersionNumber
	if number == 0 {
		number = doc.CurrentVersion()
	}
	version, err := h.versionRepo.GetVersionByNumber(ctx, docID, number)
	if err != nil {
		return nil, err
	}
	ttl := time.Duration(secs) * time.Second
	url, err := h.store.PresignGet(ctx, version.ObjectKey(), ttl)
	if err != nil {
		return nil, fmt.Errorf("presign download: %w", err)
	}
	return &DownloadURLResult{URL: url, ExpiresAt: time.Now().Add(ttl)}, nil
}
