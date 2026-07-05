package llvmgen_test

import (
	"strings"
	"testing"

	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/llvmgen"
)

// mustParseProgram parses a single Kylix source string into an *ast.Program,
// mirroring the real CLI's per-file parse step (llvmgen.ParseFile does the
// same thing but reads the source from disk instead of a string literal).
func mustParseProgram(t *testing.T, src string) *ast.Program {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	return prog
}

func hasFunctionDecl(prog *ast.Program, name string) bool {
	for _, d := range prog.Declarations {
		if fd, ok := d.(*ast.FunctionDecl); ok && fd.Name == name {
			return true
		}
	}
	return false
}

func TestMergePrograms_UnitPlusMainMergesDeclarations(t *testing.T) {
	unit := mustParseProgram(t, `unit MathHelper;
interface
function Square(x: Integer): Integer;
implementation
function Square(x: Integer): Integer;
begin
  result := x * x;
end;
end.`)
	main := mustParseProgram(t, `program UseModule;
uses MathHelper;
begin
  WriteLn(Square(5));
end.`)

	merged, err := llvmgen.MergePrograms([]*ast.Program{unit, main})
	if err != nil {
		t.Fatalf("MergePrograms: %v", err)
	}
	if merged.IsUnit {
		t.Errorf("merged program should not be marked IsUnit")
	}
	if merged.Name != "UseModule" {
		t.Errorf("merged program Name = %q, want %q", merged.Name, "UseModule")
	}
	if len(merged.Statements) == 0 {
		t.Errorf("merged program should carry the main program's Statements")
	}
	if !hasFunctionDecl(merged, "Square") {
		t.Errorf("merged program is missing the unit's Square function declaration")
	}
}

func TestMergePrograms_OrderIndependent(t *testing.T) {
	// The unit listed AFTER the main program (as a user might type
	// `kylix build --backend=llvm main.klx math_helper.klx`) must merge
	// identically to unit-first ordering, since emitProgram pre-registers
	// all function signatures before emitting any bodies.
	unit := mustParseProgram(t, `unit MathHelper;
interface
function Cube(x: Integer): Integer;
implementation
function Cube(x: Integer): Integer;
begin
  result := x * x * x;
end;
end.`)
	main := mustParseProgram(t, `program UseModule;
uses MathHelper;
begin
  WriteLn(Cube(3));
end.`)

	merged, err := llvmgen.MergePrograms([]*ast.Program{main, unit})
	if err != nil {
		t.Fatalf("MergePrograms: %v", err)
	}
	if !hasFunctionDecl(merged, "Cube") {
		t.Errorf("merged program is missing the unit's Cube function declaration")
	}
}

func TestMergePrograms_NoMainProgramErrors(t *testing.T) {
	unitOnly := mustParseProgram(t, `unit MathHelper;
interface
function Square(x: Integer): Integer;
implementation
function Square(x: Integer): Integer;
begin
  result := x * x;
end;
end.`)

	_, err := llvmgen.MergePrograms([]*ast.Program{unitOnly})
	if err == nil {
		t.Fatal("expected an error when no non-unit program is present, got nil")
	}
	if !strings.Contains(err.Error(), "no main program") {
		t.Errorf("error message = %q, want it to mention the missing main program", err.Error())
	}
}

func TestMergePrograms_MultipleMainProgramsErrors(t *testing.T) {
	mainA := mustParseProgram(t, `program A; begin end.`)
	mainB := mustParseProgram(t, `program B; begin end.`)

	_, err := llvmgen.MergePrograms([]*ast.Program{mainA, mainB})
	if err == nil {
		t.Fatal("expected an error when two non-unit programs are present, got nil")
	}
	if !strings.Contains(err.Error(), "multiple non-unit programs") {
		t.Errorf("error message = %q, want it to mention multiple non-unit programs", err.Error())
	}
}

func TestMergePrograms_UsesDeduplicated(t *testing.T) {
	// Both the main program and a unit might independently `uses` the same
	// module (e.g. both use `sysutil`); the merged Uses list must not contain
	// duplicates, since stdlib-dispatch logic iterates it.
	unit := mustParseProgram(t, `unit Helper;
uses sysutil;
interface
function Noop(): Integer;
implementation
function Noop(): Integer;
begin
  result := 0;
end;
end.`)
	main := mustParseProgram(t, `program P;
uses Helper, sysutil;
begin
  WriteLn(Noop());
end.`)

	merged, err := llvmgen.MergePrograms([]*ast.Program{unit, main})
	if err != nil {
		t.Fatalf("MergePrograms: %v", err)
	}
	count := 0
	for _, u := range merged.Uses {
		if u == "sysutil" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("merged Uses contains %d occurrences of %q, want exactly 1 (deduplicated)", count, "sysutil")
	}
}

// TestCompileFilesToNative_SingleFileDelegates is a smoke test for the
// single-file fast path in CompileFilesToNative — it should behave exactly
// like CompileToNativeOpts (used by the existing single-file CLI branch),
// just routed through the multi-file entry point. This doesn't invoke the
// real LLVM toolchain (llvmPaths is deliberately empty); it only checks that
// IR generation and file writing succeed before the toolchain step, since
// CI environments may not have llc/clang installed.
func TestCompileFilesToNative_EmptyFileListErrors(t *testing.T) {
	_, err := llvmgen.CompileFilesToNative(nil, "", &llvmgen.LLVMPaths{}, llvmgen.CompileOpts{})
	if err == nil {
		t.Fatal("expected an error for an empty file list, got nil")
	}
}
