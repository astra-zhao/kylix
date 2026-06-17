package compiler_test

import (
	"os"
	"testing"

	"kylix/pkg/compiler"
)

// Multi-parameter generic constraint tests (M2.1.1)

func TestGenericMulti_TwoConstraints(t *testing.T) {
	src := `program Test;
type
  IComparable = interface
    function CompareTo(): Integer;
  end;
  IHashable = interface
    function HashCode(): Integer;
  end;
  TMap<K: IComparable, V: IHashable> = class
  end;
var m: TMap<Integer, String>;
begin end.`
	f := genMultiTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Both Integer and String should fail (built-ins don't implement user interfaces)
	codes := genMultiCollectCodes(r.Diagnostics)
	klx104 := 0
	for _, c := range codes {
		if c == compiler.ErrGenericConstraint {
			klx104++
		}
	}
	if klx104 < 2 {
		t.Errorf("expected ≥2 KLX104 errors for both K and V, got %d: %v", klx104, codes)
	}
}

func TestGenericMulti_OnlyFirstConstraintViolated(t *testing.T) {
	src := `program Test;
type
  IComparable = interface
    function CompareTo(): Integer;
  end;
  TMap<K: IComparable, V> = class
  end;
var m: TMap<Integer, String>;
begin end.`
	f := genMultiTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Only K (Integer) should fail; V (String) is unconstrained
	codes := genMultiCollectCodes(r.Diagnostics)
	klx104 := 0
	for _, c := range codes {
		if c == compiler.ErrGenericConstraint {
			klx104++
		}
	}
	if klx104 != 1 {
		t.Errorf("expected exactly 1 KLX104 (only K), got %d: %v", klx104, codes)
	}
	// Verify error message mentions parameter K
	for _, d := range r.Diagnostics {
		if d.Code == compiler.ErrGenericConstraint {
			if !genMultiContains(d.Message, "K") {
				t.Errorf("expected message to mention parameter K, got: %s", d.Message)
			}
		}
	}
}

func TestGenericMulti_BothSatisfied(t *testing.T) {
	src := `program Test;
type
  IFoo = interface
    procedure Foo();
  end;
  TPair<A, B> = class
  end;
var p: TPair<Integer, String>;
begin end.`
	f := genMultiTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Unconstrained generics: no KLX104 errors
	for _, d := range r.Diagnostics {
		if d.Code == compiler.ErrGenericConstraint {
			t.Errorf("unexpected KLX104 for unconstrained TPair: %s", d.Message)
		}
	}
}

func TestGenericMulti_ParameterOrderPreserved(t *testing.T) {
	// Verify parameter order is correctly used for matching
	src := `program Test;
type
  IFirst = interface
    procedure First();
  end;
  ISecond = interface
    procedure Second();
  end;
  TThree<A: IFirst, B, C: ISecond> = class
  end;
var t: TThree<Integer, String, Boolean>;
begin end.`
	f := genMultiTmpKlx(t, src)
	r, err := compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	// A=Integer should fail (IFirst), B=String unconstrained, C=Boolean should fail (ISecond)
	codes := genMultiCollectCodes(r.Diagnostics)
	klx104 := 0
	hasA, hasC := false, false
	for _, d := range r.Diagnostics {
		if d.Code == compiler.ErrGenericConstraint {
			klx104++
			if genMultiContains(d.Message, "A") && genMultiContains(d.Message, "IFirst") {
				hasA = true
			}
			if genMultiContains(d.Message, "C") && genMultiContains(d.Message, "ISecond") {
				hasC = true
			}
		}
	}
	if klx104 != 2 {
		t.Errorf("expected exactly 2 KLX104 (A and C), got %d: %v", klx104, codes)
	}
	if !hasA {
		t.Error("expected error about parameter A failing IFirst")
	}
	if !hasC {
		t.Error("expected error about parameter C failing ISecond")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func genMultiTmpKlx(t *testing.T, src string) string {
	t.Helper()
	f := t.TempDir() + "/test.klx"
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}

func genMultiCollectCodes(diags []compiler.Diagnostic) []string {
	codes := make([]string, len(diags))
	for i, d := range diags {
		codes[i] = d.Code
	}
	return codes
}

func genMultiContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
