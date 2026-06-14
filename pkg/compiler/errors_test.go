package compiler_test

import (
	"strings"
	"testing"

	"kylix/pkg/compiler"
)

func TestErrorCodes_TypeMismatch(t *testing.T) {
	result := compileTC(t, `program Test;
var x: Integer;
begin
  x := 'hello';
end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	d := result.Diagnostics[0]
	if d.Code != compiler.ErrTypeMismatch {
		t.Errorf("expected code %s, got %q", compiler.ErrTypeMismatch, d.Code)
	}
}

func TestErrorCodes_Undeclared(t *testing.T) {
	result := compileTC(t, `program Test;
begin
  x := 42;
end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	if !hasDiagCode(result, compiler.ErrUndeclared) {
		t.Errorf("expected code %s, got %v", compiler.ErrUndeclared, diagCodes(result))
	}
}

func TestErrorCodes_WrongArity(t *testing.T) {
	result := compileTC(t, `program Test;
function Add(a: Integer; b: Integer): Integer;
begin result := a + b; end;
begin
  WriteLn(Add(1, 2, 3));
end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	if !hasDiagCode(result, compiler.ErrWrongArity) {
		t.Errorf("expected code %s, got %v", compiler.ErrWrongArity, diagCodes(result))
	}
}

func TestErrorCodes_MissingMethod(t *testing.T) {
	result := compileTC(t, `program Test;
type
  IFoo = interface
    procedure Bar();
  end;
  TFoo = class implements IFoo
  end;
begin end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	if !hasDiagCode(result, compiler.ErrMissingMethod) {
		t.Errorf("expected code %s, got %v", compiler.ErrMissingMethod, diagCodes(result))
	}
	// Should have a hint
	for _, d := range result.Diagnostics {
		if d.Code == compiler.ErrMissingMethod && d.Hint == "" {
			t.Error("expected hint for missing method error")
		}
	}
}

func TestErrorCodes_MultipleErrors(t *testing.T) {
	result := compileTC(t, `program Test;
var x: Integer;
var s: String;
begin
  s := 42;
  x := 'hello';
end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	if len(result.Diagnostics) < 2 {
		t.Fatalf("expected >= 2 diagnostics, got %d", len(result.Diagnostics))
	}
	for _, d := range result.Diagnostics {
		if d.Code == "" {
			t.Errorf("diagnostic has no code: %s", d.Message)
		}
	}
}

func TestErrorCodes_Format(t *testing.T) {
	d := compiler.Diagnostic{
		File:    "main.klx",
		Line:    10,
		Column:  5,
		Level:   "error",
		Code:    compiler.ErrUndeclared,
		Message: "undeclared variable 'x'",
		Hint:    "declare it with 'var x: Type;'",
	}
	full := d.Format()
	if !strings.Contains(full, "KLX201") {
		t.Errorf("Format should contain error code, got: %s", full)
	}
	if !strings.Contains(full, "undeclared") {
		t.Errorf("Format should contain message, got: %s", full)
	}

	fullFmt := d.FormatFull()
	if !strings.Contains(fullFmt, "main.klx:10:5") {
		t.Errorf("FormatFull should contain location, got: %s", fullFmt)
	}
	if !strings.Contains(fullFmt, "help:") {
		t.Errorf("FormatFull should contain hint, got: %s", fullFmt)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func hasDiagCode(result *compiler.Result, code string) bool {
	for _, d := range result.Diagnostics {
		if d.Code == code {
			return true
		}
	}
	return false
}

func diagCodes(result *compiler.Result) []string {
	var codes []string
	for _, d := range result.Diagnostics {
		codes = append(codes, d.Code)
	}
	return codes
}
