package commands_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Briyantama/HRIS-Stery/services/document/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/document/domain"
	"github.com/Briyantama/stery-kit/go/objstore"
	"uuid"
)

// repo records the created document; the other repository methods are not used by upload.
type repo struct {
	domain.DocumentRepository
	created *domain.DocumentFile
}

func (r *repo) CreateDocument(_ context.Context, d *domain.DocumentFile) error {
	r.created = d
	return nil
}

func TestUploadSignsTenantPrefixedKey(t *testing.T) {
	r := &repo{}
	store := objstore.NewMemory()
	h := commands.NewUploadDocumentHandler(r, store, "hris-documents")
	tenant, entity := uuid.NewV7().String(), uuid.NewV7().String()
	cmd := commands.UploadDocumentCommand{
		TenantID: tenant, UploadedBy: uuid.NewV7().String(), EntityType: "EMPLOYEE", EntityID: entity,
		FileName: "cv.pdf", MimeType: "application/pdf", SizeBytes: 10, Classification: "PUBLIC",
	}

	res, err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	wantKey := tenant + "/EMPLOYEE/" + entity + "/1/" + res.DocumentID
	if r.created.ObjectKey() != wantKey || r.created.Bucket() != "hris-documents" {
		t.Fatalf("stored %q in %q, want %q", r.created.ObjectKey(), r.created.Bucket(), wantKey)
	}
	if err := store.Upload(objstore.PresignedRequest{URL: res.PresignedURL, Headers: res.Headers}, "application/pdf", make([]byte, 10)); err != nil {
		t.Fatalf("presigned upload rejected: %v", err)
	}
	if got := store.Keys(); len(got) != 1 || got[0] != wantKey {
		t.Fatalf("object keys %v", got)
	}
}

func TestUploadValidation(t *testing.T) {
	h := commands.NewUploadDocumentHandler(&repo{}, objstore.NewMemory(), "b")
	base := commands.UploadDocumentCommand{
		TenantID: uuid.NewV7().String(), UploadedBy: uuid.NewV7().String(), EntityType: "EMPLOYEE", EntityID: uuid.NewV7().String(),
		FileName: "a", MimeType: "text/plain", SizeBytes: 1, Classification: "PUBLIC",
	}
	for name, mutate := range map[string]func(*commands.UploadDocumentCommand){
		"too large":       func(c *commands.UploadDocumentCommand) { c.SizeBytes = domain.MaxFileBytes + 1 },
		"zero size":       func(c *commands.UploadDocumentCommand) { c.SizeBytes = 0 },
		"entity not uuid": func(c *commands.UploadDocumentCommand) { c.EntityID = "x" },
		"bad type":        func(c *commands.UploadDocumentCommand) { c.EntityType = "CAR" },
		"long name":       func(c *commands.UploadDocumentCommand) { c.FileName = strings.Repeat("a", 256) },
	} {
		cmd := base
		mutate(&cmd)
		if _, err := h.Handle(context.Background(), cmd); !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("%s: got %v, want ErrInvalidInput", name, err)
		}
	}
}
