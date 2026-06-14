package compiler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kylix/pkg/compiler"
)

func writeTC(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.klx")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func compileTC(t *testing.T, src string) *compiler.Result {
	t.Helper()
	f := writeTC(t, src)
	result, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: filepath.Join(t.TempDir(), "out.go"),
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func hasDiag(result *compiler.Result, substr string) bool {
	for _, d := range result.Diagnostics {
		if strings.Contains(d.Message, substr) {
			return true
		}
	}
	return false
}

// ── undeclared variable ───────────────────────────────────────────────────────

func TestTypeCheck_UndeclaredVar(t *testing.T) {
	result := compileTC(t, `program Test;
begin
  x := 42;
end.`)
	if result.Success {
		t.Fatal("expected failure for undeclared variable")
	}
	if !hasDiag(result, "undeclared") {
		t.Errorf("expected 'undeclared' diagnostic, got: %v", result.Diagnostics)
	}
}

func TestTypeCheck_DeclaredVarOK(t *testing.T) {
	result := compileTC(t, `program Test;
var x: Integer;
begin
  x := 42;
  WriteLn(x);
end.`)
	for _, d := range result.Diagnostics {
		if strings.Contains(d.Message, "undeclared") {
			t.Errorf("false positive undeclared: %s", d.Message)
		}
	}
}

// ── function arity ────────────────────────────────────────────────────────────

func TestTypeCheck_WrongArity(t *testing.T) {
	result := compileTC(t, `program Test;
function Add(a: Integer; b: Integer): Integer;
begin result := a + b; end;

begin
  WriteLn(Add(1, 2, 3));
end.`)
	if result.Success {
		t.Fatal("expected failure for wrong arity")
	}
	if !hasDiag(result, "wrong number of arguments") {
		t.Errorf("expected arity diagnostic, got: %v", result.Diagnostics)
	}
}

func TestTypeCheck_CorrectArity(t *testing.T) {
	result := compileTC(t, `program Test;
function Add(a: Integer; b: Integer): Integer;
begin result := a + b; end;

begin
  WriteLn(Add(1, 2));
end.`)
	if hasDiag(result, "wrong number") {
		t.Error("false positive arity check")
	}
}

// ── type assignment compatibility ─────────────────────────────────────────────

func TestTypeCheck_StringToInteger(t *testing.T) {
	result := compileTC(t, `program Test;
var x: Integer;
begin
  x := 'hello';
end.`)
	if result.Success {
		t.Fatal("expected failure for string assigned to Integer")
	}
	if !hasDiag(result, "cannot assign String") {
		t.Errorf("expected type mismatch diagnostic, got: %v", result.Diagnostics)
	}
}

func TestTypeCheck_IntegerToString(t *testing.T) {
	result := compileTC(t, `program Test;
var s: String;
begin
  s := 42;
end.`)
	if result.Success {
		t.Fatal("expected failure for integer assigned to String")
	}
	if !hasDiag(result, "cannot assign Integer") {
		t.Errorf("expected type mismatch diagnostic, got: %v", result.Diagnostics)
	}
}

func TestTypeCheck_CorrectTypes(t *testing.T) {
	result := compileTC(t, `program Test;
var x: Integer;
var s: String;
begin
  x := 42;
  s := 'hello';
  WriteLn(s);
end.`)
	for _, d := range result.Diagnostics {
		if d.Level == "error" {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}
