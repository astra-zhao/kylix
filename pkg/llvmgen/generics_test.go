package llvmgen_test

import (
	"strings"
	"testing"
)

// ===== Generic monomorphization codegen tests (Milestone 2 Phase 3) =====

func TestIR_GenericTemplateNotEmitted(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TBox<T> = class
    Value: T;
  end;
begin
end.`)
	// The bare template should not produce a struct/vtable line.
	if strings.Contains(ir, "%TBox = type") {
		t.Errorf("template TBox<T> should not emit %%TBox = type; got:\n%s", ir)
	}
}

func TestIR_GenericInstantiationEmitsSpecialized(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TBox<T> = class
    Value: T;
  end;
var
  bi: TBox<Integer>;
begin
end.`)
	assertIRContains(t, ir, "%TBox_Integer = type")
	// Field T should be specialized to i64.
	assertIRContains(t, ir, "i64")
}

func TestIR_GenericMultipleInstantiations(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TBox<T> = class
    Value: T;
  end;
var
  bi: TBox<Integer>;
  bs: TBox<String>;
begin
end.`)
	assertIRContains(t, ir, "%TBox_Integer = type")
	assertIRContains(t, ir, "%TBox_String = type")
}

func TestIR_GenericConstructorCallEmitsMalloc(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TBox<T> = class
    Value: T;
  end;
var
  bi: TBox<Integer>;
begin
  bi := TBox<Integer>.Create;
end.`)
	// Constructor goes through emitConstructor which uses @malloc.
	assertIRContains(t, ir, "call ptr @malloc")
	// And references the specialized struct.
	assertIRContains(t, ir, "%TBox_Integer")
}

func TestIR_GenericMethodSpecialized(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TBox<T> = class
    Value: T;
    function Get(): T;
    begin
      result := 0;
    end;
  end;
var
  bi: TBox<Integer>;
begin
end.`)
	// The Get() return type T is substituted to i64 in the Integer specialization.
	assertIRContains(t, ir, "define i64 @TBox_Integer_Get")
}

func TestIR_NonGenericConstructorAlsoWorks(t *testing.T) {
	// Regression: plain TFoo.Create should still route through emitConstructor.
	ir := generateIR(t, `program test;
type
  TFoo = class
    Value: Integer;
  end;
var
  f: TFoo;
begin
  f := TFoo.Create;
end.`)
	assertIRContains(t, ir, "call ptr @malloc")
	assertIRContains(t, ir, "%TFoo")
}
