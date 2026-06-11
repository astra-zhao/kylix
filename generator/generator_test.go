package generator

import (
	"strings"
	"testing"

	"kylix/lexer"
	"kylix/parser"
)

func compile(input string) (string, []string) {
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return "", p.Errors()
	}
	g := New()
	return g.Generate(prog), nil
}

func assertNoErrors(t *testing.T, errs []string) {
	t.Helper()
	if len(errs) > 0 {
		t.Fatalf("parser errors: %v", errs)
	}
}

func assertContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("expected output to contain %q\ngot:\n%s", substr, output)
	}
}

func assertNotContains(t *testing.T, output, substr string) {
	t.Helper()
	if strings.Contains(output, substr) {
		t.Errorf("expected output NOT to contain %q\ngot:\n%s", substr, output)
	}
}

// ---------------------------------------------------------------------------
// Hello world
// ---------------------------------------------------------------------------

func TestGenerate_HelloWorld(t *testing.T) {
	input := `
program Hello;
begin
  WriteLn('Hello, World!');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "package main")
	assertContains(t, out, "func main()")
	assertContains(t, out, `fmt.Println`)
	assertContains(t, out, "Hello, World!")
}

// ---------------------------------------------------------------------------
// Variable declarations
// ---------------------------------------------------------------------------

func TestGenerate_VarDecl(t *testing.T) {
	input := `
program VarTest;
var
  x: Integer;
  s: String;
begin
  x := 42;
  s := 'hello';
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "package main")
	assertContains(t, out, "func main()")
}

// ---------------------------------------------------------------------------
// Function declaration
// ---------------------------------------------------------------------------

func TestGenerate_FunctionDecl(t *testing.T) {
	input := `
program FuncTest;

function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

begin
  var r := Add(3, 4);
  WriteLn(r);
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "func Add(")
	assertContains(t, out, "return result")
}

// ---------------------------------------------------------------------------
// If statement
// ---------------------------------------------------------------------------

func TestGenerate_IfStatement(t *testing.T) {
	input := `
program IfTest;
begin
  var x := 10;
  if x > 5 then
    WriteLn('big')
  else
    WriteLn('small');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "if ")
	assertContains(t, out, "else {")
}

// ---------------------------------------------------------------------------
// While loop
// ---------------------------------------------------------------------------

func TestGenerate_WhileLoop(t *testing.T) {
	input := `
program WhileTest;
begin
  var i := 0;
  while i < 10 do
  begin
    i := i + 1;
  end;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "for ")
}

// ---------------------------------------------------------------------------
// For loop
// ---------------------------------------------------------------------------

func TestGenerate_ForLoop(t *testing.T) {
	input := `
program ForTest;
begin
  var i: Integer;
  for i := 1 to 5 do
    WriteLn(i);
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "for ")
}

// ---------------------------------------------------------------------------
// Class declaration
// ---------------------------------------------------------------------------

func TestGenerate_ClassDecl(t *testing.T) {
	input := `
program ClassTest;

class Animal
private
  var Name: String;
public
  constructor Create(name: String);
  begin
    Name := name;
  end;

  procedure Speak;
  begin
    WriteLn(Name);
  end;
end;

begin
  var a := Animal.Create('Cat');
  a.Speak();
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "type Animal struct")
	assertContains(t, out, "&Animal{")
}

// ---------------------------------------------------------------------------
// Map type
// ---------------------------------------------------------------------------

func TestGenerate_MapType(t *testing.T) {
	input := `
program MapTest;
var
  m: map[String]Integer;
begin
  m['key'] := 42;
  WriteLn(m['key']);
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "map[string]int64")
}

// ---------------------------------------------------------------------------
// Try/except
// ---------------------------------------------------------------------------

func TestGenerate_TryExcept(t *testing.T) {
	input := `
program TryTest;
begin
  try
    WriteLn('try');
  except
    WriteLn('caught');
  end;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "func()")
}

// ---------------------------------------------------------------------------
// Boolean literals
// ---------------------------------------------------------------------------

func TestGenerate_BooleanLiterals(t *testing.T) {
	input := `
program BoolTest;
var b: Boolean;
begin
  b := true;
  b := false;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "true")
	assertContains(t, out, "false")
}

// ---------------------------------------------------------------------------
// Infix arithmetic
// ---------------------------------------------------------------------------

func TestGenerate_ArithmeticExpressions(t *testing.T) {
	input := `
program ArithTest;
begin
  var x := 2 + 3 * 4 - 1;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "+")
	assertContains(t, out, "*")
}

// ---------------------------------------------------------------------------
// Nil literal
// ---------------------------------------------------------------------------

func TestGenerate_NilLiteral(t *testing.T) {
	input := `
program NilTest;
var p: String;
begin
  p := nil;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "nil")
}

// ---------------------------------------------------------------------------
// Package declaration
// ---------------------------------------------------------------------------

