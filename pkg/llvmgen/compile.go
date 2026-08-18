// compile.go — end-to-end LLVM compilation pipeline.
// Kylix source → AST → LLVM IR (.ll) → object (.o) → native binary
package llvmgen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
)

// LLVMPaths holds the LLVM tool locations.
type LLVMPaths struct {
	LLC   string // llc binary path
	Clang string // clang binary path
	Opt   string // opt binary path (optional; IR-level optimization)
}

// FindLLVM looks for llc, clang, and opt in common install locations.
func FindLLVM() (*LLVMPaths, error) {
	searchDirs := []string{
		`C:\LLVM\bin`,                // Windows (CI installs to C:\LLVM, v6.2.0)
		`C:\Program Files\LLVM\bin`,  // Windows (default LLVM installer dir)
		"/opt/homebrew/opt/llvm/bin", // Homebrew ARM
		"/usr/local/opt/llvm/bin",    // Homebrew x86
		"/usr/bin",                   // Linux system
		"/usr/local/bin",
	}
	// v6.3.0: bundled LLVM next to the executable (distribution form B:
	// kylix + llvm/bin + llvm/lib) takes precedence over system installs.
	if exe, err := os.Executable(); err == nil {
		searchDirs = append([]string{filepath.Join(filepath.Dir(exe), "llvm", "bin")}, searchDirs...)
	}

	find := func(name string) string {
		// v6.3.0: bundled LLVM next to the executable (kylix + llvm/bin) has
		// top priority — a self-contained distribution must not pick up a
		// different system clang/llc.
		if exe, err := os.Executable(); err == nil {
			bundled := filepath.Join(filepath.Dir(exe), "llvm", "bin", name)
			if _, err := os.Stat(bundled); err == nil {
				return bundled
			}
		}
		// Try PATH first
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
		for _, dir := range searchDirs {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return ""
	}

	llc := find("llc")
	clang := find("clang")
	opt := find("opt") // optional; only needed for --llvm-opt

	if llc == "" {
		return nil, fmt.Errorf("llc not found; install LLVM (brew install llvm or apt install llvm)")
	}
	if clang == "" {
		return nil, fmt.Errorf("clang not found; install clang (brew install llvm or apt install clang)")
	}

	return &LLVMPaths{LLC: llc, Clang: clang, Opt: opt}, nil
}

// CompileResult holds the output paths from LLVM compilation.
type CompileResult struct {
	IRFile  string // .ll file
	ObjFile string // .o file
	BinFile string // final native binary
}

// CompileToNative runs the full pipeline:
//  1. Parse Kylix source
//  2. Generate LLVM IR
//  3. llc: .ll → .o
//  4. clang: .o → native binary
func CompileToNative(srcFile, outBin string, llvmPaths *LLVMPaths) (*CompileResult, error) {
	return CompileToNativeOpts(srcFile, outBin, llvmPaths, CompileOpts{})
}

// CompileOpts configures optional codegen parameters (e.g., optimization).
type CompileOpts struct {
	// OptLevel selects LLVM optimization tier: "" / "0" / "1" / "2" / "3" / "s".
	// Empty defaults to -O0 (no optimization).
	OptLevel string

	// DebugInfo (v4.5.0): when true, emit DWARF debug info so LLDB/GDB can
	// resolve function names + source files (kylix build --backend=llvm -g).
	// Implies -O0: optimization reorders/drops instructions, making debug info
	// misleading, so OptLevel is forced to "" when DebugInfo is on.
	DebugInfo bool

	// Target (v6.2.0): cross-compilation target "os/arch" (e.g. "linux/amd64",
	// "windows/amd64", "darwin/arm64"). Empty means the host (runtime.GOOS/GOARCH).
	// Drives the LLVM IR target triple + datalayout, llc codegen, clang link
	// flags and the system-library search. See tripleFor.
	Target string
}

// appendHomebrewLib adds -L + -Wl,-rpath for a Homebrew-installed library on
// macOS (openssl/sqlite/curl). No-op on other platforms or when the dir is
// absent (Linux uses the system default path). v6.2.0.
func appendHomebrewLib(clangArgs *[]string, brewName string) {
	for _, dir := range []string{
		"/opt/homebrew/opt/" + brewName + "/lib", // Homebrew ARM
		"/usr/local/opt/" + brewName + "/lib",    // Homebrew x86
	} {
		if _, err := os.Stat(dir); err == nil {
			*clangArgs = append(*clangArgs, "-L"+dir, "-Wl,-rpath,"+dir)
			return
		}
	}
}

// resolveTarget parses a "os/arch" cross-compile target into (os, arch).
// An empty target means the host platform.
func resolveTarget(target string) (string, string) {
	if target == "" {
		return runtime.GOOS, runtime.GOARCH
	}
	if i := strings.Index(target, "/"); i >= 0 {
		return target[:i], target[i+1:]
	}
	return runtime.GOOS, runtime.GOARCH
}

// tripleFor returns the LLVM target triple + datalayout for an os/arch pair
// (v6.2.0). The IR header carries these so llc produces the right object file;
// previously the triple was hardcoded to arm64-apple-macosx, so cross-compiling
// (or even running the LLVM backend on Linux/Windows) emitted Mach-O objects
// that the platform linker could not read.
func tripleFor(osName, arch string) (triple, datalayout string) {
	switch osName + "/" + arch {
	case "darwin/arm64":
		return "arm64-apple-macosx15.0.0", "e-m:o-i64:64-i128:128-n32:64-S128"
	case "darwin/amd64":
		return "x86_64-apple-macosx15.0.0", "e-m:o-p270:32:32-p271:32:32-p272:64:64-i64:64-i128:128-f80:128-n8:16:32:64-S128"
	case "linux/amd64":
		return "x86_64-pc-linux-gnu", "e-m:e-p270:32:32-p271:32:32-p272:64:64-i64:64-i128:128-f80:128-n8:16:32:64-S128"
	case "linux/arm64":
		return "aarch64-unknown-linux-gnu", "e-m:e-i8:8:32-i16:16:32-i64:64-i128:128-n32:64-S128"
	case "windows/amd64":
		return "x86_64-pc-windows-msvc", "e-m:w-p270:32:32-p271:32:32-p272:64:64-i64:64-i128:128-f80:128-n8:16:32:64-S128"
	}
	// Fallback: treat unknown as the arm64 macOS default (backwards compatible).
	return "arm64-apple-macosx15.0.0", "e-m:o-i64:64-i128:128-n32:64-S128"
}

// CompileToNativeOpts compiles with options.
func CompileToNativeOpts(srcFile, outBin string, llvmPaths *LLVMPaths, opts CompileOpts) (*CompileResult, error) {
	// Read and parse source
	src, err := os.ReadFile(srcFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", srcFile, err)
	}

	l := lexer.New(string(src))
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		return nil, fmt.Errorf("parse errors: %s", strings.Join(errs, "; "))
	}

	return compileASTWithOpts(prog, srcFile, outBin, llvmPaths, opts)
}

