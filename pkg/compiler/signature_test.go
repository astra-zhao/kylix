package compiler_test

import (
	"os"
	"testing"

	"kylix/pkg/compiler"
)

// Method signature verification for class→interface implementation.
// This is a strict-mode follow-up to v2.1.2's name-only check.

func TestSignature_MatchingSignature(t *testing.T) {
	src := `program Test;
type
  IFoo = interface
    function Bar(x: Integer): String;
  end;
  TBox<T: IFoo> = class
  end;
  TGood = class implements IFoo
    function Bar(x: Integer): String; begin result := 'ok'; end;
  end;
var b: TBox<TGood>;
begin end.`
	r, err := sigCompile(t, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range r.Diagnostics {
		if d.Code == compiler.ErrGenericConstraint {
			t.Errorf("unexpected KLX104: %s", d.Message)
		}
	}
}

func TestSignature_WrongReturnType(t *testing.T) {
	src := `program Test;
type
  IFoo = interface
    function Bar(): String;
  end;
  TBox<T: IFoo> = class
  end;
  TBad = class implements IFoo
    function Bar(): Integer; begin result := 0; end;
  end;
var b: TBox<TBad>;
begin end.`
	r, err := sigCompile(t, src)
	if err != nil {
		t.Fatal(err)
	}
	if !sigHasCode(r.Diagnostics, compiler.ErrGenericConstraint) {
		t.Errorf("expected KLX104 — Bar return type mismatch (Integer vs String), got: %v",
			sigCodes(r.Diagnostics))
	}
}

func TestSignature_WrongParamCount(t *testing.T) {
	src := `program Test;
type
  IFoo = interface
    function Bar(a: Integer; b: Integer): Integer;
  end;
  TBox<T: IFoo> = class
  end;
  TBad = class implements IFoo
    function Bar(a: Integer): Integer; begin result := a; end;
  end;
var b: TBox<TBad>;
begin end.`
	r, err := sigCompile(t, src)
	if err != nil {
		t.Fatal(err)
	}
	if !sigHasCode(r.Diagnostics, compiler.ErrGenericConstraint) {
		t.Errorf("expected KLX104 — Bar parameter count mismatch (1 vs 2), got: %v",
			sigCodes(r.Diagnostics))
	}
}

func TestSignature_WrongParamType(t *testing.T) {
	src := `program Test;
type
  IFoo = interface
    function Bar(x: Integer): Integer;
  end;
  TBox<T: IFoo> = class
  end;
  TBad = class implements IFoo
    function Bar(x: String): Integer; begin result := 0; end;
  end;
var b: TBox<TBad>;
begin end.`
	r, err := sigCompile(t, src)
	if err != nil {
		t.Fatal(err)
	}
	if !sigHasCode(r.Diagnostics, compiler.ErrGenericConstraint) {
		t.Errorf("expected KLX104 — Bar param type mismatch (String vs Integer), got: %v",
			sigCodes(r.Diagnostics))
	}
}

func TestSignature_AliasedTypesMatch(t *testing.T) {
	// Type alias should be transparent for signature comparison.
	src := `program Test;
type
  UserId = Integer;
  IFoo = interface
    function Get(id: UserId): String;
  end;
  TBox<T: IFoo> = class
  end;
  TGood = class implements IFoo
    function Get(id: Integer): String; begin result := ''; end;
  end;
var b: TBox<TGood>;
begin end.`
	r, err := sigCompile(t, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range r.Diagnostics {
		if d.Code == compiler.ErrGenericConstraint {
			t.Errorf("unexpected KLX104 — UserId is alias for Integer, signatures should match: %s",
				d.Message)
		}
	}
}

func TestSignature_ProcedureWithMatchingParams(t *testing.T) {
	src := `program Test;
type
  IAction = interface
    procedure Do(s: String);
  end;
  TBox<T: IAction> = class
  end;
  TWorker = class implements IAction
    procedure Do(s: String); begin end;
  end;
var b: TBox<TWorker>;
begin end.`
	r, err := sigCompile(t, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range r.Diagnostics {
		if d.Code == compiler.ErrGenericConstraint {
			t.Errorf("unexpected KLX104 for procedure: %s", d.Message)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func sigCompile(t *testing.T, src string) (*compiler.Result, error) {
	t.Helper()
	f := t.TempDir() + "/test.klx"
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return compiler.CompileFile(f, compiler.Options{
		OutputFile: t.TempDir() + "/out.go",
	})
}

func sigHasCode(diags []compiler.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func sigCodes(diags []compiler.Diagnostic) []string {
	out := make([]string, len(diags))
	for i, d := range diags {
		out[i] = d.Code + ": " + d.Message
	}
	return out
}
