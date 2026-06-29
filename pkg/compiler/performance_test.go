package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestIncrementalCompilationPerformance verifies that incremental compilation
// with cache is significantly faster than full recompilation.
func TestIncrementalCompilationPerformance(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, ".cache")

	// Create 10 Kylix files to simulate a medium-sized project
	files := make([]string, 10)
	for i := 0; i < 10; i++ {
		src := fmt.Sprintf(`unit unit%d;
interface
function Func%d(): Integer;
implementation
function Func%d(): Integer;
begin
  result := %d;
end;
end.`, i, i, i, i*10)
		filePath := filepath.Join(tmpDir, fmt.Sprintf("unit%d.klx", i))
		if err := os.WriteFile(filePath, []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
		files[i] = filePath
	}

	// Main program that uses all units
	mainSrc := `program Main;
uses unit0, unit1, unit2, unit3, unit4, unit5, unit6, unit7, unit8, unit9;
begin
  WriteLn(Func0() + Func5());
end.`
	mainFile := filepath.Join(tmpDir, "main.klx")
	if err := os.WriteFile(mainFile, []byte(mainSrc), 0644); err != nil {
		t.Fatal(err)
	}
	files = append(files, mainFile)

	opts := Options{
		OutputFile: filepath.Join(tmpDir, "main.go"),
		CacheDir:   cacheDir,
	}

	// First compile (cold cache)
	t.Log("First compile (cold cache)...")
	start := time.Now()
	result, err := CompileProject(files, opts)
	coldDuration := time.Since(start)
	if err != nil {
		t.Fatalf("First compile failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("First compile failed: %+v", result.Diagnostics)
	}
	t.Logf("Cold compile: %v", coldDuration)

	// Second compile (warm cache, no changes)
	t.Log("Second compile (warm cache, unchanged files)...")
	start = time.Now()
	result, err = CompileProject(files, opts)
	warmDuration := time.Since(start)
	if err != nil {
		t.Fatalf("Second compile failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("Second compile failed: %+v", result.Diagnostics)
	}
	t.Logf("Warm compile: %v", warmDuration)

	// Speedup ratio
	speedup := float64(coldDuration) / float64(warmDuration)
	t.Logf("Speedup: %.1fx", speedup)

	// Incremental compilation should be at least 2x faster
	// (relaxed threshold for CI environments with variable I/O latency)
	if speedup < 2.0 {
		t.Errorf("incremental compile not fast enough: %.1fx speedup (expected at least 2x)", speedup)
	}

	// Modify one file and recompile (partial cache hit)
	t.Log("Third compile (one file changed)...")
	modifiedSrc := `unit unit5;
interface
function Func5(): Integer;
implementation
function Func5(): Integer;
begin
  result := 999;  // changed
end;
end.`
	if err := os.WriteFile(files[5], []byte(modifiedSrc), 0644); err != nil {
		t.Fatal(err)
	}

	start = time.Now()
	result, err = CompileProject(files, opts)
	partialDuration := time.Since(start)
	if err != nil {
		t.Fatalf("Third compile failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("Third compile failed: %+v", result.Diagnostics)
	}
	t.Logf("Partial cache compile: %v", partialDuration)

	// Partial cache should still be faster than cold
	partialSpeedup := float64(coldDuration) / float64(partialDuration)
	t.Logf("Partial cache speedup vs cold: %.1fx", partialSpeedup)
	if partialSpeedup < 1.5 {
		t.Errorf("partial cache not effective: %.1fx speedup (expected at least 1.5x)", partialSpeedup)
	}
}
