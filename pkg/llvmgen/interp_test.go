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
	assertInterpContains(t, ir, "call i32 @snprintf")
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
	if c := strings.Count(ir, "call i32 @snprintf"); c != 2 {
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
