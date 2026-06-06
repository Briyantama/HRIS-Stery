package postgres

import (
	"context"
	"fmt"
	"time"

	shared "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/document-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DocumentRepository implements the DocumentRepository port
type DocumentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{pool: pool}
}

// CreateDocument stores a new document file
func (r *DocumentRepository) CreateDocument(ctx context.Context, doc *domain.DocumentFile) error {
	query := `
		INSERT INTO document.files (
			file_id, tenant_id, uploaded_by, entity_type, entity_id,
			file_name, mime_type, size_bytes, minio_bucket, minio_key,
			status, classification, retention_days, current_version, total_versions,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	return shared.WithTenantTx(ctx, r.pool, shared.TenantID(doc.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			doc.ID().String(),
			doc.TenantID().String(),
			doc.UploadedBy(),
			string(doc.EntityType()),
			doc.EntityID(),
			doc.FileName(),
			doc.MimeType(),
			doc.SizeBytes(),
			doc.MinioBucket(),
			doc.MinioKey(),
			string(doc.Status()),
			string(doc.Classification()),
			doc.RetentionDays(),
			doc.CurrentVersion(),
			doc.TotalVersions(),
			doc.CreatedAt(),
			doc.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert document: %w", err)
		}
		return nil
	})
}

// GetDocumentByID retrieves a document by ID
func (r *DocumentRepository) GetDocumentByID(ctx context.Context, tenantID domain.TenantID, docID domain.DocumentID) (*domain.DocumentFile, error) {
	var doc *domain.DocumentFile

	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		row := tx.QueryRow(ctx,
			`SELECT file_id, tenant_id, uploaded_by, entity_type, entity_id, file_name, mime_type, size_bytes, minio_bucket, minio_key, status, classification, retention_days, current_version, total_versions, created_at, updated_at FROM document.files WHERE file_id = $1 AND tenant_id = $2`,
			docID.String(),
			tenantID.String(),
		)

		var (
			fileID, tenID, uploadedBy, entityType, entityID, fileName, mimeType, minioBucket, minioKey, status, classification string
			sizeBytes                                                                                                          int64
			retentionDays, currentVersion, totalVersions                                                                       int32
			createdAt, updatedAt                                                                                               time.Time
		)

		if scanErr := row.Scan(
			&fileID, &tenID, &uploadedBy, &entityType, &entityID,
			&fileName, &mimeType, &sizeBytes, &minioBucket, &minioKey,
			&status, &classification, &retentionDays, &currentVersion, &totalVersions,
			&createdAt, &updatedAt,
		); scanErr != nil {
			if scanErr == pgx.ErrNoRows {
				return domain.ErrDocumentNotFound
			}
			return fmt.Errorf("scan document: %w", scanErr)
		}

		parsedDoc, rehydrateErr := domain.RehydrateDocumentFile(
			domain.MustNewDocumentID(fileID),
			domain.MustNewTenantID(tenID),
			uploadedBy,
			domain.EntityType(entityType),
			entityID,
			fileName,
			mimeType,
			sizeBytes,
			minioBucket,
			minioKey,
			domain.FileStatus(status),
			domain.FileClassification(classification),
			retentionDays,
			currentVersion,
			totalVersions,
			createdAt,
			updatedAt,
		)
		if rehydrateErr != nil {
			return fmt.Errorf("rehydrate document: %w", rehydrateErr)
		}

		doc = parsedDoc
		return nil
	})

	if err != nil {
		if err == domain.ErrDocumentNotFound {
			return nil, domain.ErrDocumentNotFound
		}
		return nil, fmt.Errorf("get document by id: %w", err)
	}

	return doc, nil
}

// GetDocumentsByEntity retrieves documents for an entity with pagination
func (r *DocumentRepository) GetDocumentsByEntity(ctx context.Context, tenantID domain.TenantID, entityType domain.EntityType, entityID string, limit, offset int) ([]*domain.DocumentFile, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var docs []*domain.DocumentFile
	var total int

	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		// Get total count
		countRow := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM document.files WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3`,
			tenantID.String(),
			string(entityType),
			entityID,
		)
		if err := countRow.Scan(&total); err != nil {
			return fmt.Errorf("count documents: %w", err)
		}

		// Get paginated results
		rows, err := tx.Query(ctx,
			`SELECT file_id, tenant_id, uploaded_by, entity_type, entity_id, file_name, mime_type, size_bytes, minio_bucket, minio_key, status, classification, retention_days, current_version, total_versions, created_at, updated_at FROM document.files WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3 ORDER BY created_at DESC LIMIT $4 OFFSET $5`,
			tenantID.String(),
			string(entityType),
			entityID,
			limit,
			offset,
		)
		if err != nil {
			return fmt.Errorf("query documents: %w", err)
		}
		defer rows.Close()

		docs, err = r.scanDocuments(rows, tenantID)
		return err
	})

	if err != nil {
		return nil, 0, fmt.Errorf("get documents by entity: %w", err)
	}

	return docs, total, nil
}

