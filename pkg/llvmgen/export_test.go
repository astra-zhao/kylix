package llvmgen_test

import (
	"strings"
	"testing"

	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/llvmgen"
)

func TestExport_DefaultSymbol(t *testing.T) {
	src := `program test;
[Export]
function Add(a, b: Integer): Integer;
begin
  result := a + b;
end;
begin
end.`
	ir := generateIR(t, src)
	assertIRContains(t, ir, "define i64 @Add(i64 %a, i64 %b)")
	assertIRContains(t, ir, "define void @kylix_free(ptr %p)")
	assertIRContains(t, ir, "call void @free(ptr %p)")
}

func TestExport_CustomSymbol(t *testing.T) {
	src := `program test;
[Export('calc_multiply')]
function Multiply(a, b: Integer): Integer;
begin
  result := a * b;
end;

function CallInternal(): Integer;
begin
  result := Multiply(3, 4);
end;
begin
end.`
	ir := generateIR(t, src)
	assertIRContains(t, ir, "define i64 @calc_multiply(i64 %a, i64 %b)")
	// Internal call should resolve to @calc_multiply
	assertIRContains(t, ir, "call i64 @calc_multiply(")
	assertIRContains(t, ir, "define void @kylix_free(ptr %p)")
}

func TestExport_SharedMode_GlobalCtors(t *testing.T) {
	src := `program mylib;
var initialized: Integer;
[Export]
function GetVal(): Integer;
begin
  result := initialized;
end;
begin
  initialized := 42;
end.`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	ir, err := llvmgen.GenerateWithOpts(prog, "mylib.klx", llvmgen.CompileOpts{Shared: true})
	if err != nil {
		t.Fatalf("GenerateWithOpts failed: %v", err)
	}

	// In shared mode, @main should NOT be emitted
	if strings.Contains(ir, "define i32 @main(") || strings.Contains(ir, "define i32 @main()") {
		t.Errorf("expected no @main in shared library mode, got:\n%s", ir)
	}

	// Should emit @__kylix_lib_init and @llvm.global_ctors
	assertIRContains(t, ir, "define void @__kylix_lib_init()")
	assertIRContains(t, ir, "@llvm.global_ctors = appending global [1 x { i32, ptr, ptr }] [{ i32, ptr, ptr } { i32 65535, ptr @__kylix_lib_init, ptr null }]")
	assertIRContains(t, ir, "define void @kylix_free(ptr %p)")
}

func TestTriple_MobileTargets(t *testing.T) {
	cases := []struct {
		target   string
		expected string
	}{
		{"android/arm64", "aarch64-linux-android30"},
		{"android/amd64", "x86_64-linux-android30"},
		{"ios/arm64", "arm64-apple-ios16.0.0"},
		{"ios/simulator-arm64", "arm64-apple-ios16.0.0-simulator"},
	}

	src := `program test; begin end.`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()

	for _, tc := range cases {
		ir, err := llvmgen.GenerateWithOpts(prog, "test.klx", llvmgen.CompileOpts{Target: tc.target})
		if err != nil {
			t.Fatalf("target %s failed: %v", tc.target, err)
		}
		expectedLine := `target triple = "` + tc.expected + `"`
		if !strings.Contains(ir, expectedLine) {
			t.Errorf("for target %s, expected %q, got:\n%s", tc.target, expectedLine, ir)
		}
	}
}

func TestNDK_ProbeGraceful(t *testing.T) {
	// Should not panic when no NDK is configured
	_ = llvmgen.FindAndroidNdk()
}
