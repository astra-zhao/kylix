package llvmgen_test

import (
	"os"
	"strings"
	"testing"

	"kylix/pkg/llvmgen"
)

// gc_test.go — tests for the Boehm GC mode (--gc=boehm, v0.10.0).
//
// Contract:
//   - GC mode: user-data allocations (class instances, records, strings,
//     arrays, Variant boxes, closures) route through @GC_malloc; the runtime
//     declares GC_malloc/GC_realloc.
//   - Default mode: IR is byte-identical to the pre-GC emitter — no GC_malloc
//     references, no GC declares (protects the IR fixed point and the baked
//     stdlib IR).

const gcSampleSrc = `program p;
type
  TPoint = class
    X, Y: Integer;
    constructor Create(ax, ay: Integer);
    begin
      X := ax;
      Y := ay;
    end;
  end;
var
  i: Integer;
  s: String;
  pt: TPoint;
begin
  for i := 1 to 100 do
  begin
    pt := TPoint.Create(i, i * 2);
    s := 'item-' + IntToStr(i);
    WriteLn(s, ' ', pt.X + pt.Y);
  end;
end.`

func generateIRWithGC(t *testing.T, src string) string {
	t.Helper()
	ir, err := llvmgen.GenerateWithOpts(mustParse(t, src), "test.klx", llvmgen.CompileOpts{GC: "boehm"})
	if err != nil {
		t.Fatalf("GenerateWithOpts(gc) failed: %v\nIR:\n%s", err, ir)
	}
	return ir
}

func TestGC_DeclaresGCSymbols(t *testing.T) {
	ir := generateIRWithGC(t, gcSampleSrc)
	if !strings.Contains(ir, "declare ptr @GC_malloc(i64 noundef)") {
		t.Errorf("GC mode missing @GC_malloc declare\nIR:\n%s", ir)
	}
	if !strings.Contains(ir, "declare ptr @GC_realloc(ptr noundef, i64 noundef)") {
		t.Errorf("GC mode missing @GC_realloc declare\nIR:\n%s", ir)
	}
}

func TestGC_ClassConstructorUsesGCAlloc(t *testing.T) {
	// Class instances are calloc'd by default; in GC mode they must come from
	// GC_malloc (zeroed, scanned by the collector).
	ir := generateIRWithGC(t, gcSampleSrc)
	if !strings.Contains(ir, "call ptr @GC_malloc(i64") {
		t.Errorf("GC mode: class constructor did not route through GC_malloc\nIR:\n%s", ir)
	}
}

func TestGC_StringConcatUsesGCAlloc(t *testing.T) {
	ir := generateIRWithGC(t, gcSampleSrc)
	// emitStringConcat is the highest-frequency user-data allocation.
	n := strings.Count(ir, "call ptr @GC_malloc(i64")
	if n < 2 {
		t.Errorf("GC mode: expected several GC_malloc sites (class + string concat), got %d\nIR:\n%s", n, ir)
	}
}

func TestGC_DefaultIRHasNoGCSymbols(t *testing.T) {
	// The default (non-GC) path must stay byte-identical to the pre-GC
	// emitter: no GC_malloc calls, no GC declares anywhere.
	ir := generateIR(t, gcSampleSrc)
	if strings.Contains(ir, "GC_malloc") || strings.Contains(ir, "GC_realloc") {
		t.Errorf("default mode leaked GC symbols into IR\nIR:\n%s", ir)
	}
	if !strings.Contains(ir, "call ptr @malloc(i64") {
		t.Errorf("default mode no longer allocates via malloc?\nIR:\n%s", ir)
	}
}

func TestGC_GCAndDefaultDifferOnlyInAllocSymbols(t *testing.T) {
	// Both modes compile the same program; the GC variant must differ from
	// the default variant only by malloc/calloc → GC_malloc (plus declares).
	def := generateIR(t, gcSampleSrc)
	gc := generateIRWithGC(t, gcSampleSrc)
	if def == gc {
		t.Fatal("GC and default IR are identical — the GC gate is not wired")
	}
	// Normalize allocation symbols to a single token. In GC mode the zalloc
	// sites (default: calloc) also route to GC_malloc — GC_malloc zero-fills
	// like calloc — so malloc/calloc/GC_malloc must all collapse to one token.
	// The token is chosen so it can never collide with IR text.
	tok := "@@ALLOC@@"
	norm := func(s string, syms ...string) string {
		for _, sym := range syms {
			s = strings.ReplaceAll(s, sym, tok)
		}
		return s
	}
	allocForms := []string{
		"call ptr @malloc(i64",
		"call ptr @calloc(i64 1, i64",
		"call ptr @GC_malloc(i64",
	}
	defNorm := norm(def, allocForms...)
	gcNorm := norm(gc, allocForms...)
	// GC mode additionally declares GC_realloc — strip the two declare lines
	// before comparing.
	gcNorm = strings.ReplaceAll(gcNorm, "declare ptr @GC_malloc(i64 noundef)\n", "")
	gcNorm = strings.ReplaceAll(gcNorm, "declare ptr @GC_realloc(ptr noundef, i64 noundef)\n", "")
	if defNorm != gcNorm {
		t.Errorf("GC mode changed more than the allocation symbols\n--- default ---\n%s\n--- gc ---\n%s", defNorm, gcNorm)
	}
}

func TestGC_CacheKeyIncludesGCOption(t *testing.T) {
	// GC/non-GC objects must never collide in the incremental cache.
	src := t.TempDir() + "/p.klx"
	if err := os.WriteFile(src, []byte("program p;\nbegin\n  WriteLn('hi');\nend."), 0644); err != nil {
		t.Fatal(err)
	}
	k1, err := llvmgen.ComputeCacheKey(src, llvmgen.CompileOpts{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	k2, err := llvmgen.ComputeCacheKey(src, llvmgen.CompileOpts{GC: "boehm"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if k1 == k2 {
		t.Error("cache key ignores the GC option — GC/non-GC objects would be reused across modes")
	}
}

func TestGC_RejectsWindowsTarget(t *testing.T) {
	// llvm-mingw has no libgc — the windows target must be rejected with a
	// clear error (both here and at the CLI layer).
	llvmPaths, err := llvmgen.FindLLVM()
	if err != nil {
		t.Skip("LLVM toolchain not available")
	}
	// CompileToNativeOpts surfaces the error at the link stage; use a source
	// file in tmp so the pipeline can parse it.
	dir := t.TempDir()
	srcPath := dir + "/p.klx"
	code := "program p;\nbegin\n  WriteLn('hi');\nend."
	if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = llvmgen.CompileToNativeOpts(srcPath, dir+"/p", llvmPaths, llvmgen.CompileOpts{GC: "boehm", Target: "windows/amd64"})
	if err == nil {
		t.Fatal("expected error for --gc=boehm + windows target")
	}
	if !strings.Contains(err.Error(), "--gc=boehm is not supported for the windows target") {
		t.Errorf("unexpected error: %v", err)
	}
}
