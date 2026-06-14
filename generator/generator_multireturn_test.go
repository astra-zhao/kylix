package generator_test

import (
	"kylix/generator"
	"kylix/lexer"
	"kylix/parser"
	"strings"
	"testing"
)

// parseGen parses Kylix source and generates Go code.
func parseGen(t *testing.T, src string) string {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	g := generator.New()
	return g.Generate(prog)
}

func TestGenerateMultiReturnFunction(t *testing.T) {
	src := `program Test;
function Swap(a: Integer; b: Integer): (Integer, Integer);
begin
  result := (b, a);
end;
begin
end.`
	out := parseGen(t, src)

	// Should generate: func Swap(a int64, b int64) (int64, int64) { return b, a }
	if !strings.Contains(out, "func Swap") {
		t.Error("expected func Swap")
	}
	if !strings.Contains(out, "(int64, int64)") {
		t.Error("expected multi-return signature (int64, int64)")
	}
	if !strings.Contains(out, "return b, a") {
		t.Error("expected 'return b, a' from result := (b, a)")
	}
}

func TestGenerateMultiReturnCall(t *testing.T) {
	src := `program Test;
function Pair(): (Integer, Integer);
begin result := (10, 20); end;

begin
  var x, y := Pair();
  WriteLn(x);
end.`
	out := parseGen(t, src)

	// Should generate: x, y := Pair()
	if !strings.Contains(out, "x, y := Pair()") {
		t.Errorf("expected 'x, y := Pair()', got:\n%s", out)
	}
}

func TestGenerateTupleReturnStatement(t *testing.T) {
	src := `program Test;
function Foo(): (Integer, String);
begin
  return (42, 'hello');
end;
begin
end.`
	out := parseGen(t, src)

	// return (42, 'hello') → return 42, "hello"
	if !strings.Contains(out, "return 42") && !strings.Contains(out, `"hello"`) {
		t.Errorf("expected 'return 42, \"hello\"', got:\n%s", out)
	}
}

func TestGenerateMultiReturnNestedTuple(t *testing.T) {
	src := `program Test;
function Triple(): (Integer, Integer, Integer);
begin
  result := (1, 2, 3);
end;
begin
end.`
	out := parseGen(t, src)

	// result := (1, 2, 3) → return 1, 2, 3
	if !strings.Contains(out, "return 1, 2, 3") {
		t.Errorf("expected 'return 1, 2, 3', got:\n%s", out)
	}
}
