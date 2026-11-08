package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
	"github.com/Briyantama/stery-kit/go/pgtenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const fileColumns = `file_id, tenant_id, uploaded_by, entity_type, entity_id, file_name, mime_type, size_bytes, bucket, object_key, status, classification, retention_days, current_version, total_versions, created_at, updated_at`

// DocumentRepository implements the DocumentRepository port
type DocumentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{pool: pool}
}

// CreateDocument stores a new document file and its version 1 row in one transaction.
func (r *DocumentRepository) CreateDocument(ctx context.Context, doc *domain.DocumentFile) error {
	v1, err := domain.NewDocumentVersion(doc.ID(), 1, doc.ObjectKey(), doc.SizeBytes(), doc.UploadedBy())
	if err != nil {
		return err
	}
	return pgtenant.WithTenantTx(ctx, r.pool, doc.TenantID().String(), "", func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO document.files (`+fileColumns+`) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
			doc.ID().String(), doc.TenantID().String(), doc.UploadedBy(), string(doc.EntityType()), doc.EntityID(),
			doc.FileName(), doc.MimeType(), doc.SizeBytes(), doc.Bucket(), doc.ObjectKey(),
			string(doc.Status()), string(doc.Classification()), doc.RetentionDays(), doc.CurrentVersion(), doc.TotalVersions(),
			doc.CreatedAt(), doc.UpdatedAt(),
		); err != nil {
			return fmt.Errorf("insert document: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO document.file_versions (version_id, file_id, version_number, object_key, size_bytes, created_by, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			v1.ID().String(), v1.FileID().String(), v1.VersionNumber(), v1.ObjectKey(), v1.SizeBytes(), v1.CreatedBy(), v1.CreatedAt(),
		); err != nil {
			return fmt.Errorf("insert first version: %w", err)
		}
		return nil
	})
}

// GetDocumentByID retrieves a document by ID
func (r *DocumentRepository) GetDocumentByID(ctx context.Context, tenantID domain.TenantID, docID domain.DocumentID) (*domain.DocumentFile, error) {
	var doc *domain.DocumentFile
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		var err error
		doc, err = scanDocument(tx.QueryRow(ctx, `SELECT `+fileColumns+` FROM document.files WHERE file_id = $1 AND tenant_id = $2`, docID.String(), tenantID.String()))
		return err
	})
	if err != nil {
		if errors.Is(err, domain.ErrDocumentNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get document by id: %w", err)
	}
	return doc, nil
}

// GetDocumentsByEntity retrieves the documents of an entity that the viewer may see, with pagination
func (r *DocumentRepository) GetDocumentsByEntity(ctx context.Context, tenantID domain.TenantID, viewer domain.Viewer, entityType domain.EntityType, entityID string, limit, offset int) ([]*domain.DocumentFile, int, error) {
	where, args := restrictTo(viewer, `tenant_id = $1 AND entity_type = $2 AND entity_id = $3`, []any{tenantID.String(), string(entityType), entityID})
	return r.list(ctx, tenantID, where, args, limit, offset)
}

// SearchDocuments searches the documents the viewer may see by file name, with pagination
func (r *DocumentRepository) SearchDocuments(ctx context.Context, tenantID domain.TenantID, viewer domain.Viewer, query string, limit, offset int) ([]*domain.DocumentFile, int, error) {
	where, args := restrictTo(viewer, `tenant_id = $1 AND file_name ILIKE $2`, []any{tenantID.String(), "%" + query + "%"})
	return r.list(ctx, tenantID, where, args, limit, offset)
}

// restrictTo narrows a list query to what the viewer may see; it is domain.Viewer.CanView in SQL.
func restrictTo(v domain.Viewer, where string, args []any) (string, []any) {
	if v.HR {
		return where, args
	}
	args = append(args, strings.ToLower(v.UserID))
	n := len(args)
	return fmt.Sprintf(`%s AND classification <> 'SECRET' AND (uploaded_by::text = $%d OR (entity_type = 'EMPLOYEE' AND entity_id::text = $%d))`, where, n, n), args
}

func (r *DocumentRepository) list(ctx context.Context, tenantID domain.TenantID, where string, args []any, limit, offset int) ([]*domain.DocumentFile, int, error) {
	var docs []*domain.DocumentFile
	var total int
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM document.files WHERE `+where, args...).Scan(&total); err != nil {
			return fmt.Errorf("count documents: %w", err)
		}
		n := len(args)
		rows, err := tx.Query(ctx,
			fmt.Sprintf(`SELECT %s FROM document.files WHERE %s ORDER BY created_at DESC, file_id LIMIT $%d OFFSET $%d`, fileColumns, where, n+1, n+2),
			append(args, limit, offset)...)
		if err != nil {
			return fmt.Errorf("query documents: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			doc, err := scanDocument(rows)
			if err != nil {
				return err
			}
			docs = append(docs, doc)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list documents: %w", err)
	}
	return docs, total, nil
}

func scanDocument(row pgx.Row) (*domain.DocumentFile, error) {
	var (
		fileID, tenID, uploadedBy, entityType, entityID, fileName, mimeType, bucket, objectKey, status, classification string
		sizeBytes                                                                                                      int64
		retentionDays, currentVersion, totalVersions                                                                   int32
		createdAt, updatedAt                                                                                           time.Time
	)
	if err := row.Scan(&fileID, &tenID, &uploadedBy, &entityType, &entityID, &fileName, &mimeType, &sizeBytes, &bucket, &objectKey,
		&status, &classification, &retentionDays, &currentVersion, &totalVersions, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDocumentNotFound
		}
		return nil, fmt.Errorf("scan document: %w", err)
	}
	id, err := domain.NewDocumentIDFromString(fileID)
	if err != nil {
		return nil, err
	}
	doc, err := domain.RehydrateDocumentFile(id, domain.MustNewTenantID(tenID), uploadedBy, domain.EntityType(entityType), entityID,
		fileName, mimeType, sizeBytes, bucket, objectKey, domain.FileStatus(status), domain.FileClassification(classification),
		retentionDays, currentVersion, totalVersions, createdAt, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("rehydrate document: %w", err)
	}
	return doc, nil
}

// UpdateDocumentStatus updates file status
func (r *DocumentRepository) UpdateDocumentStatus(ctx context.Context, tenantID domain.TenantID, docID domain.DocumentID, status domain.FileStatus) error {
	return r.exec(ctx, tenantID, "update document status",
		`UPDATE document.files SET status = $1, updated_at = NOW() WHERE file_id = $2 AND tenant_id = $3`,
		string(status), docID.String(), tenantID.String())
}

// SetVersions persists the version counters
func (r *DocumentRepository) SetVersions(ctx context.Context, tenantID domain.TenantID, docID domain.DocumentID, current, total int32) error {
	return r.exec(ctx, tenantID, "update version counters",
		`UPDATE document.files SET current_version = $1, total_versions = $2, updated_at = NOW() WHERE file_id = $3 AND tenant_id = $4`,
		current, total, docID.String(), tenantID.String())
}

// DeleteDocument soft-deletes a document
func (r *DocumentRepository) DeleteDocument(ctx context.Context, tenantID domain.TenantID, docID domain.DocumentID) error {
	return r.UpdateDocumentStatus(ctx, tenantID, docID, domain.StatusDeleted)
}

func (r *DocumentRepository) exec(ctx context.Context, tenantID domain.TenantID, what, sql string, args ...any) error {
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrDocumentNotFound
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrDocumentNotFound) {
			return err
		}
		return fmt.Errorf("%s: %w", what, err)
	}
	return nil
}
