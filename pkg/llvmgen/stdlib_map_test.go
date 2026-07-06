package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_map tests — verify map[K]V language-level type lowers to the
// hash-table runtime (htab_new/htab_get/htab_put).

func TestMap_VarDeclInitializesHtab(t *testing.T) {
	ir := generateIR(t, `program p;
var m: map[String]Integer;
begin
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_htab_new")
	if strings.Contains(ir, "is not an array") {
		t.Errorf("map var decl should not hit array path\nIR:\n%s", ir)
	}
}

func TestMap_IndexGetRoutesToHtabGet(t *testing.T) {
	ir := generateIR(t, `program p;
var m: map[String]Integer;
begin
  var s := m['key'];
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_htab_get")
}

func TestMap_IndexPutRoutesToHtabPut(t *testing.T) {
	ir := generateIR(t, `program p;
var m: map[String]Integer;
begin
  m['key'] := 42;
end.`)
	assertIRContains(t, ir, "call void @__kylix_htab_put")
}

func TestMap_IntegerValueStringified(t *testing.T) {
	// m['key'] := 42 → IntToStr (snprintf %lld) before htab_put
	ir := generateIR(t, `program p;
var m: map[String]Integer;
begin
  m['key'] := 42;
end.`)
	assertIRContains(t, ir, "snprintf")
}

func TestMap_StringValueDirect(t *testing.T) {
	// String values go directly to htab_put (no IntToStr)
	ir := generateIR(t, `program p;
var m: map[String]String;
begin
  m['key'] := 'val';
end.`)
	assertIRContains(t, ir, "call void @__kylix_htab_put")
}

func TestMap_RoundTrip(t *testing.T) {
	// Put + Get in same program
	ir := generateIR(t, `program p;
var m: map[String]Integer;
begin
  m['key'] := 42;
  var s := m['key'];
end.`)
	assertIRContains(t, ir, "call void @__kylix_htab_put")
	assertIRContains(t, ir, "call ptr @__kylix_htab_get")
}
