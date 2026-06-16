package compiler_test

import (
	"os"
	"testing"

	"kylix/pkg/compiler"
)

func TestTypeInfer_IntegerLiteral(t *testing.T) {
	src := `program Test;
begin
  var x := 42;
  x := 'hello';
end.`
	f := inferTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !inferHasDiagCode(r.Diagnostics, compiler.ErrTypeMismatch) {
		t.Errorf("expected KLX101 for string→integer, got: %v", inferFmtDiags(r.Diagnostics))
	}
}

func TestTypeInfer_StringLiteral(t *testing.T) {
	src := `program Test;
begin
  var name := 'Alice';
  name := 99;
end.`
	f := inferTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !inferHasDiagCode(r.Diagnostics, compiler.ErrTypeMismatch) {
		t.Errorf("expected KLX101 for integer→string, got: %v", inferFmtDiags(r.Diagnostics))
	}
}

func TestTypeInfer_FunctionReturn(t *testing.T) {
	src := `program Test;
function GetAge(): Integer;
begin result := 30; end;
begin
  var age := GetAge();
  age := 'thirty';
end.`
	f := inferTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !inferHasDiagCode(r.Diagnostics, compiler.ErrTypeMismatch) {
		t.Errorf("expected KLX101 for string→inferred Integer, got: %v", inferFmtDiags(r.Diagnostics))
	}
}

func TestTypeInfer_NoFalsePositive(t *testing.T) {
	src := `program Test;
begin
  var x := 42;
  var s := 'hello';
  x := 99;
  s := 'world';
end.`
	f := inferTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Diagnostics) > 0 {
		t.Errorf("expected no errors for compatible assignments, got: %v", inferFmtDiags(r.Diagnostics))
	}
}

func TestTypeInfer_BoolLiteral(t *testing.T) {
	src := `program Test;
begin
  var flag := true;
  flag := 42;
end.`
	f := inferTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !inferHasDiagCode(r.Diagnostics, compiler.ErrTypeMismatch) {
		t.Errorf("expected KLX101 for integer→inferred Boolean, got: %v", inferFmtDiags(r.Diagnostics))
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func inferTmpKlx(t *testing.T, src string) string {
	t.Helper()
	f := t.TempDir() + "/test.klx"
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}

func inferHasDiagCode(diags []compiler.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func inferFmtDiags(diags []compiler.Diagnostic) []string {
	out := make([]string, len(diags))
	for i, d := range diags {
		out[i] = d.Code + ": " + d.Message
	}
	return out
}
