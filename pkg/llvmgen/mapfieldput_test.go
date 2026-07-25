package llvmgen_test

import (
	"strings"
	"testing"
)

// TestMapFieldIndexPut_Assign (v5.6.0) guards the fix for `obj.MapField[k] := v`
// (assign to a map-typed class field index). emitAssign previously fell through
// to emitArrayIndex→emitMapFieldIndexGet (a READ) and stored the RHS to the
// read result — for a Boolean map that was `store i1 …, ptr <icmp-result>`,
// an llc type error (`%t defined with type i1 but expected ptr`). It must now
// route to emitMapFieldIndexPut → `call void @__kylix_htab_put`.
func TestMapFieldIndexPut_Assign(t *testing.T) {
	ir := generateIR(t, `program p;
type
  TBag = class
    Seen: map[String]Boolean;
  end;
var b: TBag;
begin
  b := TBag.Create;
  b.Seen['x'] := true;
end.`)
	// Must emit a map PUT (htab_put), not a read+store-to-i1.
	if !strings.Contains(ir, "call void @__kylix_htab_put") {
		t.Errorf("expected `b.Seen['x'] := true` to emit htab_put\nIR:\n%s", ir)
	}
	// Must not store an i1 into a ptr-typed register (the old broken path).
	if strings.Contains(ir, "store i1 ") {
		// A `store i1 …, ptr %tN` (ptr operand is an i1 icmp result) is the bug.
		// Allow `store i1` only into an i1 alloca (a real Boolean local), not
		// into a ptr register. Heuristic: the bug's signature is a `store i1`
		// whose pointer operand is a `%tN` (icmp/tmp), not a `%v_*_bool` alloca.
		for _, line := range strings.Split(ir, "\n") {
			l := strings.TrimSpace(line)
			if strings.HasPrefix(l, "store i1 ") && strings.Contains(l, ", ptr %t") {
				t.Errorf("map-field Boolean assign must not store i1 into a ptr tmp (was the read+store bug)\nline: %s\nIR:\n%s", l, ir)
				break
			}
		}
	}
}

// TestMapIndexPut_BooleanZext (v5.6.0) guards the fix for `m[k] := true` on a
// local map[String]Boolean. The Boolean `true` (i1) must be zext'd to i64
// before snprintf("%lld"), else llc errors (`i1 defined but expected i64`).
func TestMapIndexPut_BooleanZext(t *testing.T) {
	ir := generateIR(t, `program p;
var m: map[String]Boolean;
begin
  m['x'] := true;
end.`)
	// The put must zext the i1 to i64 before stringifying (snprintf expects i64).
	if !strings.Contains(ir, "zext i1") {
		t.Errorf("expected `m['x'] := true` to zext i1→i64 before stringifying\nIR:\n%s", ir)
	}
	if !strings.Contains(ir, "call void @__kylix_htab_put") {
		t.Errorf("expected a htab_put for `m['x'] := true`\nIR:\n%s", ir)
	}
}
