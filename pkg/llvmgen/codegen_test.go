package llvmgen_test

import (
	"strings"
	"testing"

	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/llvmgen"
)

func generateIR(t *testing.T, src string) string {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	ir, err := llvmgen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	return ir
}

func assertIRContains(t *testing.T, ir, substr string) {
	t.Helper()
	if !strings.Contains(ir, substr) {
		t.Errorf("expected IR to contain %q\nActual IR:\n%s", substr, ir)
	}
}

// ===== Module structure =====

func TestIR_ModuleHeader(t *testing.T) {
	ir := generateIR(t, `program hello; begin end.`)
	assertIRContains(t, ir, "source_filename")
	assertIRContains(t, ir, "target triple")
	assertIRContains(t, ir, "declare i32 @printf")
}

func TestIR_MainFunction(t *testing.T) {
	ir := generateIR(t, `program test; begin end.`)
	assertIRContains(t, ir, "define i32 @main()")
	assertIRContains(t, ir, "ret i32 0")
}

// ===== Integer expressions =====

func TestIR_IntegerLiteral(t *testing.T) {
	ir := generateIR(t, `program test; begin WriteLn(42); end.`)
	assertIRContains(t, ir, "add i64 0, 42")
}

func TestIR_IntegerArithmetic(t *testing.T) {
	ir := generateIR(t, `program test;
var x: Integer;
begin
  x := 3 + 4;
  x := 10 - 2;
  x := 6 * 7;
end.`)
	assertIRContains(t, ir, "add i64")
	assertIRContains(t, ir, "sub i64")
	assertIRContains(t, ir, "mul i64")
}

func TestIR_IntegerComparison(t *testing.T) {
	ir := generateIR(t, `program test;
var b: Boolean;
begin
  b := 3 > 2;
end.`)
	assertIRContains(t, ir, "icmp sgt i64")
}

// ===== Boolean =====

func TestIR_BooleanLiteral(t *testing.T) {
	ir := generateIR(t, `program test;
var b: Boolean;
begin
  b := true;
end.`)
	assertIRContains(t, ir, "add i1 0, 1")
}

func TestIR_BooleanNot(t *testing.T) {
	ir := generateIR(t, `program test;
var b: Boolean;
begin
  b := not true;
end.`)
	assertIRContains(t, ir, "xor i1")
}

// ===== Strings =====

func TestIR_StringConstant(t *testing.T) {
	ir := generateIR(t, `program test;
begin
  WriteLn('Hello, World!');
end.`)
	assertIRContains(t, ir, "Hello, World!")
	assertIRContains(t, ir, "call i32 @puts")
}

func TestIR_StringConstantDeclared(t *testing.T) {
	ir := generateIR(t, `program test;
begin
  WriteLn('hello');
end.`)
	// String constants are emitted at end of module
	assertIRContains(t, ir, `constant`)
	assertIRContains(t, ir, `hello`)
}

// ===== WriteLn variants =====

func TestIR_WriteLnInteger(t *testing.T) {
	ir := generateIR(t, `program test;
begin
  WriteLn(99);
end.`)
	assertIRContains(t, ir, "%lld")
	assertIRContains(t, ir, "@printf")
}

func TestIR_WriteLnBoolean(t *testing.T) {
	ir := generateIR(t, `program test;
begin
  WriteLn(true);
end.`)
	assertIRContains(t, ir, "true")
	assertIRContains(t, ir, "select i1")
}

// ===== Control flow =====

func TestIR_IfStatement(t *testing.T) {
	ir := generateIR(t, `program test;
begin
  if 1 > 0 then
    WriteLn('yes');
end.`)
	assertIRContains(t, ir, "br i1")
	assertIRContains(t, ir, "icmp sgt")
}

func TestIR_IfElseStatement(t *testing.T) {
	ir := generateIR(t, `program test;
begin
  if 1 > 0 then
    WriteLn('yes')
  else
    WriteLn('no');
end.`)
	assertIRContains(t, ir, "br i1")
	assertIRContains(t, ir, "yes")
	assertIRContains(t, ir, "no")
}

func TestIR_WhileLoop(t *testing.T) {
	ir := generateIR(t, `program test;
var i: Integer;
begin
  i := 0;
  while i < 5 do
  begin
    i := i + 1;
  end;
end.`)
	assertIRContains(t, ir, "icmp slt i64")
	assertIRContains(t, ir, "br label")
}

func TestIR_ForLoop(t *testing.T) {
	ir := generateIR(t, `program test;
var i: Integer;
begin
  for i := 1 to 10 do
    WriteLn(i);
end.`)
	assertIRContains(t, ir, "icmp sle i64")
	assertIRContains(t, ir, "add i64")
}

// ===== Functions =====

func TestIR_FunctionDecl(t *testing.T) {
	ir := generateIR(t, `program test;
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;
begin
end.`)
	assertIRContains(t, ir, "define i64 @Add(i64 %a, i64 %b)")
	assertIRContains(t, ir, "alloca i64")
}

func TestIR_ExternalFunctionSkipped(t *testing.T) {
	ir := generateIR(t, `program test;
function GoHelper(): Integer; external;
begin
  WriteLn(42);
end.`)
	// External function must NOT emit a define
	if strings.Contains(ir, "define i64 @GoHelper") {
		t.Error("external function should not be emitted as LLVM definition")
	}
}

// ===== IntToStr =====

func TestIR_IntToStr(t *testing.T) {
	ir := generateIR(t, `program test;
begin
  WriteLn(IntToStr(42));
end.`)
	assertIRContains(t, ir, "snprintf")
	assertIRContains(t, ir, "alloca [24 x i8]")
}
