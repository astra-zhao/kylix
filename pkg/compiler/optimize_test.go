package compiler_test

import (
	"testing"

	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/compiler"
)

// Constant propagation + dead code elimination tests (v2.6.0 task 2).

func parseOpt(t *testing.T, src string) *ast.Program {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse error: %s", errs[0])
	}
	return prog
}

func TestOptimize_DeadCodeAfterReturn(t *testing.T) {
	src := `program Test;
function Foo(): Integer;
begin
  result := 42;
  return;
  WriteLn('unreachable');
  WriteLn('also unreachable');
end;
begin end.`

	prog := parseOpt(t, src)
	compiler.OptimizeProgram(prog)

	// Find Foo's body.
	for _, decl := range prog.Declarations {
		if fd, ok := decl.(*ast.FunctionDecl); ok && fd.Name == "Foo" {
			stmts := fd.Body.Statements
			// Should have: result := 42; return; (2 statements, not 4)
			if len(stmts) != 2 {
				t.Errorf("expected 2 statements after DCE, got %d", len(stmts))
			}
			// Last statement should be the return.
			if _, ok := stmts[1].(*ast.ReturnStatement); !ok {
				t.Error("expected return as last statement")
			}
			return
		}
	}
	t.Fatal("Foo function not found")
}

func TestOptimize_DeadCodeAfterRaise(t *testing.T) {
	src := `program Test;
function Bar(): Integer;
begin
  raise Exception.Create('error');
  WriteLn('unreachable');
end;
begin end.`

	prog := parseOpt(t, src)
	compiler.OptimizeProgram(prog)

	for _, decl := range prog.Declarations {
		if fd, ok := decl.(*ast.FunctionDecl); ok && fd.Name == "Bar" {
			stmts := fd.Body.Statements
			// Should have only the raise statement.
			if len(stmts) != 1 {
				t.Errorf("expected 1 statement after DCE, got %d", len(stmts))
			}
			if _, ok := stmts[0].(*ast.RaiseStatement); !ok {
				t.Error("expected raise as first statement")
			}
			return
		}
	}
	t.Fatal("Bar function not found")
}

func TestOptimize_NoDeadCode(t *testing.T) {
	src := `program Test;
function Baz(): Integer;
begin
  result := 1;
  result := result + 2;
  result := result * 3;
end;
begin end.`

	prog := parseOpt(t, src)
	compiler.OptimizeProgram(prog)

	for _, decl := range prog.Declarations {
		if fd, ok := decl.(*ast.FunctionDecl); ok && fd.Name == "Baz" {
			stmts := fd.Body.Statements
			// No terminator — all 3 statements should remain.
			if len(stmts) != 3 {
				t.Errorf("expected 3 statements (no DCE), got %d", len(stmts))
			}
			return
		}
	}
	t.Fatal("Baz function not found")
}

func TestOptimize_DeadCodeAfterExit(t *testing.T) {
	src := `program Test;
procedure Proc();
begin
  Exit;
  WriteLn('unreachable');
end;
begin end.`

	prog := parseOpt(t, src)
	compiler.OptimizeProgram(prog)

	for _, decl := range prog.Declarations {
		if fd, ok := decl.(*ast.FunctionDecl); ok && fd.Name == "Proc" {
			stmts := fd.Body.Statements
			// Exit + nothing after.
			if len(stmts) != 1 {
				t.Errorf("expected 1 statement (Exit only), got %d", len(stmts))
			}
			return
		}
	}
	t.Fatal("Proc not found")
}

func TestOptimize_DeadCodeAfterBreak(t *testing.T) {
	src := `program Test;
function Foo(): Integer;
begin
  while true do
  begin
    break;
    WriteLn('unreachable inside loop');
  end;
  result := 1;
end;
begin end.`

	prog := parseOpt(t, src)
	compiler.OptimizeProgram(prog)

	for _, decl := range prog.Declarations {
		if fd, ok := decl.(*ast.FunctionDecl); ok && fd.Name == "Foo" {
			// The function body has: while loop + result := 1
			// The while loop body should have only: break (1 stmt)
			for _, stmt := range fd.Body.Statements {
				if ws, ok := stmt.(*ast.WhileStatement); ok {
					loopStmts := ws.Body.Statements
					if len(loopStmts) != 1 {
						t.Errorf("expected 1 stmt in loop body (break only), got %d", len(loopStmts))
					}
				}
			}
			return
		}
	}
	t.Fatal("Foo not found")
}
