package compiler_test

import (
	"os"
	"path/filepath"
	"testing"

	"kylix/pkg/compiler"
)

// Project-level checking tests (Task 3 — CheckProject).

func TestCheckProject_HappyPath(t *testing.T) {
	dir := t.TempDir()
	mathFile := filepath.Join(dir, "math.klx")
	mainFile := filepath.Join(dir, "main.klx")

	mathSrc := `unit math;
function Square(n: Integer): Integer;
begin
  result := n * n;
end;
`
	mainSrc := `program Main;
uses math;
begin
  WriteLn(Square(5));
end.
`
	os.WriteFile(mathFile, []byte(mathSrc), 0644)
	os.WriteFile(mainFile, []byte(mainSrc), 0644)

	r, err := compiler.CheckProject([]string{mathFile, mainFile})
	if err != nil {
		t.Fatal(err)
	}
	if !r.Success {
		t.Errorf("expected success, got diagnostics: %v", projDiagMessages(r.Diagnostics))
	}
}

func TestCheckProject_UndeclaredFunctionCall(t *testing.T) {
	dir := t.TempDir()
	mainFile := filepath.Join(dir, "main.klx")

	src := `program Main;
begin
  WriteLn(Cube(5));
end.
`
	os.WriteFile(mainFile, []byte(src), 0644)

	r, err := compiler.CheckProject([]string{mainFile})
	if err != nil {
		t.Fatal(err)
	}
	if r.Success {
		t.Fatal("expected failure for undeclared Cube")
	}
	if !projHasCode(r.Diagnostics, compiler.ErrUndeclared) {
		t.Errorf("expected KLX201 for undeclared Cube, got: %v", projDiagMessages(r.Diagnostics))
	}
}

func TestCheckProject_CrossFileSymbol(t *testing.T) {
	// Cross-file: Square defined in math.klx, called in main.klx → should NOT report undeclared.
	dir := t.TempDir()
	files := writeProjFiles(t, dir, map[string]string{
		"math.klx": `unit math;
function Square(n: Integer): Integer;
begin result := n * n; end;
`,
		"main.klx": `program Main;
uses math;
begin
  WriteLn(Square(5));
end.
`,
	})

	r, err := compiler.CheckProject(files)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Success {
		t.Errorf("cross-file Square call should not be reported as undeclared, got: %v",
			projDiagMessages(r.Diagnostics))
	}
}

func TestCheckProject_UnknownUnit(t *testing.T) {
	dir := t.TempDir()
	mainFile := filepath.Join(dir, "main.klx")

	src := `program Main;
uses no_such_unit;
begin end.
`
	os.WriteFile(mainFile, []byte(src), 0644)

	r, err := compiler.CheckProject([]string{mainFile})
	if err != nil {
		t.Fatal(err)
	}
	if r.Success {
		t.Fatal("expected failure for unknown unit")
	}
	found := false
	for _, d := range r.Diagnostics {
		if projContains(d.Message, "no_such_unit") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected diagnostic mentioning 'no_such_unit', got: %v", projDiagMessages(r.Diagnostics))
	}
}

func TestCheckProject_StdlibUnitOK(t *testing.T) {
	// 'sysutil' is a known stdlib unit — should not report unknown.
	dir := t.TempDir()
	mainFile := filepath.Join(dir, "main.klx")

	src := `program Main;
uses sysutil;
begin end.
`
	os.WriteFile(mainFile, []byte(src), 0644)

	r, err := compiler.CheckProject([]string{mainFile})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range r.Diagnostics {
		if projContains(d.Message, "unknown unit") {
			t.Errorf("sysutil is stdlib, should not be flagged as unknown: %s", d.Message)
		}
	}
}

func TestCheckProject_TypeMismatchAcrossFiles(t *testing.T) {
	// math.klx defines Square; main.klx assigns its return to wrong-type variable.
	dir := t.TempDir()
	files := writeProjFiles(t, dir, map[string]string{
		"math.klx": `unit math;
function Square(n: Integer): Integer;
begin result := n * n; end;
`,
		"main.klx": `program Main;
uses math;
var s: String;
begin
  s := 42;
end.
`,
	})

	r, err := compiler.CheckProject(files)
	if err != nil {
		t.Fatal(err)
	}
	if !projHasCode(r.Diagnostics, compiler.ErrTypeMismatch) {
		t.Errorf("expected KLX101 for Integer→String, got: %v", projDiagMessages(r.Diagnostics))
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeProjFiles(t *testing.T, dir string, files map[string]string) []string {
	t.Helper()
	out := make([]string, 0, len(files))
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	return out
}

func projHasCode(diags []compiler.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func projDiagMessages(diags []compiler.Diagnostic) []string {
	out := make([]string, len(diags))
	for i, d := range diags {
		out[i] = d.Code + ": " + d.Message
	}
	return out
}

func projContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
