package lsp_test

import (
	"path/filepath"
	"testing"

	"kylix/pkg/lsp"
)

// diagnostics_test.go — v6.8.0 undefined-identifier warnings (severity 2).

// warnings returns the warning diagnostics (severity 2) for src.
func warnings(t *testing.T, src string) []lsp.Diagnostic {
	t.Helper()
	klxTestDir(t) // ensure stdlib/klx exists; the check runs only then
	// Point KYLIX_HOME so used stdlib symbols load and are not flagged.
	t.Setenv("KYLIX_HOME", filepath.Dir(filepath.Dir(klxTestDir(t))))
	doc := lsp.NewDocument("file:///test.klx", src)
	var out []lsp.Diagnostic
	for _, d := range doc.Diagnostics {
		if d.Severity == 2 {
			out = append(out, d)
		}
	}
	return out
}

func TestUndefinedIdentifierWarning(t *testing.T) {
	ws := warnings(t, `program p;
var
  x: Integer;
begin
  x := 10;
  WriteLn(x + typoVar);
end.`)
	if len(ws) != 1 {
		t.Fatalf("expected 1 undefined-identifier warning, got %d: %+v", len(ws), ws)
	}
	if ws[0].Message != "Undefined identifier 'typoVar'" {
		t.Errorf("unexpected message: %q", ws[0].Message)
	}
}

func TestNoWarningForDeclared(t *testing.T) {
	ws := warnings(t, `program p;
var
  x, y: Integer;
  s: String;
begin
  x := 5;
  y := x + 1;
  s := 'hi';
  WriteLn(x + y);
  WriteLn(s);
end.`)
	if len(ws) != 0 {
		t.Errorf("declared identifiers must not warn: %+v", ws)
	}
}

func TestNoWarningForClassMembers(t *testing.T) {
	// TypeDecl-wrapped class: fields, method params, and locals must be known.
	ws := warnings(t, `program p;
type
  TCounter = class
    Value: Integer;
    procedure Add(n: Integer);
  end;
procedure TCounter.Add(n: Integer);
begin
  self.Value := self.Value + n;
end;
begin
end.`)
	if len(ws) != 0 {
		t.Errorf("class members/params must not warn: %+v", ws)
	}
}

func TestNoWarningForStdlibFunctions(t *testing.T) {
	ws := warnings(t, `program p;
uses jsonutil;
var
  m: map[String]Variant;
begin
  m := JsonDecodeMap('{"a":1}');
  WriteLn(JsonGetInt(m, 'a'));
end.`)
	if len(ws) != 0 {
		t.Errorf("stdlib functions must not warn: %+v", ws)
	}
}

func TestNoWarningForLoopAndLambda(t *testing.T) {
	ws := warnings(t, `program p;
var
  total: Integer;
begin
  for i := 1 to 10 do
    total := total + i;
  var f := lambda(x: Integer): Integer
    begin
      result := x * 2;
    end;
  WriteLn(f(total));
end.`)
	if len(ws) != 0 {
		t.Errorf("loop vars / lambda params must not warn: %+v", ws)
	}
}
