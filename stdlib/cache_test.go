package stdlib

import (
	"testing"
	"time"
)

func TestCache_PutGet(t *testing.T) {
	c := NewCache(4, 0)
	c.Put("a", 1)
	c.Put("b", "hello")
	if v, ok := c.Get("a"); !ok || v.(int) != 1 {
		t.Errorf("Get(a) = %v, %v; want 1, true", v, ok)
	}
	if !c.Has("b") {
		t.Error("Has(b) should be true")
	}
	if c.Has("missing") {
		t.Error("Has(missing) should be false")
	}
}

func TestCache_LRU_Eviction(t *testing.T) {
	c := NewCache(2, 0)
	c.Put("a", 1)
	c.Put("b", 2)
	// Access "a" so "b" becomes least-recently-used.
	c.Get("a")
	c.Put("c", 3) // capacity 2 → evict "b"

	if c.Has("b") {
		t.Error("b should have been evicted as LRU")
	}
	if !c.Has("a") || !c.Has("c") {
		t.Error("a and c should still be present")
	}
	if c.Size() != 2 {
		t.Errorf("Size = %d, want 2", c.Size())
	}
}

func TestCache_UpdateExistingKey(t *testing.T) {
	c := NewCache(4, 0)
	c.Put("k", "old")
	c.Put("k", "new")
	if v, _ := c.Get("k"); v != "new" {
		t.Errorf("Get(k) = %v, want new", v)
	}
	if c.Size() != 1 {
		t.Errorf("Size = %d, want 1 (update should not grow size)", c.Size())
	}
}

func TestCache_TTL_Expiry(t *testing.T) {
	c := NewCache(4, 0)
	c.PutWithTTL("temp", "x", 50) // 50ms TTL
	if !c.Has("temp") {
		t.Fatal("temp should be present immediately")
	}
	time.Sleep(80 * time.Millisecond)
	if c.Has("temp") {
		t.Error("temp should have expired")
	}
	// Non-TTL entry survives.
	c.Put("perm", "y")
	time.Sleep(20 * time.Millisecond)
	if !c.Has("perm") {
		t.Error("perm (no TTL) should still be present")
	}
}

func TestCache_DefaultTTLFromConstructor(t *testing.T) {
	c := NewCache(4, 50) // default 50ms TTL
	c.Put("k", "v")
	if !c.Has("k") {
		t.Fatal("k should be present immediately")
	}
	time.Sleep(80 * time.Millisecond)
	if c.Has("k") {
		t.Error("k should have expired via default TTL")
	}
}

func TestCache_Sweep(t *testing.T) {
	c := NewCache(8, 0)
	c.PutWithTTL("a", 1, 50)
	c.PutWithTTL("b", 2, 50)
	c.Put("perm", 3) // no TTL
	time.Sleep(80 * time.Millisecond)

	removed := c.Sweep()
	if removed != 2 {
		t.Errorf("Sweep removed %d, want 2", removed)
	}
	if !c.Has("perm") {
		t.Error("perm should survive sweep")
	}
}

func TestCache_DeleteAndClear(t *testing.T) {
	c := NewCache(4, 0)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Delete("a")
	if c.Has("a") {
		t.Error("a should be deleted")
	}
	c.Clear()
	if c.Size() != 0 {
		t.Errorf("Size after Clear = %d, want 0", c.Size())
	}
}

func TestCache_GetString(t *testing.T) {
	c := NewCache(4, 0)
	c.Put("s", "value")
	if c.GetString("s") != "value" {
		t.Errorf("GetString(s) = %q, want value", c.GetString("s"))
	}
	if c.GetString("missing") != "" {
		t.Errorf("GetString(missing) = %q, want empty", c.GetString("missing"))
	}
}

func TestCache_DefaultCapacityGuard(t *testing.T) {
	c := NewCache(0, 0) // invalid capacity → default 16
	c.Put("a", 1)
	if !c.Has("a") {
		t.Error("cache with default capacity should still work")
	}
}
