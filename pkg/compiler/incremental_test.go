package compiler_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"kylix/pkg/compiler"
)

// Incremental compilation tests (Task 4).

func TestIncremental_CacheHit(t *testing.T) {
	dir := t.TempDir()
	mathFile := filepath.Join(dir, "math.klx")
	mainFile := filepath.Join(dir, "main.klx")

	os.WriteFile(mathFile, []byte(`unit math;
function Square(n: Integer): Integer; begin result := n * n; end;
`), 0644)
	os.WriteFile(mainFile, []byte(`program Main;
uses math;
begin WriteLn(Square(5)); end.
`), 0644)

	opts := compiler.Options{
		OutputFile: filepath.Join(dir, "out.go"),
		CacheDir:   dir,
	}

	// First build: cold cache
	r1, err := compiler.CompileProject([]string{mathFile, mainFile}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !r1.Success {
		t.Fatalf("first build failed: %v", r1.Diagnostics)
	}

	// Verify cache files were created
	cacheDir := filepath.Join(dir, ".kylix-cache")
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatalf("cache dir not created: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 cache entries, got %d", len(entries))
	}

	// Second build: should use cache (no failures, no time spent on parse+gen)
	r2, err := compiler.CompileProject([]string{mathFile, mainFile}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !r2.Success {
		t.Fatalf("warm build failed: %v", r2.Diagnostics)
	}
}

func TestIncremental_PartialRebuild(t *testing.T) {
	dir := t.TempDir()
	mathFile := filepath.Join(dir, "math.klx")
	mainFile := filepath.Join(dir, "main.klx")

	os.WriteFile(mathFile, []byte(`unit math;
function Square(n: Integer): Integer; begin result := n * n; end;
`), 0644)
	os.WriteFile(mainFile, []byte(`program Main;
uses math;
begin WriteLn(Square(5)); end.
`), 0644)

	opts := compiler.Options{
		OutputFile: filepath.Join(dir, "out.go"),
		CacheDir:   dir,
	}

	// Cold build
	if r, err := compiler.CompileProject([]string{mathFile, mainFile}, opts); err != nil || !r.Success {
		t.Fatalf("cold build failed")
	}

	// Modify math.klx to invalidate its cache entry
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(mathFile, []byte(`unit math;
function Square(n: Integer): Integer; begin result := n * n + 1; end;
`), 0644)

	// Should still build successfully — cache is partially invalidated
	r3, err := compiler.CompileProject([]string{mathFile, mainFile}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !r3.Success {
		t.Fatalf("partial rebuild failed: %v", r3.Diagnostics)
	}
}

func TestIncremental_CacheInvalidation(t *testing.T) {
	dir := t.TempDir()
	mainFile := filepath.Join(dir, "main.klx")
	os.WriteFile(mainFile, []byte(`program Main;
begin WriteLn('hi'); end.
`), 0644)

	opts := compiler.Options{
		OutputFile: filepath.Join(dir, "out.go"),
		CacheDir:   dir,
	}

	// Build once → cache populated
	compiler.CompileProject([]string{mainFile}, opts)

	cache := compiler.NewBuildCache(dir)
	if entry := cache.Load(mainFile); entry == nil {
		t.Fatal("expected cache hit after first build")
	}

	// Manually invalidate
	cache.Invalidate(mainFile)
	if entry := cache.Load(mainFile); entry != nil {
		t.Error("expected nil after Invalidate")
	}
}

func TestIncremental_FingerprintBasedOnSize(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "test.klx")
	os.WriteFile(src, []byte("unit test;\n"), 0644)

	cache := compiler.NewBuildCache(dir)
	cache.Store(src, "package main")

	// Same size + mtime → cache hit
	if entry := cache.Load(src); entry == nil {
		t.Error("expected cache hit immediately after Store")
	}

	// Modify content → different size → cache miss
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(src, []byte("unit test;\n// extra comment\n"), 0644)

	if entry := cache.Load(src); entry != nil {
		t.Error("expected cache miss after content change (different size)")
	}
}
