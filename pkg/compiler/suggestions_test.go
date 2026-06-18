package compiler_test

import (
	"strings"
	"testing"

	"kylix/pkg/compiler"
)

func TestHint_TypeConversion_StringToInteger(t *testing.T) {
	result := compileTC(t, `program Test;
var x: Integer;
begin
  x := 'hello';
end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	d := result.Diagnostics[0]
	if !strings.Contains(d.Hint, "StrToInt") {
		t.Errorf("expected StrToInt hint, got: %q", d.Hint)
	}
}

func TestHint_TypeConversion_IntegerToString(t *testing.T) {
	result := compileTC(t, `program Test;
var s: String;
begin
  s := 42;
end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	d := result.Diagnostics[0]
	if !strings.Contains(d.Hint, "IntToStr") {
		t.Errorf("expected IntToStr hint, got: %q", d.Hint)
	}
}

func TestHint_SpellCorrection_Undeclared(t *testing.T) {
	result := compileTC(t, `program Test;
var userName: String;
begin
  userNaem := 'hello';
end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	d := result.Diagnostics[0]
	if !strings.Contains(d.Hint, "userName") {
		t.Errorf("expected hint suggesting 'userName', got: %q", d.Hint)
	}
}

func TestHint_SpellCorrection_NoClose(t *testing.T) {
	// 'xyz' has no close match in scope — no hint expected
	result := compileTC(t, `program Test;
var longVariableNameHere: Integer;
begin
  xyz := 42;
end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	d := result.Diagnostics[0]
	// hint may be empty or unrelated — just verify no false positive "did you mean"
	if strings.Contains(d.Hint, "longVariable") {
		t.Errorf("unexpected hint for unrelated name: %q", d.Hint)
	}
}

func TestSuggestions_Levenshtein(t *testing.T) {
	cases := []struct {
		target    string
		want      string
		wantFound bool
	}{
		{"userName", []string{"userName"}[0], true}, // exact
		{"userNaem", "userName", true},              // 2 transpositions
		{"usr", "userName", false},                  // too different
		{"WritLn", "WriteLn", true},                 // 1 edit
	}
	for _, tc := range cases {
		got := compiler.NearestName(tc.target, []string{"userName", "WriteLn", "ReadLn"}, 2)
		if tc.wantFound && got != tc.want {
			t.Errorf("NearestName(%q) = %q, want %q", tc.target, got, tc.want)
		}
		if !tc.wantFound && got != "" {
			t.Errorf("NearestName(%q) = %q, expected empty (no close match)", tc.target, got)
		}
	}
}

func TestHint_MissingMethodHint(t *testing.T) {
	result := compileTC(t, `program Test;
type
  IFoo = interface
    procedure Bar();
  end;
  TFoo = class implements IFoo
  end;
begin end.`)
	if result.Success {
		t.Fatal("expected failure")
	}
	d := result.Diagnostics[0]
	if !strings.Contains(d.Hint, "Bar") {
		t.Errorf("expected hint mentioning 'Bar', got: %q", d.Hint)
	}
}
