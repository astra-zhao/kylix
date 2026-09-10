package compiler

// Internal tests for the compiler-binary self-hash fingerprint (v0.7.1 P4).
// Lives in the internal package so it can swap codegenHashOverride — the
// public tests can't reach it.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCache_StaleOnCodegenHash(t *testing.T) {
	dir := t.TempDir()
	cache := NewBuildCache(dir)

	src := filepath.Join(dir, "gen.klx")
	if err := os.WriteFile(src, []byte("unit gen;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Use a deterministic override so the test doesn't depend on hashing the
	// real test binary.
	old := codegenHashOverride
	codegenHashOverride = "aaaa111122223333"
	defer func() { codegenHashOverride = old }()

	cache.Store(src, "// from old compiler")
	if e := cache.Load(src); e == nil {
		t.Fatal("expected hit while codegen hash is unchanged")
	}

	// Simulate a rebuilt compiler.
	codegenHashOverride = "bbbb444455556666"
	if e := cache.Load(src); e != nil {
		t.Error("expected cache miss after the compiler hash changed")
	}

	// Re-store under the new hash, then go back to the old: entries are keyed
	// to the hash they were produced by, in both directions.
	cache.Store(src, "// from new compiler")
	codegenHashOverride = "aaaa111122223333"
	if e := cache.Load(src); e != nil {
		t.Error("expected miss when entry was produced by a different binary")
	}
	codegenHashOverride = "bbbb444455556666"
	if e := cache.Load(src); e == nil {
		t.Error("expected hit for the entry produced by the matching binary")
	}
}
