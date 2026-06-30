package llvmgen_test

import (
	"strings"
	"testing"
)

func TestBreak_InsideFor(t *testing.T) {
	ir := generateExcIR(t, `program p;
var i: Integer;
begin
  for i := 1 to 10 do
  begin
    if i = 5 then
      break;
  end;
end.`)
	// break → br to exit label; no "outside loop" error
	if strings.Count(ir, "call void @longjmp") > 0 {
		// Should not raise, just branch
		t.Error("break should not emit longjmp")
	}
	// The IR should contain a conditional branch (from if) and an unconditional
	// branch from break.
	if !strings.Contains(ir, "br label") {
		t.Errorf("expected br label from break\nIR:\n%s", ir)
	}
}

func TestContinue_InsideWhile(t *testing.T) {
	ir := generateExcIR(t, `program p;
var i: Integer;
begin
  i := 0;
  while i < 10 do
  begin
    i := i + 1;
    if i = 5 then
      continue;
    WriteLn(i);
  end;
end.`)
	// continue → br to header label
	if !strings.Contains(ir, "br label") {
		t.Errorf("expected br label from continue\nIR:\n%s", ir)
	}
}

func TestCase_IntegerSwitch(t *testing.T) {
	ir := generateExcIR(t, `program p;
var n: Integer;
begin
  n := 2;
  case n of
    1: WriteLn('one');
    2: WriteLn('two');
    3: WriteLn('three');
  end;
end.`)
	// case lowered to LLVM switch instruction
	if !strings.Contains(ir, "switch i64") {
		t.Errorf("expected switch i64\nIR:\n%s", ir)
	}
	if !strings.Contains(ir, "i64 1,") || !strings.Contains(ir, "i64 2,") {
		t.Errorf("expected case value literals\nIR:\n%s", ir)
	}
}

func TestMatch_PatternBranches(t *testing.T) {
	ir := generateExcIR(t, `program p;
var n: Integer;
begin
  n := 3;
  match n {
    1: WriteLn('one');
    2: WriteLn('two');
    _: WriteLn('other');
  }
end.`)
	// match lowered to icmp eq chain
	if !strings.Contains(ir, "icmp eq i64") {
		t.Errorf("expected icmp eq i64\nIR:\n%s", ir)
	}
}

func TestForEach_OverString(t *testing.T) {
	ir := generateExcIR(t, `program p;
var c: Integer;
begin
  for c in 'hello' do
    WriteLn(c);
end.`)
	// forEach uses strlen as bound and getelementptr for element access
	if !strings.Contains(ir, "call i64 @strlen") {
		t.Errorf("expected strlen for foreach bound\nIR:\n%s", ir)
	}
	if !strings.Contains(ir, "getelementptr inbounds i8") {
		t.Errorf("expected getelementptr for element\nIR:\n%s", ir)
	}
}
