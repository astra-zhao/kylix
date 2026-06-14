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
