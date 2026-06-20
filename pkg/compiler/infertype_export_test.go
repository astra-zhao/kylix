package compiler_test

import (
	"testing"

	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/compiler"
)

// findProbeRHS locates the RHS of `__probe := <expr>` in a parsed program.
func findProbeRHS(prog *ast.Program) ast.Expression {
	for _, stmt := range prog.Statements {
		if asgn, ok := stmt.(*ast.AssignmentStatement); ok {
			if id, ok := asgn.Name.(*ast.Identifier); ok && id.Value == "__probe" {
				return asgn.Value
			}
		}
	}
	return nil
}

// Tests for the exported InferType used by the REPL ':type' command.

// inferProbe parses a program containing `__probe := <expr>;` and returns the
// inferred type of <expr> via compiler.InferType.
func inferProbe(t *testing.T, decls, expr string) string {
	t.Helper()
	src := "program P;\n" + decls + "begin\n  __probe := " + expr + ";\nend.\n"
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse error: %s", errs[0])
	}
	// Find __probe assignment value.
	probeExpr := findProbeRHS(prog)
	if probeExpr == nil {
		t.Fatal("could not find __probe assignment")
	}
	return compiler.InferType(prog, probeExpr)
}

func TestInferType_Literals(t *testing.T) {
	cases := []struct{ expr, want string }{
		{"42", "Integer"},
		{"3.14", "Real"},
		{"'hello'", "String"},
		{"true", "Boolean"},
	}
	for _, tc := range cases {
		got := inferProbe(t, "", tc.expr)
		if got != tc.want {
			t.Errorf("InferType(%q) = %q, want %q", tc.expr, got, tc.want)
		}
	}
}

func TestInferType_Comparison(t *testing.T) {
	got := inferProbe(t, "", "1 < 2")
	if got != "Boolean" {
		t.Errorf("InferType('1 < 2') = %q, want Boolean", got)
	}
}

func TestInferType_Arithmetic(t *testing.T) {
	if got := inferProbe(t, "", "3 + 4"); got != "Integer" {
		t.Errorf("InferType('3 + 4') = %q, want Integer", got)
	}
	if got := inferProbe(t, "", "3.0 + 4"); got != "Real" {
		t.Errorf("InferType('3.0 + 4') = %q, want Real", got)
	}
}

func TestInferType_FunctionReturn(t *testing.T) {
	decls := "function GetAge(): Integer;\nbegin result := 30; end;\n"
	got := inferProbe(t, decls, "GetAge()")
	if got != "Integer" {
		t.Errorf("InferType('GetAge()') = %q, want Integer (from return type)", got)
	}
}

func TestInferType_StringConcat(t *testing.T) {
	got := inferProbe(t, "", "'a' + 'b'")
	if got != "String" {
		t.Errorf("InferType(string concat) = %q, want String", got)
	}
}

func TestInferType_Unknown(t *testing.T) {
	// An unknown identifier yields empty (caller falls back).
	got := inferProbe(t, "", "someUnknownThing")
	if got != "" {
		t.Errorf("InferType(unknown) = %q, want empty", got)
	}
}
