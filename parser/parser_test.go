package parser

import (
	"kylix/ast"
	"kylix/lexer"
	"testing"
)

func newParser(input string) *Parser {
	l := lexer.New(input)
	return New(l)
}

func checkNoErrors(t *testing.T, p *Parser) {
	t.Helper()
	errs := p.Errors()
	if len(errs) == 0 {
		return
	}
	t.Errorf("parser has %d errors:", len(errs))
	for _, e := range errs {
		t.Errorf("  parser error: %q", e)
	}
	t.FailNow()
}

// ─── Literals ───────────────────────────────────────────────────────────────

func TestParseIntegerLiteral(t *testing.T) {
	p := newParser(`program Test; begin var x: Integer; x := 42; end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	if prog == nil {
		t.Fatal("ParseProgram returned nil")
	}
}

func TestParseFloatLiteral(t *testing.T) {
	p := newParser(`program Test; begin var x: Real; x := 3.14; end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	if prog == nil {
		t.Fatal("ParseProgram returned nil")
	}
}

func TestParseBooleanLiteral(t *testing.T) {
	for _, src := range []string{
		`program Test; begin var b: Boolean; b := true; end.`,
		`program Test; begin var b: Boolean; b := false; end.`,
	} {
		p := newParser(src)
		prog := p.ParseProgram()
		checkNoErrors(t, p)
		if prog == nil {
			t.Fatal("ParseProgram returned nil")
		}
	}
}

func TestParseStringLiteral(t *testing.T) {
	p := newParser(`program Test; begin var s: String; s := 'hello'; end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

// ─── Expressions ────────────────────────────────────────────────────────────

func TestParseInfixExpression(t *testing.T) {
	tests := []struct {
		src string
	}{
		{`program T; begin var x: Integer; x := 1 + 2; end.`},
		{`program T; begin var x: Integer; x := 3 - 1; end.`},
		{`program T; begin var x: Integer; x := 4 * 2; end.`},
		{`program T; begin var x: Integer; x := 10 div 2; end.`},
		{`program T; begin var b: Boolean; b := 1 = 1; end.`},
		{`program T; begin var b: Boolean; b := 1 <> 2; end.`},
		{`program T; begin var b: Boolean; b := 1 < 2; end.`},
		{`program T; begin var b: Boolean; b := 1 > 0; end.`},
	}
	for _, tt := range tests {
		p := newParser(tt.src)
		prog := p.ParseProgram()
		checkNoErrors(t, p)
		if prog == nil {
			t.Fatalf("nil program for: %s", tt.src)
		}
	}
}

func TestParsePrefixNot(t *testing.T) {
	p := newParser(`program T; begin var b: Boolean; b := not true; end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

func TestParseCallExpression(t *testing.T) {
	p := newParser(`program T; begin WriteLn('hi'); end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	if len(prog.Statements) == 0 {
		t.Fatal("expected at least one statement")
	}
}

func TestParseCallExpressionMultiArg(t *testing.T) {
	p := newParser(`program T; begin WriteLn('a', 'b', 'c'); end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatal("expected ExpressionStatement")
	}
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatal("expected CallExpression")
	}
	if len(call.Arguments) != 3 {
		t.Fatalf("expected 3 args, got %d", len(call.Arguments))
	}
}

func TestParseMemberAccess(t *testing.T) {
	p := newParser(`program T; begin var x: Integer; x := obj.Field; end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

func TestParseIndexExpression(t *testing.T) {
	p := newParser(`program T; begin var x: Integer; x := arr[0]; end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

// ─── Statements ─────────────────────────────────────────────────────────────

func TestParseIfStatement(t *testing.T) {
	p := newParser(`program T; begin if x > 0 then WriteLn('pos'); end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	if len(prog.Statements) == 0 {
		t.Fatal("expected statement")
	}
	_, ok := prog.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", prog.Statements[0])
	}
}

func TestParseIfElseStatement(t *testing.T) {
	p := newParser(`program T; begin if x > 0 then WriteLn('pos') else WriteLn('neg'); end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	stmt, ok := prog.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", prog.Statements[0])
	}
	if stmt.Alternative == nil {
		t.Fatal("expected alternative branch")
	}
}

func TestParseWhileStatement(t *testing.T) {
	p := newParser(`program T; begin var i: Integer; i := 0; while i < 10 do begin i := i + 1; end; end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	// Find the while statement
	found := false
	for _, s := range prog.Statements {
		if _, ok := s.(*ast.WhileStatement); ok {
			found = true
		}
	}
	if !found {
		t.Fatal("expected WhileStatement")
	}
}

func TestParseForStatement(t *testing.T) {
	p := newParser(`program T; begin var i: Integer; for i := 1 to 10 do WriteLn(i); end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, s := range prog.Statements {
		if _, ok := s.(*ast.ForStatement); ok {
			found = true
		}
	}
	if !found {
		t.Fatal("expected ForStatement")
	}
}

func TestParseAssignment(t *testing.T) {
	p := newParser(`program T; begin var x: Integer; x := 5; end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

// ─── Declarations ───────────────────────────────────────────────────────────

func TestParseFunctionDecl(t *testing.T) {
	src := `
program T;
function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;
begin
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, d := range prog.Declarations {
		if fn, ok := d.(*ast.FunctionDecl); ok {
			if fn.Name == "Add" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected FunctionDecl 'Add'")
	}
}

func TestParseProcedureDecl(t *testing.T) {
	src := `
program T;
procedure Greet(name: String);
begin
  WriteLn('Hello, ', name);
end;
begin
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, d := range prog.Declarations {
		if fn, ok := d.(*ast.FunctionDecl); ok {
			if fn.Name == "Greet" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected FunctionDecl 'Greet'")
	}
}

func TestParseVarDecl(t *testing.T) {
	src := `program T; var x: Integer; var s: String; begin end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	count := 0
	for _, d := range prog.Declarations {
		if _, ok := d.(*ast.VarDecl); ok {
			count++
		}
	}
	if count < 2 {
		t.Fatalf("expected at least 2 VarDecl, got %d", count)
	}
}

func TestParseConstDecl(t *testing.T) {
	src := `program T; const MaxSize = 100; begin end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, d := range prog.Declarations {
		if _, ok := d.(*ast.ConstDecl); ok {
			found = true
		}
	}
	if !found {
		t.Fatal("expected ConstDecl")
	}
}

// ─── Classes ────────────────────────────────────────────────────────────────

func TestParseClassDecl(t *testing.T) {
	src := `
program T;
class Animal
private
  var Name: String;
public
  constructor Create(name: String);
  begin
    Name := name;
  end;
  procedure Speak; virtual;
  begin
    WriteLn(Name);
  end;
end;
begin
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, d := range prog.Declarations {
		if cls, ok := d.(*ast.ClassDecl); ok {
			if cls.Name == "Animal" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected ClassDecl 'Animal'")
	}
}

func TestParseInheritedCall_Simple(t *testing.T) {
	src := `
program T;
class Dog inherits Animal
public
  constructor Create(name: String; age: Integer; breed: String);
  begin
    inherited Create(name, age);
  end;
end;
begin
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

// ─── Try/Except ─────────────────────────────────────────────────────────────

func TestParseTryExcept(t *testing.T) {
	src := `
program T;
begin
  try
    WriteLn('trying');
  except
    WriteLn('error');
  end;
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, s := range prog.Statements {
		if _, ok := s.(*ast.TryStatement); ok {
			found = true
		}
	}
	if !found {
		t.Fatal("expected TryStatement")
	}
}

// ─── Is / As ────────────────────────────────────────────────────────────────

func TestParseIsExpression(t *testing.T) {
	src := `program T; begin var b: Boolean; b := obj is TAnimal; end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

func TestParseAsExpression(t *testing.T) {
	src := `program T; begin var a: TAnimal; a := obj as TAnimal; end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

// ─── Map / Array ─────────────────────────────────────────────────────────────

func TestParseMapDecl(t *testing.T) {
	src := `program T; var m: map[String]Integer; begin end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

func TestParseArrayLiteral(t *testing.T) {
	src := `program T; begin var a: array of Integer; a := [1, 2, 3]; end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	_ = prog
}

// ─── String Interpolation ────────────────────────────────────────────────────

func TestParseStringInterpolation(t *testing.T) {
	src := `program T; begin WriteLn($"Hello, {name}!"); end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	if len(prog.Statements) == 0 {
		t.Fatal("expected at least one statement")
	}
	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", prog.Statements[0])
	}
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", stmt.Expression)
	}
	if len(call.Arguments) == 0 {
		t.Fatal("expected at least one argument")
	}
	_, ok = call.Arguments[0].(*ast.StringInterpolation)
	if !ok {
		t.Fatalf("expected StringInterpolation argument, got %T", call.Arguments[0])
	}
}

// ─── Match Statement ─────────────────────────────────────────────────────────

func TestParseMatchStatement(t *testing.T) {
	src := `
program T;
begin
  match x {
    1 => WriteLn('one');
    2 => WriteLn('two');
  };
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, s := range prog.Statements {
		if ms, ok := s.(*ast.MatchStatement); ok {
			found = true
			if len(ms.Branches) != 2 {
				t.Errorf("expected 2 branches, got %d", len(ms.Branches))
			}
		}
	}
	if !found {
		t.Fatal("expected MatchStatement")
	}
}

// ─── Try/Finally ─────────────────────────────────────────────────────────────

func TestParseTryCatch(t *testing.T) {
	src := `
program T;
begin
  try
    WriteLn('body');
  except
    on E: Exception do
      WriteLn(E.Message);
  end;
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, s := range prog.Statements {
		if ts, ok := s.(*ast.TryStatement); ok {
			found = true
			if len(ts.OnClauses) == 0 {
				t.Error("expected at least one OnClause")
			}
		}
	}
	if !found {
		t.Fatal("expected TryStatement with OnClauses")
	}
}

func TestParseTryFinally(t *testing.T) {
	src := `
program T;
begin
  try
    WriteLn('body');
  finally
    WriteLn('cleanup');
  end;
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, s := range prog.Statements {
		if ts, ok := s.(*ast.TryStatement); ok {
			found = true
			if ts.FinallyBlock == nil {
				t.Error("expected FinallyBlock")
			}
		}
	}
	if !found {
		t.Fatal("expected TryStatement with FinallyBlock")
	}
}

// ─── Additional coverage ─────────────────────────────────────────────────────

func TestParseProgram_Basic(t *testing.T) {
	src := `
program Hello;
var
  x: Integer;
begin
  x := 42;
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	if prog.Name != "Hello" {
		t.Fatalf("expected program name 'Hello', got %q", prog.Name)
	}
	if len(prog.Declarations) == 0 {
		t.Fatal("expected at least one declaration")
	}
	_, ok := prog.Declarations[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected VarDecl, got %T", prog.Declarations[0])
	}
}

func TestParseFunction(t *testing.T) {
	src := `
program T;
function Multiply(a: Integer; b: Integer): Integer;
begin
  result := a * b;
end;
begin
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, d := range prog.Declarations {
		if fn, ok := d.(*ast.FunctionDecl); ok && fn.Name == "Multiply" {
			found = true
			if fn.ReturnType == nil {
				t.Error("expected non-nil ReturnType")
			}
			if len(fn.Parameters) != 2 {
				t.Errorf("expected 2 parameters, got %d", len(fn.Parameters))
			}
		}
	}
	if !found {
		t.Fatal("expected FunctionDecl 'Multiply'")
	}
}

func TestParseProcedure(t *testing.T) {
	src := `
program T;
procedure Print(msg: String);
begin
  WriteLn(msg);
end;
begin
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	found := false
	for _, d := range prog.Declarations {
		if fn, ok := d.(*ast.FunctionDecl); ok && fn.Name == "Print" {
			found = true
			if fn.ReturnType != nil {
				t.Error("procedure should have nil ReturnType")
			}
			if len(fn.Parameters) != 1 {
				t.Errorf("expected 1 parameter, got %d", len(fn.Parameters))
			}
		}
	}
	if !found {
		t.Fatal("expected FunctionDecl 'Print' (procedure)")
	}
}

func TestParseClass(t *testing.T) {
	src := `
program T;
class Person
private
  var Name: String;
  var Age: Integer;
public
  constructor Create(name: String; age: Integer);
  begin
    Name := name;
    Age := age;
  end;
  function GetName: String;
  begin
    result := Name;
  end;
end;
begin
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	var cls *ast.ClassDecl
	for _, d := range prog.Declarations {
		if c, ok := d.(*ast.ClassDecl); ok && c.Name == "Person" {
			cls = c
			break
		}
	}
	if cls == nil {
		t.Fatal("expected ClassDecl 'Person'")
	}
	if len(cls.Fields) < 2 {
		t.Errorf("expected at least 2 fields, got %d", len(cls.Fields))
	}
	if len(cls.Methods) < 2 {
		t.Errorf("expected at least 2 methods, got %d", len(cls.Methods))
	}
}

func TestParseInheritedCall_WithClass(t *testing.T) {
	src := `
program T;
class Dog inherits Animal
public
  constructor Create(name: String; age: Integer);
  begin
    inherited Create(name, age);
  end;
end;
begin
end.`
	p := newParser(src)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	var dog *ast.ClassDecl
	for _, d := range prog.Declarations {
		if c, ok := d.(*ast.ClassDecl); ok && c.Name == "Dog" {
			dog = c
			break
		}
	}
	if dog == nil {
		t.Fatal("expected ClassDecl 'Dog'")
	}
	if dog.Parent != "Animal" {
		t.Errorf("expected Parent 'Animal', got %q", dog.Parent)
	}
	if len(dog.Methods) == 0 {
		t.Fatal("expected at least one method")
	}
	body := dog.Methods[0].Body
	if body == nil || len(body.Statements) == 0 {
		t.Fatal("expected method body with statements")
	}
	_, ok := body.Statements[0].(*ast.InheritedStatement)
	if !ok {
		t.Fatalf("expected InheritedStatement, got %T", body.Statements[0])
	}
}

// ─── Program structure ───────────────────────────────────────────────────────

func TestParseProgramName(t *testing.T) {
	p := newParser(`program Hello; begin end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	if prog.Name != "Hello" {
		t.Fatalf("expected program name 'Hello', got %q", prog.Name)
	}
}

func TestParseUnitName(t *testing.T) {
	p := newParser(`unit myunit; begin end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	if prog.UnitName != "myunit" {
		t.Fatalf("expected unit name 'myunit', got %q", prog.UnitName)
	}
}

func TestParseEmptyProgram(t *testing.T) {
	p := newParser(`program Empty; begin end.`)
	prog := p.ParseProgram()
	checkNoErrors(t, p)
	if prog == nil {
		t.Fatal("expected non-nil program")
	}
}
