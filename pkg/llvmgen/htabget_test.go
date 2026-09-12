package llvmgen_test

import (
	"strings"
	"testing"
)

// TestHtabGet_NullOnMiss (v0.5.6) guards the fix for `map[String]Boolean`
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

// TestHtab_MagicCheck (v0.8.0 P2) guards the table-header magic guard: every
// htab entry point must call __kylix_htab_check, the guard must compare the
// first header word against the magic constant, and htab_new must store it.
// This is the loud-failure guard for the v0.7.0 P1 "value slot passed where a
// table pointer was expected" bug class (a forgotten load used to scribble
// over the heap until something unrelated crashed).
func TestHtab_MagicCheck(t *testing.T) {
	ir := generateIR(t, `program p;
var m: map[String]Integer;
begin
  m['a'] := 1;
  WriteLn(m['a']);
end.`)
	if !strings.Contains(ir, "define void @__kylix_htab_check(ptr %t)") {
		t.Errorf("expected the htab_check guard to be emitted\nIR:\n%s", ir)
	}
	// entry guard calls in every entry point (find/put/get/has/del/size/clear/keys)
	n := strings.Count(ir, "call void @__kylix_htab_check(ptr")
	if n < 7 {
		t.Errorf("expected >=7 htab_check entry calls, got %d\nIR:\n%s", n, ir)
	}
	if !strings.Contains(ir, "icmp eq i64") || !strings.Contains(ir, "4774189848202271051") {
		t.Errorf("expected the magic comparison (0x4241545F485F594B = 4774189848202271051)\nIR:\n%s", ir)
	}
	if !strings.Contains(ir, "call ptr @malloc(i64 24)") {
		t.Errorf("expected htab_new to allocate the 24-byte header (magic+buckets+size)\nIR:\n%s", ir)
	}
	if !strings.Contains(ir, "unreachable") {
		t.Errorf("expected `unreachable` after the noreturn exit call (LLVM demands a terminator)\nIR:\n%s", ir)
	}
}
