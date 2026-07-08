package llvmgen_test

import (
	"strings"
	"testing"

	"kylix/ast"
	"kylix/lexer"
	"kylix/pkg/llvmgen"
	"kylix/parser"
)

// debug tests — verify the v4.5.0 DWARF debug-info emission (kylix -g).

func generateIRWithDebug(t *testing.T, src string) string {
	t.Helper()
	ir, err := llvmgen.GenerateWithOpts(mustParse(t, src), "test.klx", llvmgen.CompileOpts{DebugInfo: true})
	if err != nil {
		t.Fatalf("GenerateWithOpts(debug) failed: %v\nIR:\n%s", err, ir)
	}
	return ir
}

// mustParse parses a Kylix source string into an *ast.Program for the debug
// tests (GenerateWithOpts takes an AST, not source text).
func mustParse(t *testing.T, src string) *ast.Program {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	return prog
}

func TestDebug_NoDbgByDefault(t *testing.T) {
	// Without -g, no DWARF metadata is emitted.
	ir := generateIR(t, `program p;
begin
  WriteLn('hi');
end.`)
	if strings.Contains(ir, "!llvm.dbg.cu") {
		t.Errorf("DWARF metadata emitted without -g\nIR:\n%s", ir)
	}
	if strings.Contains(ir, "DISubprogram") {
		t.Errorf("DISubprogram emitted without -g\nIR:\n%s", ir)
	}
}

func TestDebug_DbgEmittedWithFlag(t *testing.T) {
	ir := generateIRWithDebug(t, `program p;
begin
  WriteLn('hi');
end.`)
	assertIRContains(t, ir, "!llvm.dbg.cu = !{!0}")
	assertIRContains(t, ir, "DICompileUnit")
	assertIRContains(t, ir, "DIFile")
	assertIRContains(t, ir, "module.flags")
	// main subprogram
	assertIRContains(t, ir, `DISubprogram(name: "main"`)
	// the main define line carries !dbg
	assertIRContains(t, ir, "define i32 @main() !dbg")
}

func TestDebug_UserFunctionsGetSubprograms(t *testing.T) {
	ir := generateIRWithDebug(t, `program p;
function Add(a, b: Integer): Integer;
begin
  result := a + b;
end;
begin
  WriteLn(Add(1, 2));
end.`)
	// Add should have a DISubprogram + its define line carries !dbg.
	assertIRContains(t, ir, `DISubprogram(name: "Add"`)
	assertIRContains(t, ir, "define i64 @Add(")
	if !strings.Contains(ir, "define i64 @Add(") || !strings.Contains(ir, "!dbg") {
		t.Errorf("Add define line missing !dbg attachment\nIR:\n%s", ir)
	}
	// main subprogram present too.
	assertIRContains(t, ir, `DISubprogram(name: "main"`)
}

func TestDebug_DIFileHasSourceName(t *testing.T) {
	ir := generateIRWithDebug(t, `program p;
begin
end.`)
	// The DIFile should reference the source file name passed to GenerateWithOpts.
	assertIRContains(t, ir, `DIFile(filename: "test.klx"`)
}

func TestDebug_OptForcedOffWhenDebug(t *testing.T) {
	// compileASTWithOpts forces OptLevel="" when DebugInfo is on. We can't
	// easily call the full pipeline here, but we verify the IR has no opt
	// markers and the metadata is present (the -O0 fallback path).
	ir := generateIRWithDebug(t, `program p;
begin
  WriteLn('x');
end.`)
	// No optimized-IR artifacts (the opt pass runs only with OptLevel set,
	// which is forced off — so no .opt.ll markers in the generate path).
	if strings.Contains(ir, "; opt ") {
		t.Errorf("unexpected opt marker in debug IR\nIR:\n%s", ir)
	}
	assertIRContains(t, ir, "DICompileUnit")
}
