package docgen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kylix/pkg/docgen"
)

// Code example extraction tests (v2.5.0 task 2).

func TestDocGen_CodeExampleInFunction(t *testing.T) {
	src := "// Reverse returns the reversed string.\n" +
		"//\n" +
		"// ```pascal\n" +
		"// WriteLn(Reverse('abc'));  // cba\n" +
		"// ```\n" +
		"function Reverse(s: String): String;\n" +
		"begin result := ''; end;\n"
	f := writeDocFile(t, src)
	doc, err := docgen.GenerateFile(f)
	if err != nil {
		t.Fatal(err)
	}
	md := docgen.RenderMarkdown(doc)

	if !strings.Contains(md, "```pascal") {
		t.Errorf("expected code block in markdown")
	}
	if !strings.Contains(md, "Reverse('abc')") {
		t.Errorf("expected example code preserved")
	}
}

func TestDocGen_CodeExampleInUnit(t *testing.T) {
	src := "// Math utilities.\n" +
		"//\n" +
		"// Example:\n" +
		"// ```pascal\n" +
		"// WriteLn(Abs(-5));  // 5\n" +
		"// ```\n" +
		"unit mathutil;\n" +
		"function Abs(x: Integer): Integer;\n" +
		"begin result := x; end;\n"
	f := writeDocFile(t, src)
	doc, err := docgen.GenerateFile(f)
	if err != nil {
		t.Fatal(err)
	}
	md := docgen.RenderMarkdown(doc)

	if !strings.Contains(md, "Abs(-5)") {
		t.Errorf("expected unit-level example preserved")
	}
}

func TestDocGen_MultilineComment(t *testing.T) {
	src := `// This is a multi-line comment.
// It spans multiple lines.
// Each line is preserved.
function Foo(): Integer;
begin result := 0; end;
`
	f := writeDocFile(t, src)
	doc, err := docgen.GenerateFile(f)
	if err != nil {
		t.Fatal(err)
	}
	md := docgen.RenderMarkdown(doc)

	if !strings.Contains(md, "multi-line comment") {
		t.Errorf("expected first line preserved, got:\n%s", md)
	}
	if !strings.Contains(md, "Each line is preserved") {
		t.Errorf("expected last line preserved, got:\n%s", md)
	}
}

func TestDocGen_NoCodeBlock(t *testing.T) {
	src := `// Simple function.
function Foo(): Integer;
begin result := 0; end;
`
	f := writeDocFile(t, src)
	doc, err := docgen.GenerateFile(f)
	if err != nil {
		t.Fatal(err)
	}
	md := docgen.RenderMarkdown(doc)

	// Should have exactly 1 code block (the signature), not 2.
	count := strings.Count(md, "```pascal")
	if count != 1 {
		t.Errorf("expected 1 code block (signature only), got %d", count)
	}
}

func writeDocFile(t *testing.T, src string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "test.klx")
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}
