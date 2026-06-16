package compiler_test

import (
	"os"
	"testing"

	"kylix/pkg/compiler"
)

func TestGenericConstraint_BuiltinViolation(t *testing.T) {
	src := `program Test;
type
  IComparable = interface
    function CompareTo(): Integer;
  end;
  TBox<T: IComparable> = class
  end;
var b: TBox<Integer>;
begin end.`
	f := genTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !genHasDiagCode(r.Diagnostics, compiler.ErrGenericConstraint) {
		t.Errorf("expected KLX104 for Integer not satisfying IComparable, got: %v", genFmtDiags(r.Diagnostics))
	}
}

func TestGenericConstraint_CustomType(t *testing.T) {
	src := `program Test;
type
  IComparable = interface
    function CompareTo(): Integer;
  end;
  TBox<T: IComparable> = class
  end;
  TMyType = class implements IComparable
    function CompareTo(): Integer; begin result := 0; end;
  end;
var b: TBox<TMyType>;
begin end.`
	f := genTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range r.Diagnostics {
		if d.Code == compiler.ErrGenericConstraint {
			t.Errorf("unexpected KLX104 for custom type implementing interface: %s", d.Message)
		}
	}
}

func TestGenericConstraint_NoConstraint(t *testing.T) {
	src := `program Test;
type
  TBox<T> = class
  end;
var b: TBox<Integer>;
begin end.`
	f := genTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range r.Diagnostics {
		if d.Code == compiler.ErrGenericConstraint {
			t.Errorf("unexpected KLX104 when no constraint declared: %s", d.Message)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func genTmpKlx(t *testing.T, src string) string {
	t.Helper()
	f := t.TempDir() + "/test.klx"
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}

func genHasDiagCode(diags []compiler.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func genFmtDiags(diags []compiler.Diagnostic) []string {
	out := make([]string, len(diags))
	for i, d := range diags {
		out[i] = d.Code + ": " + d.Message
	}
	return out
}
