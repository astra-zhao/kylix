package compiler_test

import (
	"testing"

	"kylix/pkg/compiler"
)

// TestErrorRecovery_AllSemanticErrors verifies that all semantic errors from
// different checking phases are reported in a single compilation pass.
func TestErrorRecovery_AllSemanticErrors(t *testing.T) {
	result := compileTC(t, `program Test;
type
  IFoo = interface
    procedure Bar();
    procedure Baz();
  end;
  TFoo = class implements IFoo
  end;
var x: Integer;
begin
  x := 'hello';
end.`)

	if result.Success {
		t.Fatal("expected failure")
	}

	// Should report: KLX301 (missing Bar), KLX301 (missing Baz), KLX101 (type mismatch)
	if len(result.Diagnostics) < 3 {
		t.Errorf("expected >= 3 diagnostics, got %d: %v", len(result.Diagnostics), result.Diagnostics)
	}

	codes := make(map[string]int)
	for _, d := range result.Diagnostics {
		codes[d.Code]++
	}
	if codes[compiler.ErrMissingMethod] < 2 {
		t.Errorf("expected >= 2 KLX301 errors, got %d", codes[compiler.ErrMissingMethod])
	}
	if codes[compiler.ErrTypeMismatch] < 1 {
		t.Errorf("expected >= 1 KLX101 error, got %d", codes[compiler.ErrTypeMismatch])
	}
}

// TestErrorRecovery_InterfaceAndArity verifies interface + arity errors both reported.
func TestErrorRecovery_InterfaceAndArity(t *testing.T) {
	result := compileTC(t, `program Test;
type
  IAnimal = interface
    procedure Speak();
  end;
  TDog = class implements IAnimal
  end;
function Add(a: Integer; b: Integer): Integer;
begin result := a + b; end;
begin
  WriteLn(Add(1, 2, 3));
end.`)

	if result.Success {
		t.Fatal("expected failure")
	}

	codes := make(map[string]int)
	for _, d := range result.Diagnostics {
		codes[d.Code]++
	}
	if codes[compiler.ErrMissingMethod] == 0 {
		t.Error("expected KLX301 (missing Speak)")
	}
	if codes[compiler.ErrWrongArity] == 0 {
		t.Error("expected KLX202 (wrong arity)")
	}
}

// TestErrorRecovery_TypeAndUndeclared verifies type mismatch + undeclared both reported.
func TestErrorRecovery_TypeAndUndeclared(t *testing.T) {
	result := compileTC(t, `program Test;
var x: Integer;
begin
  x := 'bad type';
  unknownVar := 42;
end.`)

	if result.Success {
		t.Fatal("expected failure")
	}

	if len(result.Diagnostics) < 2 {
		t.Errorf("expected >= 2 diagnostics, got %d", len(result.Diagnostics))
	}
}

// TestErrorRecovery_ParseErrorStopsFurther verifies that parse errors stop
// semantic checks (avoiding noisy false positives from incomplete AST).
func TestErrorRecovery_ParseErrorStopsFurther(t *testing.T) {
	// Deliberately broken syntax
	result := compileTC(t, `program Test;
begin
  := 42;
end.`)

	if result.Success {
		t.Fatal("expected failure")
	}

	// Should only have parse errors, not type/semantic errors
	for _, d := range result.Diagnostics {
		if d.Code != "" && d.Code != compiler.ErrParseGeneric &&
			d.Code != compiler.ErrUnexpectedToken && d.Code != compiler.ErrMissingToken {
			t.Errorf("unexpected non-parse error after parse failure: %s %s", d.Code, d.Message)
		}
	}
}
