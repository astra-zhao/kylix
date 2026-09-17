package llvmgen

import (
	"testing"
)

func TestLLVMPaths_OptFieldOptional(t *testing.T) {
	// FindLLVM resolves opt if present, but Opt may be "" when opt isn't
	// installed — that must not be a fatal error (clang is required, llc/opt
	// are optional and the pipeline degrades gracefully: no llc → the .ll →
	// .o step falls back to `clang -x ir -c`, e.g. llvm-mingw native Windows).
	p, err := FindLLVM()
	if err != nil {
		t.Skipf("LLVM toolchain not installed on this machine: %v", err)
	}
	if p.Clang == "" {
		t.Fatalf("clang must be resolved, got Clang=%q LLC=%q", p.Clang, p.LLC)
	}
	if p.LLC == "" {
		t.Logf("llc not found — IR→object via clang -x ir fallback (OK)")
	}
	// Opt is allowed to be empty (opt binary may be absent); the compile
	// pipeline checks `paths.Opt != ""` before invoking it.
	if p.Opt == "" {
		t.Logf("opt not found — --llvm-opt will fall back to llc -O<N> only")
	}
}

func TestCompileOpts_OptLevelPassthrough(t *testing.T) {
	// CompileOpts.OptLevel is a plain string field consumed by
	// CompileToNativeOpts; verify it can hold the documented levels without
	// any normalization at construction time (the clamping happens inline).
	levels := []string{"", "1", "2", "3", "s", "z"}
	for _, lvl := range levels {
		opts := CompileOpts{OptLevel: lvl}
		if opts.OptLevel != lvl {
			t.Errorf("OptLevel not preserved: got %q want %q", opts.OptLevel, lvl)
		}
	}
}
