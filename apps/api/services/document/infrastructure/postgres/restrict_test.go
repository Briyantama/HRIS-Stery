package postgres

import (
	"strings"
	"testing"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
)

func TestRestrictTo(t *testing.T) {
	const where = "tenant_id = $1"
	base := []any{"t"}

	if w, a := restrictTo(domain.Viewer{UserID: "u", HR: true}, where, base); w != where || len(a) != 1 {
		t.Fatalf("HR must stay unrestricted, got %q %v", w, a)
	}

	w, a := restrictTo(domain.Viewer{UserID: "AB-CD"}, where, base)
	if len(a) != 2 || a[1] != "ab-cd" {
		t.Fatalf("owner arg: %v", a)
	}
	for _, want := range []string{"classification <> 'SECRET'", "uploaded_by::text = $2", "entity_id::text = $2"} {
		if !strings.Contains(w, want) {
			t.Errorf("%q missing from %q", want, w)
		}
	}
}
