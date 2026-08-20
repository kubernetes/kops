package template

import (
	"io"
	"strings"
	"testing"
)

type pruningTestData struct {
	Field string
}

func (pruningTestData) Method() string { return "from-method" }

// TestFieldsAndMapKeysRender pins the supported surface of the fork: struct fields, map keys and
// template functions.
func TestFieldsAndMapKeysRender(t *testing.T) {
	var sb strings.Builder
	tmpl := Must(New("t").Funcs(FuncMap{"upper": strings.ToUpper}).Parse(`{{upper .Field}}`))
	if err := tmpl.Execute(&sb, pruningTestData{Field: "f"}); err != nil {
		t.Fatalf("struct field rendering failed: %v", err)
	}
	if got := sb.String(); got != "F" {
		t.Fatalf("unexpected output %q", got)
	}

	sb.Reset()
	tmpl = Must(New("t").Parse(`{{.A}}/{{.B}}`))
	if err := tmpl.Execute(&sb, map[string]string{"A": "x", "B": "y"}); err != nil {
		t.Fatalf("map key rendering failed: %v", err)
	}
	if got := sb.String(); got != "x/y" {
		t.Fatalf("unexpected output %q", got)
	}
}

// TestMethodsAreNotResolved pins the fork's deliberate difference from the stdlib: field names
// never resolve to methods on the data, so the linker can keep pruning unused methods.
func TestMethodsAreNotResolved(t *testing.T) {
	tmpl := Must(New("t").Parse(`{{.Method}}`))
	err := tmpl.Execute(io.Discard, pruningTestData{})
	if err == nil {
		t.Fatal("expected an error: method resolution should be disabled in this fork")
	}
	if !strings.Contains(err.Error(), "can't evaluate field Method") {
		t.Fatalf("unexpected error: %v", err)
	}
}
