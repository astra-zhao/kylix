package llvmgen_test

import (
	"strings"
	"testing"
)

// TestMapFieldArrayRead_ZeroSlice (v0.5.6) guards the fix for the example27
// segfault. emitMapFieldIndexGet's array-valued miss path returned
// `insertvalue {ptr,len,cap} undef, ptr null, 0` — only data was null, len/cap
// stayed undef (garbage). `Length(fields)` then read a garbage len, `i < len`
// wrongly entered the element branch, GEP-ing the null data ptr → segfault
// (example27_try_except crashed in GenerateCallExpression's
// `fields := self.ClassFields[typeName]` because ClassFields is never
// populated). Now the zero slice is built on zeroinitializer (len=0, cap=0).
func TestMapFieldArrayRead_ZeroSlice(t *testing.T) {
	ir := generateIR(t, `program p;
type
  TBag = class
    Items: map[String]array of Integer;
  end;
var b: TBag;
begin
  b := TBag.Create;
end.`)
	// The array-valued map field read's miss path must build the zero slice on
	// zeroinitializer (not undef), so len/cap are 0 — not garbage that makes
	// `Length(fields)` return a junk value.
	if !strings.Contains(ir, "zeroinitializer") {
		t.Errorf("expected the array-valued map field read to use zeroinitializer for the zero slice\nIR:\n%s", ir)
	}
	// No `insertvalue { ptr, i64, i64 } undef` (the old buggy form).
	if strings.Contains(ir, "insertvalue { ptr, i64, i64 } undef") {
		t.Errorf("array map field read must not build the zero slice on undef (len/cap garbage → segfault)\nIR:\n%s", ir)
	}
}
