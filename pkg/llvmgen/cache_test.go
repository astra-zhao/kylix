package llvmgen_test

import (
	"os"
	"path/filepath"
	"testing"

	"kylix/pkg/llvmgen"
)

// cache_test.go — tests for the LLVM incremental compilation cache (v4.5.0).

func TestCache_PutAndGet(t *testing.T) {
	dir := t.TempDir()
	store, err := llvmgen.NewCacheStore(dir)
	if err != nil {
		t.Fatalf("NewCacheStore: %v", err)
	}
	srcObj := filepath.Join(dir, "src.o")
	if err := os.WriteFile(srcObj, []byte("OBJCONTENT"), 0644); err != nil {
		t.Fatal(err)
	}
	key := llvmgen.CacheKey("deadbeef")
	if err := store.Put(key, srcObj); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got := store.Get(key)
	if got == "" {
		t.Fatal("Get returned empty after Put")
	}
	b, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("read cached obj: %v", err)
	}
	if string(b) != "OBJCONTENT" {
		t.Errorf("cached content mismatch: %q", b)
	}
}

func TestCache_MissReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	store, err := llvmgen.NewCacheStore(dir)
	if err != nil {
		t.Fatalf("NewCacheStore: %v", err)
	}
	if got := store.Get(llvmgen.CacheKey("nonexistent")); got != "" {
		t.Errorf("Get returned %q for missing key, want empty", got)
	}
}

func TestCache_OverwriteOnReput(t *testing.T) {
	// Same key, new content → Put overwrites the cached object.
	dir := t.TempDir()
	store, err := llvmgen.NewCacheStore(dir)
	if err != nil {
		t.Fatalf("NewCacheStore: %v", err)
	}
	key := llvmgen.CacheKey("k1")
	obj1 := filepath.Join(dir, "o1.o")
	os.WriteFile(obj1, []byte("V1"), 0644)
	store.Put(key, obj1)
	obj2 := filepath.Join(dir, "o2.o")
	os.WriteFile(obj2, []byte("V2"), 0644)
	store.Put(key, obj2)
	got := store.Get(key)
	b, _ := os.ReadFile(got)
	if string(b) != "V2" {
		t.Errorf("expected overwrite to V2, got %q", b)
	}
}

func TestCache_DepHashes(t *testing.T) {
	// DepHashes reads files and returns content hashes; different content →
	// different hashes.
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.klx")
	f2 := filepath.Join(dir, "b.klx")
	os.WriteFile(f1, []byte("content-a"), 0644)
	os.WriteFile(f2, []byte("content-b"), 0644)
	hashes, err := llvmgen.DepHashes([]string{f1, f2})
	if err != nil {
		t.Fatalf("DepHashes: %v", err)
	}
	if len(hashes) != 2 || hashes[0] == hashes[1] {
		t.Errorf("DepHashes did not produce distinct hashes: %v", hashes)
	}
}
