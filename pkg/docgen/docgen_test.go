package docgen_test

import (
	"kylix/pkg/docgen"
	"os"
	"strings"
	"testing"
)

func write(t *testing.T, src string) string {
	t.Helper()
	f := t.TempDir() + "/test.klx"
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestGenerateFile_BasicUnit(t *testing.T) {
	f := write(t, `// Math utilities for the application.
unit math;

// Add returns the sum of a and b.
function Add(a: Integer; b: Integer): Integer;
begin result := a + b; end;
`)
	doc, err := docgen.GenerateFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Name != "math" {
		t.Errorf("expected unit name 'math', got %q", doc.Name)
	}
	if !strings.Contains(doc.Comment, "Math utilities") {
		t.Errorf("expected unit comment, got %q", doc.Comment)
	}
	if len(doc.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(doc.Entries))
	}
	e := doc.Entries[0]
	if e.Kind != "function" {
		t.Errorf("expected kind=function, got %q", e.Kind)
	}
	if e.Name != "Add" {
		t.Errorf("expected name=Add, got %q", e.Name)
	}
	if !strings.Contains(e.Comment, "sum of a and b") {
		t.Errorf("expected comment, got %q", e.Comment)
	}
}

func TestRenderMarkdown_HasSections(t *testing.T) {
	f := write(t, `unit demo;
// MyConst is a constant.
const MyConst = 42;
// MyFunc does something.
function MyFunc(): Integer; begin result := 0; end;
// TMyClass is a class.
type TMyClass = class end;
`)
	doc, err := docgen.GenerateFile(f)
	if err != nil {
		t.Fatal(err)
	}
	md := docgen.RenderMarkdown(doc)

	if !strings.Contains(md, "# demo") {
		t.Error("expected # demo header")
	}
	if !strings.Contains(md, "## Constants") {
		t.Error("expected ## Constants section")
	}
	if !strings.Contains(md, "## Functions") {
		t.Error("expected ## Functions section")
	}
	if !strings.Contains(md, "## Classes") {
		t.Error("expected ## Classes section")
	}
	if !strings.Contains(md, "### MyFunc") {
		t.Error("expected ### MyFunc entry")
	}
}

func TestRenderMarkdown_FunctionSignature(t *testing.T) {
	f := write(t, `unit demo;
function Add(a: Integer; b: Integer): Integer;
begin result := a + b; end;
`)
	doc, err := docgen.GenerateFile(f)
	if err != nil {
		t.Fatal(err)
	}
	md := docgen.RenderMarkdown(doc)

	if !strings.Contains(md, "function Add(a: Integer; b: Integer): Integer") {
		t.Errorf("expected full function signature in markdown, got:\n%s", md)
	}
}

func TestGenerateFile_NoComments(t *testing.T) {
	f := write(t, `unit silent;
function Foo(): Integer;
begin result := 0; end;
`)
	doc, err := docgen.GenerateFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Comment != "" {
		t.Errorf("expected empty unit comment, got %q", doc.Comment)
	}
	if len(doc.Entries) != 1 || doc.Entries[0].Comment != "" {
		t.Error("expected no doc comment on function")
	}
}

func TestGenerateFile_TypeAlias(t *testing.T) {
	f := write(t, `unit types;
// UserId is the primary key type for users.
type UserId = Integer;
`)
	doc, err := docgen.GenerateFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(doc.Entries))
	}
	e := doc.Entries[0]
	if e.Kind != "type" {
		t.Errorf("expected kind=type, got %q", e.Kind)
	}
	if !strings.Contains(e.Comment, "primary key") {
		t.Errorf("expected comment, got %q", e.Comment)
	}
}
