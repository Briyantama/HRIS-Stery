package queries_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Briyantama/HRIS-Stery/services/document/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/document/domain"
	"github.com/Briyantama/stery-kit/go/objstore"
	"github.com/google/uuid"
)

type docRepo struct {
	domain.DocumentRepository
	doc       *domain.DocumentFile
	gotViewer domain.Viewer
}

func (r *docRepo) GetDocumentByID(context.Context, domain.TenantID, domain.DocumentID) (*domain.DocumentFile, error) {
	return r.doc, nil
}

func (r *docRepo) SearchDocuments(_ context.Context, _ domain.TenantID, v domain.Viewer, _ string, _, _ int) ([]*domain.DocumentFile, int, error) {
	r.gotViewer = v
	return nil, 0, nil
}

type versionRepo struct {
	domain.DocumentVersionRepository
}

func (versionRepo) GetVersionByNumber(_ context.Context, f domain.DocumentID, n int32) (*domain.DocumentVersion, error) {
	return domain.NewDocumentVersion(f, n, "k", 1, uuid.NewString())
}
func (versionRepo) GetVersionsByDocumentID(context.Context, domain.DocumentID) ([]*domain.DocumentVersion, error) {
	return nil, nil
}

var (
	tenant = uuid.NewString()
	owner  = uuid.NewString()
	other  = uuid.NewString()
)

func newDoc(t *testing.T, entityType domain.EntityType, entityID string, class domain.FileClassification) *domain.DocumentFile {
	t.Helper()
	id := domain.NewDocumentID()
	doc, err := domain.NewDocumentFile(id, domain.MustNewTenantID(tenant), uuid.NewString(), entityType, entityID,
		"contract.pdf", "application/pdf", 1, "b", "k", class, 0)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestDocumentAccess(t *testing.T) {
	hr := domain.Viewer{UserID: uuid.NewString(), HR: true}
	ownerV := domain.Viewer{UserID: owner}
	otherV := domain.Viewer{UserID: other}

	cases := []struct {
		name   string
		doc    *domain.DocumentFile
		viewer domain.Viewer
		want   bool
	}{
		{"owner of employee record", newDoc(t, domain.EntityTypeEmployee, owner, domain.ClassificationConfidential), ownerV, true},
		{"other employee", newDoc(t, domain.EntityTypeEmployee, owner, domain.ClassificationPublic), otherV, false},
		{"hr sees all", newDoc(t, domain.EntityTypeEmployee, owner, domain.ClassificationSecret), hr, true},
		{"owner blocked from secret", newDoc(t, domain.EntityTypeEmployee, owner, domain.ClassificationSecret), ownerV, false},
		{"general doc of someone else", newDoc(t, domain.EntityTypeGeneral, uuid.NewString(), domain.ClassificationPublic), otherV, false},
		{"empty user id", newDoc(t, domain.EntityTypeEmployee, owner, domain.ClassificationPublic), domain.Viewer{}, false},
	}
	for _, c := range cases {
		docs := &docRepo{doc: c.doc}
		id := c.doc.ID().String()
		meta := queries.NewGetDocumentMetadataHandler(docs, versionRepo{})
		_, err := meta.Handle(context.Background(), queries.GetDocumentMetadataQuery{TenantID: tenant, Viewer: c.viewer, DocumentID: id})
		checkAccess(t, c.name+" metadata", err, c.want)

		dl := queries.NewGetDownloadURLHandler(docs, versionRepo{}, objstore.NewMemory())
		_, err = dl.Handle(context.Background(), queries.GetDownloadURLQuery{TenantID: tenant, Viewer: c.viewer, DocumentID: id})
		checkAccess(t, c.name+" download", err, c.want)
	}
}

func checkAccess(t *testing.T, name string, err error, allowed bool) {
	t.Helper()
	switch {
	case allowed && err != nil:
		t.Errorf("%s: unexpected error %v", name, err)
	case !allowed && !errors.Is(err, domain.ErrDocumentNotFound):
		t.Errorf("%s: got %v, want ErrDocumentNotFound", name, err)
	}
}

func TestListPassesViewerToRepository(t *testing.T) {
	repo := &docRepo{}
	v := domain.Viewer{UserID: owner}
	_, err := queries.NewListDocumentsHandler(repo).Handle(context.Background(),
		queries.ListDocumentsQuery{TenantID: tenant, Viewer: v, Search: "x", PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if repo.gotViewer != v {
		t.Fatalf("repository saw viewer %+v, want %+v", repo.gotViewer, v)
	}
}