// CompileASTToNative compiles an already-parsed AST to a native binary.
func CompileASTToNative(prog *ast.Program, srcFile, outBin string, llvmPaths *LLVMPaths) (*CompileResult, error) {
	return compileASTWithOpts(prog, srcFile, outBin, llvmPaths, CompileOpts{})
}

// compileASTWithOpts is the shared implementation that honors CompileOpts.
func compileASTWithOpts(prog *ast.Program, srcFile, outBin string, llvmPaths *LLVMPaths, opts CompileOpts) (*CompileResult, error) {
	// -g implies -O0: optimization reorders/drops instructions, making debug
	// info misleading. Force OptLevel off when DebugInfo is on.
	if opts.DebugInfo && opts.OptLevel != "" {
		opts.OptLevel = ""
	}
	// Generate LLVM IR
	ir, err := GenerateWithOpts(prog, srcFile, opts)
	if err != nil {
		return nil, fmt.Errorf("LLVM IR generation: %w", err)
	}

	// v4.5.0 Phase C: run the process-in-LLVM pass pipeline (ConstantFold +
	// DCE) on the generated IR. These are cheap, always-safe IR-text cleanups
	// that reduce IR/binary size for the common -O0 case and run by default
	// (no flag). They are skipped when external opt --O<N> is set, since opt
	// runs LLVM's own (stronger) DCE/mem2reg/etc. passes on the same IR.
	if opts.OptLevel == "" {
		ir = DefaultPassPipeline().Run(ir)
	}

	// Write .ll file
	base := strings.TrimSuffix(srcFile, filepath.Ext(srcFile))
	irFile := base + ".ll"
	if err := os.WriteFile(irFile, []byte(ir), 0644); err != nil {
		return nil, fmt.Errorf("write IR: %w", err)
	}

	// opt: IR-level optimization (mem2reg, inline, loop opts, DCE, etc.).
	// Runs before llc so the optimizer sees pristine IR. Only invoked when an
	// optimization level is requested and the opt binary is available; falls
	// back to llc's -O flag otherwise.
	if opts.OptLevel != "" && llvmPaths.Opt != "" {
		optLevel := opts.OptLevel
		if optLevel != "1" && optLevel != "2" && optLevel != "3" {
			optLevel = "2" // clamp s/z/default → O2
		}
		optIRFile := base + ".opt.ll"
		// opt 22+ uses --O<N> (new pass manager's default<O<N>> alias).
		optArgs := []string{"--O" + optLevel, irFile, "-o", optIRFile}
		optCmd := exec.Command(llvmPaths.Opt, optArgs...)
		if out, err := optCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("opt failed: %w\n%s", err, out)
		}
		irFile = optIRFile // feed optimized IR to llc
	}

	// llc: .ll → .o with optional optimization level
	objFile := base + ".o"

	// v4.5.0 Phase C: incremental cache. If a cached .o exists for this
	// (IR content + opts), skip llc entirely and link the cached object.
	// The key is the final IR's hash (post-pass) so any codegen change
	// invalidates. Best-effort: cache failure is non-fatal.
	cacheKey := irCacheKey(ir, opts)
	cachedHit := false
	if store := defaultLLVMCache(); store != nil {
		if cached := store.Get(cacheKey); cached != "" {
			if copyFile(cached, objFile) == nil {
				cachedHit = true
			}
		}
	}

	if !cachedHit {
		llcArgs := []string{"-filetype=obj"}
		// v5.4.0: force -O0 when no explicit level — LLVM 22 llc defaults to
		// -O2 (the full optimization pipeline), which mis-optimizes the vtable
		// load sequence (folding obj[0]=vtable-ptr + vtable[idx] into obj[idx*8],
		// corrupting indirect calls). -O0 keeps the IR's explicit load/GEP steps.
		optLevel := opts.OptLevel
		if optLevel == "" {
			optLevel = "0"
		}
		switch optLevel {
		case "0", "1", "2", "3":
			llcArgs = append(llcArgs, "-O="+optLevel)
		default:
			llcArgs = append(llcArgs, "-O=2")
		}
		// v6.2.0: cross-compilation — pin the target so llc honors it even if
		// the IR triple were lost; llc is a multi-target compiler.
		if opts.Target != "" {
			tOS, tArch := resolveTarget(opts.Target)
			triple, _ := tripleFor(tOS, tArch)
			llcArgs = append(llcArgs, "-mtriple="+triple)
		}
		llcArgs = append(llcArgs, "-o", objFile, irFile)
		llcCmd := exec.Command(llvmPaths.LLC, llcArgs...)
		if out, err := llcCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("llc failed: %w\n%s", err, out)
		}
		// Populate the cache with the freshly-compiled object.
		if store := defaultLLVMCache(); store != nil {
			_ = store.Put(cacheKey, objFile)
		}
	}

	// Determine output binary name
	if outBin == "" {
		outBin = base
	}

	// v6.2.0: cross-compilation target (host default when opts.Target empty).
	// Drives the clang --target flag, the Windows linker driver, and which
	// system-library search path to use.
	targetOS, targetArch := resolveTarget(opts.Target)

	// clang: .o → native binary
	clangArgs := []string{"-o", outBin, objFile}
	// Cross-compiling: tell clang which platform to emit for. Windows needs
	// the lld linker driver + console subsystem (no CRT sysroot is provided;
	// clang finds the MSVC/WinSDK CRT from the environment).
	if opts.Target != "" {
		triple, _ := tripleFor(targetOS, targetArch)
		clangArgs = append(clangArgs, "--target="+triple)
		if targetOS == "windows" {
			clangArgs = append(clangArgs, "-fuse-ld=lld", "-Wl,/subsystem:console")
		}
	}
	// Linux: the IR accesses string constants with absolute relocations
	// (R_X86_64_32S against .rodata), which the linker rejects under PIE
	// ("can not be used when making a PIE object"). Link non-PIE. v6.2.0.
	if targetOS == "linux" {
		clangArgs = append(clangArgs, "-no-pie")
	}

	// Link system libraries by scanning the IR for stdlib module symbols.
	// v6.2.0: the -L/rpath handling is macOS-only (Homebrew); Linux uses the
	// system default path; Windows relies on clang's .lib search.
	if strings.Contains(ir, "@__kylix_crypto_") || strings.Contains(ir, "@SHA1") {
		clangArgs = append(clangArgs, "-lcrypto")
		if targetOS == "darwin" {
			appendHomebrewLib(&clangArgs, "openssl")
		}
	}
	if strings.Contains(ir, "@__kylix_db_") || strings.Contains(ir, "@sqlite3_") {
		clangArgs = append(clangArgs, "-lsqlite3")
		if targetOS == "darwin" {
			appendHomebrewLib(&clangArgs, "sqlite")
		}
	}
	if strings.Contains(ir, "@__kylix_httpclient_") || strings.Contains(ir, "@curl_easy_") {
		clangArgs = append(clangArgs, "-lcurl")
		if targetOS == "darwin" {
			appendHomebrewLib(&clangArgs, "curl")
		}
	}
	clangCmd := exec.Command(llvmPaths.Clang, clangArgs...)
	if out, err := clangCmd.CombinedOutput(); err != nil {
		// v6.2.0: cross-compiling on a host that lacks the target's CRT/libc
		// (e.g. linking linux/amd64 from macOS) fails at link even though llc
		// produced a correct object file. Explain instead of a bare clang error.
		if opts.Target != "" && targetOS != runtime.GOOS {
			return nil, fmt.Errorf("clang link failed for target %s (%s produced; linking %s needs that platform's CRT/libc — build on %s or use CI): %w\n%s",
				opts.Target, objFile, opts.Target, targetOS, err, out)
		}
		// Retry with -v so the underlying linker error surfaces in the report
		// (clang otherwise reports only "linker command failed").
		if vout, verr := exec.Command(llvmPaths.Clang, append(clangArgs, "-v")...).CombinedOutput(); verr != nil {
			out = vout
		}
		return nil, fmt.Errorf("clang link failed: %w\n%s", err, out)
	}

	return &CompileResult{
		IRFile:  irFile,
		ObjFile: objFile,
		BinFile: outBin,
	}, nil
}
