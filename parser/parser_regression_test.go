package parser_test

import (
	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
	"testing"
)

// ── Generic Type Instantiation ────────────────────────────────────────────────

func TestParseGenericInstantiation(t *testing.T) {
	input := `program Test;
var box: TBox<Integer>;
begin
end.`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}

	if len(prog.Declarations) < 1 {
		t.Fatalf("expected 1 declaration, got %d", len(prog.Declarations))
	}

	vd1 := prog.Declarations[0].(*ast.VarDecl)
	gt1, ok := vd1.Type.(*ast.GenericType)
	if !ok {
		t.Fatalf("expected GenericType for TBox<Integer>, got %T", vd1.Type)
	}
	if gt1.Base != "TBox" {
		t.Errorf("expected Base=TBox, got %s", gt1.Base)
	}
	if len(gt1.TypeParams) != 1 {
		t.Errorf("expected 1 type param, got %d", len(gt1.TypeParams))
	}
}

func TestParseGenericTwoParams(t *testing.T) {
	input := `program Test;
type TPair<T1, T2> = class First: T1; Second: T2; end;
begin end.`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}

	if len(prog.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(prog.Declarations))
	}
	td := prog.Declarations[0].(*ast.TypeDecl)
	cd, ok := td.Type.(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", td.Type)
	}
	if len(cd.TypeParams) != 2 {
		t.Errorf("expected 2 type params, got %d", len(cd.TypeParams))
	}
}

// ── Multi-Return Value Syntax ─────────────────────────────────────────────────

func TestParseMultiReturnFunction(t *testing.T) {
	input := `program Test;
function Swap(a: Integer; b: Integer): (Integer, Integer);
begin
  result := (b, a);
end;
begin
end.`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}

	if len(prog.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(prog.Declarations))
	}

	fd := prog.Declarations[0].(*ast.FunctionDecl)
	if fd.Name != "Swap" {
		t.Errorf("expected Swap, got %s", fd.Name)
	}

	if len(fd.ReturnTypes) != 2 {
		t.Fatalf("expected 2 return types, got %d", len(fd.ReturnTypes))
	}
}

func TestParseMultiReturnAssignment(t *testing.T) {
	input := `program Test;
function Pair(): (Integer, Integer);
begin result := (1, 2); end;

begin
  var x, y := Pair();
  WriteLn(x);
end.`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}

	if len(prog.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}

	// First statement: var x, y := Pair();
	vd, ok := prog.Statements[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected VarDecl, got %T", prog.Statements[0])
	}
	if len(vd.Names) != 2 {
		t.Errorf("expected 2 names, got %d", len(vd.Names))
	}
	if vd.Names[0] != "x" || vd.Names[1] != "y" {
		t.Errorf("expected [x, y], got %v", vd.Names)
	}
}

func TestParseTupleReturn(t *testing.T) {
	input := `program Test;
function Foo(): (Integer, String);
begin
  return (42, 'hello');
end;
begin
end.`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}

	fd := prog.Declarations[0].(*ast.FunctionDecl)
	if fd.Body == nil || len(fd.Body.Statements) == 0 {
		t.Fatal("expected function body with statements")
	}

	ret, ok := fd.Body.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", fd.Body.Statements[0])
	}

	tuple, ok := ret.Value.(*ast.TupleLiteral)
	if !ok {
		t.Fatalf("expected TupleLiteral, got %T", ret.Value)
	}
	if len(tuple.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(tuple.Elements))
	}
}

// ── External Function Declarations ───────────────────────────────────────────

func TestParseExternalFunction(t *testing.T) {
	input := `
unit foo;
function Bar(x: Integer): String; external;
end.`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	if len(prog.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(prog.Declarations))
	}

	fn, ok := prog.Declarations[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected FunctionDecl, got %T", prog.Declarations[0])
	}
	if fn.Name != "Bar" {
		t.Errorf("expected name 'Bar', got %q", fn.Name)
	}
	if !fn.IsExternal {
		t.Error("expected IsExternal=true")
	}
	if fn.Body != nil {
		t.Error("external function should have nil body")
	}
}

func TestParseMultipleExternalFunctions(t *testing.T) {
	input := `
unit myunit;
function Foo(): String; external;
function Bar(x: Integer): Boolean; external;
procedure Baz(s: String); external;
end.`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(prog.Declarations) != 3 {
		t.Fatalf("expected 3 declarations, got %d", len(prog.Declarations))
	}
	for i, decl := range prog.Declarations {
		fn, ok := decl.(*ast.FunctionDecl)
		if !ok {
			t.Fatalf("decl[%d]: expected FunctionDecl", i)
		}
		if !fn.IsExternal {
			t.Errorf("decl[%d] (%s): expected IsExternal=true", i, fn.Name)
		}
	}
}

func TestParseExternalAtEndOfFile(t *testing.T) {
	// Regression: external at end of file (after all other declarations) must not error
	input := `
unit myunit;
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;
function FromGo(): String; external;
end.`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(prog.Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(prog.Declarations))
	}

	ext := prog.Declarations[1].(*ast.FunctionDecl)
	if !ext.IsExternal {
		t.Error("last function should be external")
	}
}
