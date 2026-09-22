package compiler_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"kylix/pkg/compiler"
)

// writeTempKlx writes a minimal .klx file to a temp dir and returns the path.
func writeTempKlx(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCache_StoreAndLoad(t *testing.T) {
	dir := t.TempDir()
	cache := compiler.NewBuildCache(dir)

	// Write a source file
	src := writeTempKlx(t, dir, "foo.klx", "unit foo;\n")

	// Nothing cached yet
	if e := cache.Load(src); e != nil {
		t.Fatal("expected nil for uncached file")
	}

	// Store
	cache.Store(src, "package main\n")

	// Load — should hit
	entry := cache.Load(src)
	if entry == nil {
		t.Fatal("expected cache hit after Store")
	}
	if entry.GoCode != "package main\n" {
		t.Errorf("GoCode mismatch: %q", entry.GoCode)
	}
}

func TestCache_StaleOnModTime(t *testing.T) {
	dir := t.TempDir()
	cache := compiler.NewBuildCache(dir)
	src := writeTempKlx(t, dir, "bar.klx", "unit bar;\n")

	cache.Store(src, "// v1")

	// Touch the file to change mtime
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(src, []byte("unit bar; // changed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if e := cache.Load(src); e != nil {
		t.Error("expected cache miss after file modification")
	}
}

// v0.12.0: one file's generated code can depend on another's annotations (the
// [Entity] scan reads every program), so a build-wide fingerprint is part of
// the key — editing a unit must invalidate the fragments of the others.
func TestCache_StaleOnBuildFingerprint(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.klx")
	if err := os.WriteFile(src, []byte("program P; begin end.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sibling := filepath.Join(dir, "unit.klx")
	if err := os.WriteFile(sibling, []byte("unit U; end.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	writer := compiler.NewBuildCache(dir)
	writer.SetFingerprint([]string{src, sibling})
	writer.Store(src, "// body")

	reader := compiler.NewBuildCache(dir)
	reader.SetFingerprint([]string{src, sibling})
	if entry := reader.Load(src); entry == nil {
		t.Fatal("expected a hit under the same fingerprint")
	}

	// The sibling changes → every fragment recorded under the old fingerprint
	// must be considered stale, even though src itself did not move.
	if err := os.WriteFile(sibling, []byte("unit U;\ninterface\nend.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	after := compiler.NewBuildCache(dir)
	after.SetFingerprint([]string{src, sibling})
	if entry := after.Load(src); entry != nil {
		t.Error("expected a miss after a sibling file changed")
	}
}

func TestCache_Invalidate(t *testing.T) {
	dir := t.TempDir()
	cache := compiler.NewBuildCache(dir)
	src := writeTempKlx(t, dir, "baz.klx", "unit baz;\n")

	cache.Store(src, "// cached")
	cache.Invalidate(src)

	if e := cache.Load(src); e != nil {
		t.Error("expected cache miss after Invalidate")
	}
}

func TestCache_MissingFile(t *testing.T) {
	dir := t.TempDir()
	cache := compiler.NewBuildCache(dir)

	// File doesn't exist — should not panic
	entry := cache.Load("/nonexistent/path/foo.klx")
	if entry != nil {
		t.Error("expected nil for nonexistent file")
	}
}

func TestCompileProject_PackageSearchDirs(t *testing.T) {
	dir := t.TempDir()

	// Create a unit in a "packages/mylib" subdir
	pkgDir := filepath.Join(dir, "packages", "mylib")
	os.MkdirAll(pkgDir, 0755)
	os.WriteFile(filepath.Join(pkgDir, "mylib.klx"), []byte(
		"unit mylib;\nfunction Greet(): String;\nbegin result := 'hi'; end;\n",
	), 0644)

	// Main program using mylib
	main := writeTempKlx(t, dir, "main.klx", `program Test;
uses mylib;
begin
  WriteLn(Greet());
end.
`)

	out := filepath.Join(dir, "out.go")
	result, err := compiler.CompileProject([]string{main}, compiler.Options{
		OutputFile:        out,
		PackageSearchDirs: []string{pkgDir},
	})
	if err != nil {
		t.Fatalf("CompileProject error: %v", err)
	}
	if !result.Success {
		for _, d := range result.Diagnostics {
			t.Logf("diag: %s", d.Message)
		}
		t.Fatal("expected success")
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}
