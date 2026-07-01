package llvmgen_test

import (
	"strings"
	"testing"
)

func assertInterpContains(t *testing.T, ir, substr string) {
	t.Helper()
	if !strings.Contains(ir, substr) {
		t.Errorf("expected IR to contain %q\nActual IR:\n%s", substr, ir)
	}
}

func TestInterp_AllocatesBuffer(t *testing.T) {
	ir := generateExcIR(t, `program p;
var n: Integer;
begin
  n := 5;
  WriteLn('n=${n}');
end.`)
	assertInterpContains(t, ir, "call ptr @malloc(i64 256)")
	assertInterpContains(t, ir, "store i8 0, ptr") // NUL init
}

func TestInterp_PlainLiteralSkipsInterp(t *testing.T) {
	ir := generateExcIR(t, `program p;
begin
  WriteLn('hello world');
end.`)
	// A plain string literal (no interpolation) does not allocate a buffer.
	if strings.Contains(ir, "call ptr @malloc(i64 256)") {
		t.Errorf("plain string literal should not use interpolation buffer\nIR:\n%s", ir)
	}
}

func TestInterp_IntegerPartUsesSnprintf(t *testing.T) {
	ir := generateExcIR(t, `program p;
var n: Integer;
begin
  n := 42;
  WriteLn('val=${n}');
end.`)
	assertInterpContains(t, ir, "call i32 (ptr, i64, ptr, ...) @snprintf")
	assertInterpContains(t, ir, "call i64 @strlen")
	// "%ld" format constant emitted.
	assertInterpContains(t, ir, `c"%ld\00"`)
}

func TestInterp_MultipleParts(t *testing.T) {
	ir := generateExcIR(t, `program p;
var a: Integer; var b: Integer;
begin
  a := 1; b := 2;
  WriteLn('a=${a} b=${b}');
end.`)
	// Two integer parts → two snprintf calls.
	if c := strings.Count(ir, "call i32 (ptr, i64, ptr, ...) @snprintf"); c != 2 {
		t.Errorf("expected 2 snprintf calls, got %d\nIR:\n%s", c, ir)
	}
}

func TestInterp_NoLongerUnhandled(t *testing.T) {
	// Previously StringInterpolation fell to the default "unhandled expr" stub.
	ir := generateExcIR(t, `program p;
var n: Integer;
begin
  n := 1;
  WriteLn('x${n}y');
end.`)
	if strings.Contains(ir, "unhandled expr") {
		t.Errorf("StringInterpolation still unhandled\nIR:\n%s", ir)
	}
}

func TestVarDecl_InitializerIsEmitted(t *testing.T) {
	// var x: Integer = 42 must store the initializer value, not just zero it.
	ir := generateExcIR(t, `program p;
var
  count: Integer = 42;
begin
  WriteLn(count);
end.`)
	if strings.Contains(ir, "add i64 0, 42") == false && strings.Contains(ir, "store i64 42") == false {
		t.Errorf("expected initializer value 42 to be stored\nIR:\n%s", ir)
	}
}

func TestVarDecl_InferredStringTypeAllocatesPtr(t *testing.T) {
	// var message := 'Hello' must allocate a ptr slot, not default to i64 —
	// previously this caused an llc type mismatch ('store i64 %t, ptr %v_..._int').
	ir := generateExcIR(t, `program p;
begin
  var message := 'Hello';
  WriteLn(message);
end.`)
	assertInterpContains(t, ir, "%v_message_str = alloca ptr")
	if strings.Contains(ir, "%v_message_int") {
		t.Errorf("inferred string var should not allocate as i64\nIR:\n%s", ir)
	}
}

func TestWriteLnMulti_BooleanArgPrintsTrueFalse(t *testing.T) {
	// WriteLn('flag: ', someBool) must append "true"/"false", not silently drop it.
	ir := generateExcIR(t, `program p;
var flag: Boolean;
begin
  flag := true;
  WriteLn('flag: ', flag);
end.`)
	assertInterpContains(t, ir, `c"true\00"`)
	assertInterpContains(t, ir, `c"false\00"`)
	assertInterpContains(t, ir, "select i1")
}

func TestMultiReturn_StructTypeEmitted(t *testing.T) {
	ir := generateExcIR(t, `program p;
function DivMod(a: Integer; b: Integer): (Integer, Integer);
begin
  result := (a div b, a mod b);
end;
begin
end.`)
	assertInterpContains(t, ir, "%__ret_DivMod = type { i64, i64 }")
	assertInterpContains(t, ir, "define %__ret_DivMod @DivMod")
}

func TestMultiReturn_TupleBuildUsesInsertvalue(t *testing.T) {
	ir := generateExcIR(t, `program p;
function Pair(x: Integer; y: Integer): (Integer, Integer);
begin
  result := (x, y);
end;
begin
end.`)
	assertInterpContains(t, ir, "insertvalue %__ret_Pair undef")
	assertInterpContains(t, ir, "insertvalue %__ret_Pair %")
	assertInterpContains(t, ir, "store %__ret_Pair")
}

func TestMultiReturn_DestructureUsesExtractvalue(t *testing.T) {
	ir := generateExcIR(t, `program p;
function GetPair(): (Integer, Integer);
begin
  result := (3, 5);
end;
var a, b: Integer;
begin
  (a, b) := GetPair();
end.`)
	assertInterpContains(t, ir, "call %__ret_GetPair @GetPair()")
	assertInterpContains(t, ir, "extractvalue %__ret_GetPair")
}
