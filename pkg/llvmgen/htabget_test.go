package llvmgen_test

import (
	"strings"
	"testing"
)

// TestHtabGet_NullOnMiss (v5.6.0) guards the fix for `map[String]Boolean`
// presence tests reading as true for missing keys. htab_get previously
// returned the empty-string ptr (non-null) on miss, so `if m[key]` (lowered to
// `htab_get → icmp ne null`) read every missing key as TRUE. In the bootstrap
// that made `self.ClassIsBase[type]` / `self.ClassTypes[type]` return true for
// every user type → MapType emitted `interface{}` for all record/class-typed
// vars → field access "undefined (type interface{})". Now htab_get returns null
// on miss, and Boolean map reads convert via `icmp ne ptr, null` (present→true,
// miss→false); string/integer reads null-guard (null → "" / 0).
func TestHtabGet_NullOnMiss(t *testing.T) {
	ir := generateIR(t, `program p;
var m: map[String]Boolean;
begin
  if m['k'] then WriteLn('yes') else WriteLn('no');
end.`)
	// htab_get's miss path must `ret ptr null` (not the empty-string ptr).
	if !strings.Contains(ir, "ret ptr null") {
		t.Errorf("expected htab_get to return null on miss (ret ptr null)\nIR:\n%s", ir)
	}
	// The Boolean map read must test presence via `icmp ne ptr <htab_get>, null`
	// (not pass the raw ptr as a bool / store ptr to an i1).
	if !strings.Contains(ir, "icmp ne ptr") {
		t.Errorf("expected a Boolean map read to emit `icmp ne ptr <result>, null` for presence\nIR:\n%s", ir)
	}
}
