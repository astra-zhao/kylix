package compiler_test

import (
	"path/filepath"
	"testing"

	"kylix/pkg/compiler"
)

func TestTypeAlias_BasicUsage(t *testing.T) {
	// UserId = Integer; assigning integer to UserId should succeed
	result := compileTC(t, `program Test;
type
  UserId = Integer;
var id: UserId;
begin
  id := 42;
  WriteLn(id);
end.`)
	for _, d := range result.Diagnostics {
		if d.Level == "error" {
			t.Errorf("unexpected error: [%s] %s", d.Code, d.Message)
		}
	}
}

func TestTypeAlias_StringAlias(t *testing.T) {
	result := compileTC(t, `program Test;
type
  UserName = String;
var name: UserName;
begin
  name := 'Alice';
  WriteLn(name);
end.`)
	for _, d := range result.Diagnostics {
		if d.Level == "error" {
			t.Errorf("unexpected error: [%s] %s", d.Code, d.Message)
		}
	}
}

func TestTypeAlias_TypeMismatchThroughAlias(t *testing.T) {
	// UserId = Integer; assigning string should fail with KLX101
	result := compileTC(t, `program Test;
type
  UserId = Integer;
var id: UserId;
begin
  id := 'wrong';
end.`)
	if result.Success {
		t.Fatal("expected failure: String assigned to UserId (Integer alias)")
	}
	if !hasDiag(result, "cannot assign String") {
		t.Errorf("expected type mismatch diagnostic, got: %v", result.Diagnostics)
	}
}

func TestTypeAlias_ChainedAlias(t *testing.T) {
	// SmallId → MyInt → Integer; should work transparently
	result := compileTC(t, `program Test;
type
  MyInt = Integer;
  SmallId = MyInt;
var x: SmallId;
begin
  x := 99;
end.`)
	for _, d := range result.Diagnostics {
		if d.Level == "error" {
			t.Errorf("unexpected error: [%s] %s", d.Code, d.Message)
		}
	}
}

func TestTypeAlias_RecursiveCycle(t *testing.T) {
	result := compileTC(t, `program Test;
type
  TypeA = TypeB;
  TypeB = TypeA;
begin end.`)
	if result.Success {
		t.Fatal("expected failure: recursive type alias")
	}
	if !hasDiagCode(result, "KLX105") {
		t.Errorf("expected KLX105 (alias cycle), got: %v", diagCodes(result))
	}
}

func TestTypeAlias_SelfReference(t *testing.T) {
	result := compileTC(t, `program Test;
type
  Loop = Loop;
begin end.`)
	if result.Success {
		t.Fatal("expected failure: self-referencing type alias")
	}
	if !hasDiagCode(result, "KLX105") {
		t.Errorf("expected KLX105, got: %v", diagCodes(result))
	}
}

func TestTypeAlias_MapAlias(t *testing.T) {
	// type UserMap = map[String]Integer; — alias of composite type
	result := compileTC(t, `program Test;
type
  UserMap = map[String]Integer;
var m: UserMap;
begin
  WriteLn('ok');
end.`)
	for _, d := range result.Diagnostics {
		if d.Level == "error" {
			t.Errorf("unexpected error: [%s] %s", d.Code, d.Message)
		}
	}
}

func TestTypeAlias_EndToEnd(t *testing.T) {
	// Full compile + run through CompileFile
	src := `program Test;
type
  Score = Integer;
  PlayerName = String;
var s: Score;
var n: PlayerName;
begin
  s := 100;
  n := 'Alice';
  WriteLn(s);
  WriteLn(n);
end.`
	f := writeTC(t, src)
	result, err := compile(t, f)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range result.Diagnostics {
		if d.Level == "error" {
			t.Errorf("unexpected error: [%s] %s", d.Code, d.Message)
		}
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

// compile is a helper that uses CompileFile with a temp output path.
func compile(t *testing.T, file string) (*compiler.Result, error) {
	t.Helper()
	return compiler.CompileFile(file, compiler.Options{
		OutputFile: filepath.Join(t.TempDir(), "out.go"),
	})
}
