package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_cache tests — verify TCache lowers to the internal hash-table
// runtime (@__kylix_htab_*) rather than stubs.

func TestCache_NewCacheCallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
begin
  var c := NewCache(4, 0);
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_htab_new")
	if strings.Contains(ir, "cache.NewCache not implemented") {
		t.Errorf("NewCache still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestCache_PutMethodDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
begin
  var c := NewCache(4, 0);
  c.Put('k', 'v');
end.`)
	assertIRContains(t, ir, "call void @__kylix_htab_put")
}

func TestCache_GetStringMethodDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
begin
  var c := NewCache(4, 0);
  var s := c.GetString('k');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_htab_get")
}

func TestCache_HasMethodDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
begin
  var c := NewCache(4, 0);
  var ok := c.Has('k');
end.`)
	assertIRContains(t, ir, "call i1 @__kylix_htab_has")
}

func TestCache_DeleteMethodDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
begin
  var c := NewCache(4, 0);
  c.Delete('k');
end.`)
	assertIRContains(t, ir, "call void @__kylix_htab_del")
}

func TestCache_SizeMethodDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
begin
  var c := NewCache(4, 0);
  var n := c.Size();
end.`)
	assertIRContains(t, ir, "call i64 @__kylix_htab_size")
}

func TestCache_ClearMethodDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
begin
  var c := NewCache(4, 0);
  c.Clear();
end.`)
	assertIRContains(t, ir, "call void @__kylix_htab_clear")
}

func TestCache_HashtabBodiesEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
begin
  var c := NewCache(4, 0);
  c.Put('k', 'v');
end.`)
	// All hash-table runtime helpers should be emitted once cache is used.
	assertIRContains(t, ir, "define ptr @__kylix_htab_new()")
	assertIRContains(t, ir, "define i64 @__kylix_htab_hash(ptr %key)")
	assertIRContains(t, ir, "define ptr @__kylix_htab_find(ptr %t, ptr %key)")
	assertIRContains(t, ir, "define void @__kylix_htab_put(ptr %t, ptr %key, ptr %val)")
	assertIRContains(t, ir, "define ptr @__kylix_htab_strdup(ptr %s)")
}

func TestCache_NotUsedNoHashtab(t *testing.T) {
	// A program that does NOT `uses cache` should not emit the hash-table runtime.
	ir := generateIR(t, `program p;
begin
  WriteLn('hi');
end.`)
	if strings.Contains(ir, "@__kylix_htab_") {
		t.Errorf("hash-table runtime emitted without `uses cache`\nIR:\n%s", ir)
	}
}
