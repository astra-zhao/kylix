package llvmgen_test

import (
	"strings"
	"testing"
)

// ===== Class codegen tests =====

func TestIR_ClassStructType(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TPoint = class
    X: Integer;
    Y: Integer;
  end;
begin
end.`)
	// Should emit a struct type with vtable ptr + 2 fields
	assertIRContains(t, ir, "%TPoint = type")
	assertIRContains(t, ir, "ptr,") // vtable pointer
}

func TestIR_ClassWithMethod(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TCounter = class
    Value: Integer;

    procedure Increment();
    begin
      Value := Value + 1;
    end;

    function GetValue(): Integer;
    begin
      result := Value;
    end;
  end;
begin
end.`)
	assertIRContains(t, ir, "%TCounter = type")
	assertIRContains(t, ir, "define void @TCounter_Increment")
	assertIRContains(t, ir, "define i64 @TCounter_GetValue")
	assertIRContains(t, ir, "ptr %self")
}

func TestIR_ClassVtable(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TAnimal = class
    function Speak(): String;
    begin
      result := 'generic sound';
    end;
  end;
begin
end.`)
	assertIRContains(t, ir, "@TAnimal_vtable = constant")
	assertIRContains(t, ir, "ptr @TAnimal_Speak")
}

func TestIR_ClassExternalMethodSkipped(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TWrapped = class
    function NativeMethod(): Integer; external;
  end;
begin
end.`)
	// External method must NOT generate a define
	if strings.Contains(ir, "define i64 @TWrapped_NativeMethod") {
		t.Error("external method should not emit LLVM define")
	}
}

func TestIR_ClassMultipleFields(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TPerson = class
    Age:  Integer;
    Name: String;
  end;
begin
end.`)
	assertIRContains(t, ir, "%TPerson = type { ptr, i64, ptr }")
}

func TestIR_ClassAndFunctionCoexist(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TBox = class
    Width: Integer;
    Height: Integer;
  end;

function MakeBox(w: Integer; h: Integer): Integer;
begin
  result := w * h;
end;

begin
  WriteLn(IntToStr(MakeBox(3, 4)));
end.`)
	assertIRContains(t, ir, "%TBox = type")
	assertIRContains(t, ir, "define i64 @MakeBox")
	assertIRContains(t, ir, "mul i64")
}