// SearchDocuments searches documents by file name with pagination
func (r *DocumentRepository) SearchDocuments(ctx context.Context, tenantID domain.TenantID, query string, limit, offset int) ([]*domain.DocumentFile, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var docs []*domain.DocumentFile
	var total int

	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		searchQuery := "%" + query + "%"

		// Get total count
		countRow := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM document.files WHERE tenant_id = $1 AND file_name ILIKE $2`,
			tenantID.String(),
			searchQuery,
		)
		if err := countRow.Scan(&total); err != nil {
			return fmt.Errorf("count documents: %w", err)
		}

		// Get paginated results
		rows, err := tx.Query(ctx,
			`SELECT file_id, tenant_id, uploaded_by, entity_type, entity_id, file_name, mime_type, size_bytes, minio_bucket, minio_key, status, classification, retention_days, current_version, total_versions, created_at, updated_at FROM document.files WHERE tenant_id = $1 AND file_name ILIKE $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
			tenantID.String(),
			searchQuery,
			limit,
			offset,
		)
		if err != nil {
			return fmt.Errorf("search documents: %w", err)
		}
		defer rows.Close()

		docs, err = r.scanDocuments(rows, tenantID)
		return err
	})

	if err != nil {
		return nil, 0, fmt.Errorf("search documents: %w", err)
	}

	return docs, total, nil
}

// scanDocuments is a helper to scan document rows
func (r *DocumentRepository) scanDocuments(rows pgx.Rows, expectedTenantID domain.TenantID) ([]*domain.DocumentFile, error) {
	var docs []*domain.DocumentFile

	for rows.Next() {
		var (
			fileID, tenID, uploadedBy, entityType, entityID, fileName, mimeType, minioBucket, minioKey, status, classification string
			sizeBytes                                                                                                          int64
			retentionDays, currentVersion, totalVersions                                                                       int32
			createdAt, updatedAt                                                                                               time.Time
		)

		if err := rows.Scan(
			&fileID, &tenID, &uploadedBy, &entityType, &entityID,
			&fileName, &mimeType, &sizeBytes, &minioBucket, &minioKey,
			&status, &classification, &retentionDays, &currentVersion, &totalVersions,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		doc, err := domain.RehydrateDocumentFile(
			domain.MustNewDocumentID(fileID),
			domain.MustNewTenantID(tenID),
			uploadedBy,
			domain.EntityType(entityType),
			entityID,
			fileName,
			mimeType,
			sizeBytes,
			minioBucket,
			minioKey,
			domain.FileStatus(status),
			domain.FileClassification(classification),
			retentionDays,
			currentVersion,
			totalVersions,
			createdAt,
			updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("rehydrate document: %w", err)
		}

		docs = append(docs, doc)
	}

	return docs, rows.Err()
}

// UpdateDocumentStatus updates file status
func (r *DocumentRepository) UpdateDocumentStatus(ctx context.Context, tenantID domain.TenantID, docID domain.DocumentID, status domain.FileStatus) error {
	err := shared.WithTenantTx(ctx, r.pool, shared.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		_, execErr := tx.Exec(ctx,
			`UPDATE document.files SET status = $1, updated_at = NOW() WHERE file_id = $2 AND tenant_id = $3`,
			string(status),
			docID.String(),
			tenantID.String(),
		)
		if execErr != nil {
			return fmt.Errorf("update document status: %w", execErr)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update document status: %w", err)
	}
	return nil
}

// DeleteDocument soft-deletes a document
func (r *DocumentRepository) DeleteDocument(ctx context.Context, tenantID domain.TenantID, docID domain.DocumentID) error {
	return r.UpdateDocumentStatus(ctx, tenantID, docID, domain.StatusDeleted)
}