func TestGenerate_PackageMain(t *testing.T) {
	input := `
program Pkg;
begin
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	if !strings.HasPrefix(strings.TrimSpace(out), "package main") {
		t.Errorf("expected output to start with 'package main', got:\n%s", out[:min(len(out), 100)])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// Required test suite
// ---------------------------------------------------------------------------

// TestGenerateHello — WriteLn('Hello') → fmt.Println("Hello")
func TestGenerateHello(t *testing.T) {
	input := `
program Hello;
begin
  WriteLn('Hello');
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `fmt.Println`)
	assertContains(t, out, `Hello`)
}

// TestGenerateVarDecl — var x: Integer → var x int64
func TestGenerateVarDecl(t *testing.T) {
	input := `
program VarDeclTest;
var
  x: Integer;
begin
  x := 1;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `var x int64`)
}

// TestGenerateIfStatement — if/then/else generates correct Go if/else
func TestGenerateIfStatement(t *testing.T) {
	input := `
program IfStmtTest;
begin
  var x := 10;
  if x > 5 then
  begin
    WriteLn('big');
  end
  else
  begin
    WriteLn('small');
  end;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "if (")
	assertContains(t, out, "else {")
	assertContains(t, out, `"big"`)
	assertContains(t, out, `"small"`)
}

// TestGenerateWhileLoop — while generates Go for loop
func TestGenerateWhileLoop(t *testing.T) {
	input := `
program WhileLoopTest;
var
  i: Integer;
begin
  i := 0;
  while i < 10 do
  begin
    i := i + 1;
  end;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, "for (")
}

// TestGenerateForLoop — for i := 1 to 10 generates Go for loop
func TestGenerateForLoop(t *testing.T) {
	input := `
program ForLoopTest;
var
  i: Integer;
begin
  for i := 1 to 10 do
  begin
    WriteLn(i);
  end;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `for i = 1; i <= 10; i++`)
}

// TestGenerateClass — class with fields and methods
func TestGenerateClass(t *testing.T) {
	input := `
program ClassDeclTest;

class Animal
private
  var Name: String;
  var Age: Integer;
public
  constructor Create(name: String; age: Integer);
  begin
    Name := name;
    Age := age;
  end;

  procedure Speak;
  begin
    WriteLn(Name);
  end;
end;

begin
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `type Animal struct`)
	assertContains(t, out, `Name string`)
	assertContains(t, out, `Age int64`)
	assertContains(t, out, `func (self *Animal) Speak(`)
}

// TestGenerateInheritedCall — inherited Create(name, age) calls the parent method via self
func TestGenerateInheritedCall(t *testing.T) {
	input := `
program InheritedCallTest;

class Animal
private
  var Name: String;
  var Age: Integer;
public
  constructor Create(name: String; age: Integer);
  begin
    Name := name;
    Age := age;
  end;
end;

class Dog inherits Animal
public
  constructor Create(name: String; age: Integer);
  begin
    inherited Create(name, age);
  end;
end;

begin
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	// The generator emits "self.Create(name, age)" — Go struct embedding handles dispatch
	assertContains(t, out, `self.Create`)
	assertContains(t, out, `name`)
	assertContains(t, out, `age`)
}

// TestGenerateTryCatch — try/except generates Go panic/recover
func TestGenerateTryCatch(t *testing.T) {
	input := `
program TryCatchTest;
begin
  try
  begin
    WriteLn('inside try');
  end
  except
  begin
    WriteLn('caught');
  end
  end;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `recover()`)
	assertContains(t, out, `defer func()`)
}

// TestGenerateFunctionReturn — function with result := ... returns correctly
func TestGenerateFunctionReturn(t *testing.T) {
	input := `
program FuncReturnTest;

function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

begin
  WriteLn(Add(2, 3));
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `func Add(`)
	assertContains(t, out, `int64`)
	assertContains(t, out, `return result`)
}

// TestGenerateMapType — map[String]Integer → Go map[string]int64
func TestGenerateMapType(t *testing.T) {
	input := `
program MapTypeTest;
var
  scores: map[String]Integer;
begin
  scores['Alice'] := 95;
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `map[string]int64`)
}

// TestGenerateStringInterpolation — interpolated strings generate fmt.Sprintf
// Kylix uses $"..." syntax (dollar before double-quoted string) for interpolation.
func TestGenerateStringInterpolation(t *testing.T) {
	input := `
program StringInterpTest;
begin
  var name := 'World';
  var greeting := $"Hello, ${name}!";
  WriteLn(greeting);
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `fmt.Sprintf`)
}

// TestGenerateMatchStatement — match generates Go switch
func TestGenerateMatchStatement(t *testing.T) {
	input := `
program MatchTest;

function Describe(value: Integer): String;
begin
  match value {
    0 => 'zero',
    1 => 'one',
    _ => 'other'
  };
end;

begin
  WriteLn(Describe(1));
end.`
	out, errs := compile(input)
	assertNoErrors(t, errs)
	assertContains(t, out, `switch _v :=`)
	assertContains(t, out, `default:`)
}
