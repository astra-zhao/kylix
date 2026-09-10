package llvmgen

// v0.7.1 P2: tests for the llvm-mingw sysroot discovery used by the
// windows cross-link path (compile.go).

import (
	"os"
	"path/filepath"
	"testing"
)

// makeFakeMingwRoot creates a directory shaped like an llvm-mingw toolchain
// root: just the x86_64 sysroot with a lib dir inside.
func makeFakeMingwRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	syslib := filepath.Join(root, "x86_64-w64-mingw32", "lib")
	if err := os.MkdirAll(syslib, 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestFindMingwSysroot_EnvOverride(t *testing.T) {
	root := makeFakeMingwRoot(t)
	t.Setenv("KYLIX_MINGW_ROOT", root)
	if got := FindMingwSysroot(); got != root {
		t.Errorf("FindMingwSysroot() = %q, want %q (KYLIX_MINGW_ROOT)", got, root)
	}
}

func TestFindMingwSysroot_NestedVersionedDir(t *testing.T) {
	// mstorsjo tarballs unpack to llvm-mingw-<ver>-ucrt-<platform>/ — pointing
	// KYLIX_MINGW_ROOT at the parent should still resolve one level down.
	parent := t.TempDir()
	syslib := filepath.Join(parent, "llvm-mingw-20260826-ucrt-macos-universal", "x86_64-w64-mingw32", "lib")
	if err := os.MkdirAll(syslib, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KYLIX_MINGW_ROOT", parent)
	got := FindMingwSysroot()
	want := filepath.Join(parent, "llvm-mingw-20260826-ucrt-macos-universal")
	if got != want {
		t.Errorf("FindMingwSysroot() = %q, want %q (nested versioned dir)", got, want)
	}
}

func TestFindMingwSysroot_Missing(t *testing.T) {
	// Point the env at an empty dir: no sysroot marker -> not found.
	empty := t.TempDir()
	t.Setenv("KYLIX_MINGW_ROOT", empty)
	if got := FindMingwSysroot(); got != "" {
		t.Errorf("FindMingwSysroot() = %q, want \"\" (no sysroot present)", got)
	}
}
