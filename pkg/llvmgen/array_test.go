package llvmgen_test

import (
	"strings"
	"testing"

	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/llvmgen"
)

// generateIRForArray is a helper to compile a Kylix source to LLVM IR text.
func generateIRForArray(t *testing.T, src string) string {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	ir, err := llvmgen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen: %v", err)
	}
	return ir
}

func TestIR_StaticArray_Alloca(t *testing.T) {
	ir := generateIRForArray(t, `
program ArrayTest;
var nums: array[1..5] of Integer;
begin
end.`)
	if !strings.Contains(ir, "alloca [5 x i64]") {
		t.Errorf("expected 'alloca [5 x i64]' in IR\n%s", ir)
	}
	if !strings.Contains(ir, "zeroinitializer") {
		t.Error("expected zero initialization of array")
	}
}

func TestIR_StaticArray_IndexAssign(t *testing.T) {
	ir := generateIRForArray(t, `
program ArrayAssign;
var nums: array[1..3] of Integer;
begin
  nums[1] := 42;
  nums[2] := 17;
end.`)
	// Should have a GEP instruction and a store
	if !strings.Contains(ir, "getelementptr inbounds [3 x i64]") {
		t.Errorf("expected GEP into [3 x i64]\n%s", ir)
	}
	if !strings.Contains(ir, "store i64") {
		t.Error("expected store i64 for element assignment")
	}
}

func TestIR_StaticArray_IndexRead(t *testing.T) {
	ir := generateIRForArray(t, `
program ArrayRead;
var
  nums: array[1..3] of Integer;
  x: Integer;
begin
  x := nums[2];
end.`)
	if !strings.Contains(ir, "load i64") {
		t.Error("expected load i64 for element read")
	}
}

func TestIR_StaticArray_PascalIndexAdjust(t *testing.T) {
	// Pascal arrays are 1-indexed; LLVM should subtract 1 for the GEP.
	ir := generateIRForArray(t, `
program PascalIndex;
var arr: array[1..5] of Integer;
begin
  arr[3] := 99;
end.`)
	if !strings.Contains(ir, "sub i64") {
		t.Error("expected 'sub i64' for Pascal 1-based → LLVM 0-based index")
	}
}

func TestIR_DynamicArray_StructLayout(t *testing.T) {
	ir := generateIRForArray(t, `
program DynArr;
var nums: array of Integer;
begin
end.`)
	if !strings.Contains(ir, "{ ptr, i64, i64 }") {
		t.Errorf("expected slice struct { ptr, i64, i64 }\n%s", ir)
	}
}

func TestIR_ArrayInLoop(t *testing.T) {
	// Full end-to-end: fill array, sum it, print result.
	ir := generateIRForArray(t, `
program ArrayLoop;
var
  nums: array[1..5] of Integer;
  i: Integer;
begin
  i := 1;
  while i <= 5 do
  begin
    nums[i] := i;
    i := i + 1;
  end;
end.`)
	if !strings.Contains(ir, "alloca [5 x i64]") {
		t.Error("expected alloca [5 x i64]")
	}
	if !strings.Contains(ir, "getelementptr inbounds [5 x i64]") {
		t.Error("expected GEP into array in loop body")
	}
}
