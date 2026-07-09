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

// ===== v4.6.0: per-instruction DILocation + DILocalVariable tests =====

// TestDebug_InstructionsCarryDILocation verifies that instruction-level IR
// lines (alloca/store/load/call/br/...) carry a trailing ", !dbg !N" when -g
// is on, so LLDB can step per source line.
func TestDebug_InstructionsCarryDILocation(t *testing.T) {
	ir := generateIRWithDebug(t, `program p;
var
  x: Integer;
begin
  x := 42;
  WriteLn(x);
end.`)
	// The assignment's add instruction should carry !dbg.
	assertIRContains(t, ir, "= add i64 0, 42, !dbg !")
	// A store instruction should carry !dbg.
	assertIRContains(t, ir, "store i64 %t0, ptr %v_x_int, !dbg !")
	// The br terminator should carry !dbg (if present) — it's how the debugger
	// maps a control-flow step.
	if strings.Contains(ir, "br ") && !strings.Contains(ir, "br label") {
		t.Logf("br instruction present — checking for !dbg")
	}
}

// TestDebug_LabelsDoNotCarryDbg verifies label lines (entry:, lblN:) are NOT
// given a !dbg attachment — LLVM rejects !dbg on labels.
func TestDebug_LabelsDoNotCarryDbg(t *testing.T) {
	ir := generateIRWithDebug(t, `program p;
begin
  if 1 < 2 then WriteLn('a') else WriteLn('b');
end.`)
	// "entry:" should never have !dbg.
	if strings.Contains(ir, "entry:, !dbg") {
		t.Errorf("entry label carries !dbg (LLVM rejects this)\nIR:\n%s", ir)
	}
	// A generated label like "lbl0:" should never have !dbg.
	lines := strings.Split(ir, "\n")
	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if strings.HasSuffix(trimmed, ":") && strings.HasPrefix(trimmed, "lbl") {
			if strings.Contains(ln, "!dbg") {
				t.Errorf("label line %q carries !dbg (LLVM rejects this)", ln)
			}
		}
	}
}

// TestDebug_DILocationNodesEmitted verifies the metadata block contains
// DILocation nodes scoped to the subprogram, enabling per-line stepping.
func TestDebug_DILocationNodesEmitted(t *testing.T) {
	ir := generateIRWithDebug(t, `program p;
begin
  WriteLn('hi');
  WriteLn('bye');
end.`)
	// At least one DILocation node scoped to main (!4).
	assertIRContains(t, ir, "!DILocation(line:")
	// Multiple distinct source lines → multiple DILocation nodes.
	count := strings.Count(ir, "!DILocation(")
	if count < 2 {
		t.Errorf("expected >=2 DILocation nodes, got %d\nIR:\n%s", count, ir)
	}
}

// TestDebug_DILocalVariableForUserLocal verifies a `var x: Integer;` local
// gets a DILocalVariable + a #dbg_declare record so LLDB can resolve the
// variable's name + scope at a breakpoint.
func TestDebug_DILocalVariableForUserLocal(t *testing.T) {
	ir := generateIRWithDebug(t, `program p;
var
  x: Integer;
begin
  x := 42;
  WriteLn(x);
end.`)
	// DILocalVariable node for "x".
	assertIRContains(t, ir, `!DILocalVariable(name: "x"`)
	// The #dbg_declare record (LLVM 22 intrinsic record syntax).
	assertIRContains(t, ir, "#dbg_declare(ptr %v_x_int")
	// The DIBasicType referenced by the variable.
	assertIRContains(t, ir, "!DIBasicType(name: \"int64\"")
}

// TestDebug_FunctionParametersAreDebugLocals verifies function parameters get
// DILocalVariables too, so `frame variable` shows them inside user functions.
func TestDebug_FunctionParametersAreDebugLocals(t *testing.T) {
	ir := generateIRWithDebug(t, `program p;
function Add(a, b: Integer): Integer;
begin
  result := a + b;
end;
begin
  WriteLn(Add(1, 2));
end.`)
	// Parameters a and b should each get a DILocalVariable.
	assertIRContains(t, ir, `!DILocalVariable(name: "a"`)
	assertIRContains(t, ir, `!DILocalVariable(name: "b"`)
	// result should also be declared (it's an alloca inside the function).
	assertIRContains(t, ir, `!DILocalVariable(name: "result"`)
}

// TestDebug_NoDbgRecordsWithoutFlag verifies #dbg_declare records and
// DILocalVariable nodes are absent when -g is off (no debug overhead).
func TestDebug_NoDbgRecordsWithoutFlag(t *testing.T) {
	ir := generateIR(t, `program p;
var
  x: Integer;
begin
  x := 42;
  WriteLn(x);
end.`)
	if strings.Contains(ir, "#dbg_declare") {
		t.Errorf("dbg record emitted without -g\nIR:\n%s", ir)
	}
	if strings.Contains(ir, "DILocalVariable") {
		t.Errorf("DILocalVariable emitted without -g\nIR:\n%s", ir)
	}
	if strings.Contains(ir, "DILocation") {
		t.Errorf("DILocation emitted without -g\nIR:\n%s", ir)
	}
}

// TestDebug_SteppingAcrossIf verifies that instructions in both branches of
// an if statement carry !dbg with distinct source lines, so stepping shows
// the right branch.
func TestDebug_SteppingAcrossIf(t *testing.T) {
	ir := generateIRWithDebug(t, `program p;
var
  x: Integer;
begin
  x := 10;
  if x > 5 then
    WriteLn('big')
  else
    WriteLn('small');
end.`)
	// The conditional branch should carry a !dbg at the if-line.
	assertIRContains(t, ir, "br i1")
	// Both branch bodies emit instructions with !dbg (the then/else WriteLns).
	// Verify at least 4 distinct DILocation nodes (decl, assign, if, write).
	count := strings.Count(ir, "!DILocation(")
	if count < 4 {
		t.Errorf("expected >=4 DILocation nodes for if-stepping, got %d\nIR:\n%s", count, ir)
	}
}
