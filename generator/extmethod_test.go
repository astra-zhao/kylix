package generator_test

import (
	"strings"
	"testing"
)

// Class method external definition tests (v2.5.0 task 5).

func TestExternalMethodDef_NoDuplicate(t *testing.T) {
	src := `program Test;
type
  TFoo = class
    function Bar(): Integer;
  end;

function TFoo.Bar(): Integer;
begin
  result := 42;
end;

begin end.`

	out := parseGen(t, src)
	// Count occurrences of "func (self *TFoo) Bar" — should be exactly 1.
	count := strings.Count(out, "func (self *TFoo) Bar")
	if count != 1 {
		t.Errorf("expected exactly 1 Bar method, got %d:\n%s", count, out)
	}
}

func TestExternalMethodDef_MultipleMethods(t *testing.T) {
	src := `program Test;
type
  TCalc = class
    function Add(a: Integer; b: Integer): Integer;
    function Sub(a: Integer; b: Integer): Integer;
  end;

function TCalc.Add(a: Integer; b: Integer): Integer;
begin result := a + b; end;

function TCalc.Sub(a: Integer; b: Integer): Integer;
begin result := a - b; end;

begin end.`

	out := parseGen(t, src)
	if strings.Count(out, "func (self *TCalc) Add") != 1 {
		t.Errorf("expected 1 Add method, got %d", strings.Count(out, "func (self *TCalc) Add"))
	}
	if strings.Count(out, "func (self *TCalc) Sub") != 1 {
		t.Errorf("expected 1 Sub method, got %d", strings.Count(out, "func (self *TCalc) Sub"))
	}
}

func TestInternalMethodDef_StillWorks(t *testing.T) {
	// Methods defined inside class body should still work (no regression).
	src := `program Test;
type
  TFoo = class
    function Bar(): Integer;
    begin
      result := 99;
    end;
  end;

begin end.`

	out := parseGen(t, src)
	if strings.Count(out, "func (self *TFoo) Bar") != 1 {
		t.Errorf("expected 1 Bar method for inline definition, got %d", strings.Count(out, "func (self *TFoo) Bar"))
	}
}

func TestExternalMethodDef_Procedure(t *testing.T) {
	src := `program Test;
type
  TLogger = class
    procedure Log(msg: String);
  end;

procedure TLogger.Log(msg: String);
begin
  WriteLn(msg);
end;

begin end.`

	out := parseGen(t, src)
	if strings.Count(out, "func (self *TLogger) Log") != 1 {
		t.Errorf("expected 1 Log method, got %d", strings.Count(out, "func (self *TLogger) Log"))
	}
}
