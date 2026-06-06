package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/document-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DocumentVersionRepository implements DocumentVersionRepository port
type DocumentVersionRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentVersionRepository(pool *pgxpool.Pool) *DocumentVersionRepository {
	return &DocumentVersionRepository{pool: pool}
}

// CreateVersion stores a new document version
func (r *DocumentVersionRepository) CreateVersion(ctx context.Context, version *domain.DocumentVersion) error {
	query := `
		INSERT INTO document.file_versions (
			version_id, file_id, version_number, minio_key, size_bytes, created_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, query,
		version.ID().String(),
		version.FileID().String(),
		version.VersionNumber(),
		version.MinioKey(),
		version.SizeBytes(),
		version.CreatedBy(),
		version.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("create version: %w", err)
	}
	return nil
}

// GetVersionByID retrieves a specific version
func (r *DocumentVersionRepository) GetVersionByID(ctx context.Context, versionID domain.DocumentVersionID) (*domain.DocumentVersion, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetVersionsByDocumentID retrieves all versions of a document
func (r *DocumentVersionRepository) GetVersionsByDocumentID(ctx context.Context, fileID domain.DocumentID) ([]*domain.DocumentVersion, error) {
	query := `SELECT version_id, file_id, version_number, minio_key, size_bytes, created_by, created_at FROM document.file_versions WHERE file_id = $1 ORDER BY version_number DESC`

	rows, err := r.pool.Query(ctx, query, fileID.String())
	if err != nil {
		return nil, fmt.Errorf("query versions: %w", err)
	}
	defer rows.Close()

	var versions []*domain.DocumentVersion
	for rows.Next() {
		var (
			versionID, fileIDStr, createdBy string
			versionNumber                   int32
			minioKey                        string
			sizeBytes                       int64
			createdAt                       time.Time
		)

		if err = rows.Scan(&versionID, &fileIDStr, &versionNumber, &minioKey, &sizeBytes, &createdBy, &createdAt); err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}

		version, err := domain.RehydrateDocumentVersion(
			domain.MustNewDocumentVersionID(versionID),
			domain.MustNewDocumentID(fileIDStr),
			versionNumber,
			minioKey,
			sizeBytes,
			createdBy,
			createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("rehydrate version: %w", err)
		}

		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// GetVersionByNumber retrieves a specific version by version number
func (r *DocumentVersionRepository) GetVersionByNumber(ctx context.Context, fileID domain.DocumentID, versionNumber int32) (*domain.DocumentVersion, error) {
	return nil, fmt.Errorf("not implemented")
}
