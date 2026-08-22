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

func TestCache_PutWithTTLUsesCachePut(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
var c := NewCache(10, 0);
begin
  c.PutWithTTL('k', 'v', 5000);
end.`)
	// PutWithTTL routes to the shared @__kylix_cache_put helper (which stores
	// a TTL record when ttlMs > 0), and that helper uses the millisecond clock.
	assertIRContains(t, ir, "call void @__kylix_cache_put(ptr")
	assertIRContains(t, ir, "define void @__kylix_cache_put(ptr %h, ptr %k, ptr %v, i64 %ttlMs)")
	assertIRContains(t, ir, "call i64 @__kylix_now_ms()")
}

func TestCache_GetReturnsVariantBox(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
var c := NewCache(10, 0);
begin
  c.Put('k', 'v');
  var v := c.Get('k');
end.`)
	// Get consults the expiry guard and returns a Variant (box_str / nilbox).
	assertIRContains(t, ir, "call i1 @__kylix_cache_expired(ptr")
	assertIRContains(t, ir, "define i1 @__kylix_cache_expired(ptr %h, ptr %k)")
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_str(ptr")
}

func TestCache_SweepUsesHtabKeys(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
var c := NewCache(10, 0);
begin
  c.PutWithTTL('k', 'v', 5000);
  var removed := c.Sweep();
end.`)
	// Sweep walks every TTL-tracked key via htab_keys.
	assertIRContains(t, ir, "call i64 @__kylix_cache_sweep(ptr")
	assertIRContains(t, ir, "define i64 @__kylix_cache_sweep(ptr %h)")
	assertIRContains(t, ir, "call { ptr, i64 } @__kylix_htab_keys(ptr")
}

func TestCache_NowMsUsesGetTimeOfDay(t *testing.T) {
	ir := generateIR(t, `program p;
uses cache;
var c := NewCache(10, 1000);
begin
  c.Put('k', 'v');
end.`)
	// Millisecond clock is wall-clock via gettimeofday (CLOCK_MONOTONIC
	// constant differs across platforms).
	assertIRContains(t, ir, "define i64 @__kylix_now_ms()")
	assertIRContains(t, ir, "call i32 @gettimeofday(ptr")
}
