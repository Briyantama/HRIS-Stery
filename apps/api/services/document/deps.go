package document

import (
	"errors"

	"github.com/Briyantama/HRIS-Stery/internal/config"
	"github.com/Briyantama/HRIS-Stery/services/document/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/document/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/document/infrastructure/postgres"
	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/objstore"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps is everything the document service needs from the process.
type Deps struct {
	Config   config.Config
	Pool     *pgxpool.Pool
	Verifier authn.Verifier
	Store    objstore.Store
}

// Build wires repositories, command and query handlers and the HTTP handler.
func Build(d Deps) (*HTTPHandler, error) {
	if d.Pool == nil || d.Verifier == nil || d.Store == nil {
		return nil, errors.New("document: Pool, Verifier and Store are required")
	}
	docs := postgres.NewDocumentRepository(d.Pool)
	versions := postgres.NewDocumentVersionRepository(d.Pool)
	create := commands.NewCreateDocumentVersionHandler(docs, versions)

	return NewHTTPHandler(
		commands.NewUploadDocumentHandler(docs, d.Store, d.Config.S3.Bucket),
		queries.NewListDocumentsHandler(docs),
		queries.NewGetDocumentMetadataHandler(docs, versions),
		queries.NewGetDownloadURLHandler(docs, versions, d.Store),
		commands.NewDeleteDocumentHandler(docs),
		commands.NewRollbackDocumentVersionHandler(docs, versions, create),
		d.Verifier,
	), nil
}
