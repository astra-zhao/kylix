package llvmgen

import (
	"fmt"
	"os/exec"
	"strings"
)

// linkWasi links a wasm32 object into a WASI command (or a reactor when
// shared). No libc, no libcrypto/sqlite/curl. The only imports are the
// wasi_snapshot_preview1 table emitted in the IR.
func linkWasi(llvmPaths *LLVMPaths, outBin, objFile, irFile string) (*CompileResult, error) {
	clangArgs := []string{
		"--target=wasm32-unknown-wasi",
		"-nostdlib",
		"-Wl,--no-entry",
		"-Wl,--export-all",
		"-o", outBin,
		objFile,
	}
	if out, err := exec.Command(llvmPaths.Clang, clangArgs...).CombinedOutput(); err != nil {
		ld, ldErr := exec.LookPath("wasm-ld")
		if ldErr != nil {
			return nil, fmt.Errorf("wasm link failed (clang --target=wasm32-unknown-wasi, and wasm-ld is not on PATH): %w\n%s", err, out)
		}
		ldArgs := []string{"--no-entry", "--export-all", "-o", outBin, objFile}
		out2, err2 := exec.Command(ld, ldArgs...).CombinedOutput()
		if err2 != nil {
			return nil, fmt.Errorf("wasm link failed: %w\nclang: %s\nwasm-ld: %s", err2, out, out2)
		}
	}
	return &CompileResult{IRFile: irFile, ObjFile: objFile, BinFile: outBin}, nil
}

// WasiLLVMTarget reports whether a --target value selects wasm32-unknown-wasi.
func WasiLLVMTarget(target string) bool {
	switch strings.TrimSpace(target) {
	case "wasm", "wasm32", "wasi", "wasi/wasm32", "wasm/wasm32":
		return true
	default:
		return false
	}
}
