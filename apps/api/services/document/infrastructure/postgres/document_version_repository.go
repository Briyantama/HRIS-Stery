package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const versionColumns = `version_id, file_id, version_number, object_key, size_bytes, created_by, created_at`

// DocumentVersionRepository implements DocumentVersionRepository port. file_versions has no RLS;
// callers load the owning document through the tenant-scoped DocumentRepository first.
type DocumentVersionRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentVersionRepository(pool *pgxpool.Pool) *DocumentVersionRepository {
	return &DocumentVersionRepository{pool: pool}
}

// CreateVersion stores a new document version
func (r *DocumentVersionRepository) CreateVersion(ctx context.Context, v *domain.DocumentVersion) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO document.file_versions (`+versionColumns+`) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		v.ID().String(), v.FileID().String(), v.VersionNumber(), v.ObjectKey(), v.SizeBytes(), v.CreatedBy(), v.CreatedAt())
	if err != nil {
		return fmt.Errorf("create version: %w", err)
	}
	return nil
}

// GetVersionsByDocumentID retrieves all versions of a document, newest first
func (r *DocumentVersionRepository) GetVersionsByDocumentID(ctx context.Context, fileID domain.DocumentID) ([]*domain.DocumentVersion, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+versionColumns+` FROM document.file_versions WHERE file_id = $1 ORDER BY version_number DESC`, fileID.String())
	if err != nil {
		return nil, fmt.Errorf("query versions: %w", err)
	}
	defer rows.Close()

	var versions []*domain.DocumentVersion
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// GetVersionByNumber retrieves a specific version by version number
func (r *DocumentVersionRepository) GetVersionByNumber(ctx context.Context, fileID domain.DocumentID, versionNumber int32) (*domain.DocumentVersion, error) {
	v, err := scanVersion(r.pool.QueryRow(ctx, `SELECT `+versionColumns+` FROM document.file_versions WHERE file_id = $1 AND version_number = $2`, fileID.String(), versionNumber))
	if err != nil {
		if errors.Is(err, domain.ErrVersionNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get version by number: %w", err)
	}
	return v, nil
}

func scanVersion(row pgx.Row) (*domain.DocumentVersion, error) {
	var (
		versionID, fileID, objectKey, createdBy string
		versionNumber                           int32
		sizeBytes                               int64
		createdAt                               time.Time
	)
	if err := row.Scan(&versionID, &fileID, &versionNumber, &objectKey, &sizeBytes, &createdBy, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVersionNotFound
		}
		return nil, fmt.Errorf("scan version: %w", err)
	}
	vid, err := domain.NewDocumentVersionIDFromString(versionID)
	if err != nil {
		return nil, err
	}
	fid, err := domain.NewDocumentIDFromString(fileID)
	if err != nil {
		return nil, err
	}
	return domain.RehydrateDocumentVersion(vid, fid, versionNumber, objectKey, sizeBytes, createdBy, createdAt)
}
