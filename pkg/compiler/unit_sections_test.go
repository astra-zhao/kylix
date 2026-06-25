package compiler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kylix/pkg/compiler"
)

func TestCompileProject_UnitInterfaceImplementation(t *testing.T) {
	dir := t.TempDir()
	unit := writeTempKlx(t, dir, "math_helper.klx", `unit MathHelper;

interface

function Square(x: Integer): Integer;
function Cube(x: Integer): Integer;
function IsEven(n: Integer): Boolean;

implementation

function Square(x: Integer): Integer;
begin
  result := x * x;
end;

function Cube(x: Integer): Integer;
begin
  result := x * x * x;
end;

function IsEven(n: Integer): Boolean;
begin
  result := (n mod 2) = 0;
end;

end.
`)
	main := writeTempKlx(t, dir, "main.klx", `program UseModule;
uses MathHelper;

begin
  WriteLn('Square of 5: ', Square(5));
  WriteLn('Cube of 3: ', Cube(3));
  WriteLn('Is 4 even? ', IsEven(4));
end.
`)
	out := filepath.Join(dir, "out.go")
	result, err := compiler.CompileProject([]string{unit, main}, compiler.Options{OutputFile: out})
	if err != nil {
		t.Fatalf("CompileProject error: %v", err)
	}
	if !result.Success {
		for _, d := range result.Diagnostics {
			t.Logf("diag: %s", d.Message)
		}
		t.Fatal("expected success")
	}

	goSrcBytes, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	goSrc := string(goSrcBytes)
	for _, want := range []string{"func Square(x int64) int64", "func Cube(x int64) int64", "func IsEven(n int64) bool"} {
		if !strings.Contains(goSrc, want) {
			t.Fatalf("expected generated Go to contain %q\n%s", want, goSrc)
		}
	}
	if strings.Contains(goSrc, "type  interface") {
		t.Fatalf("generated invalid empty interface declaration:\n%s", goSrc)
	}
	if strings.Count(goSrc, "func Square(x int64) int64") != 1 {
		t.Fatalf("expected one Square implementation, got %d\n%s", strings.Count(goSrc, "func Square(x int64) int64"), goSrc)
	}
}
