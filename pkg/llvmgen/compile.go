// compile.go — end-to-end LLVM compilation pipeline.
// Kylix source → AST → LLVM IR (.ll) → object (.o) → native binary
package llvmgen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		"/opt/homebrew/opt/llvm/bin", // Homebrew ARM
		"/usr/local/opt/llvm/bin",    // Homebrew x86
		"/usr/bin",                   // Linux system
		"/usr/local/bin",
	}

	find := func(name string) string {
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
		if opts.OptLevel != "" {
			// LLVM 22 doesn't accept letter levels (s/z) on llc directly; clamp them.
			switch opts.OptLevel {
			case "0", "1", "2", "3":
				llcArgs = append(llcArgs, "-O="+opts.OptLevel)
			default:
				llcArgs = append(llcArgs, "-O=2")
			}
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

	// clang: .o → native binary
	clangArgs := []string{"-o", outBin, objFile}
	// If the IR references OpenSSL libcrypto symbols (crypto stdlib), link
	// against libcrypto and add the Homebrew OpenSSL lib path (macOS). The
	// detection is done by scanning the IR for @__kylix_crypto_ defines,
	// which are only emitted when crypto functions are actually used.
	if strings.Contains(ir, "@__kylix_crypto_") {
		clangArgs = append(clangArgs, "-lcrypto")
		// Homebrew OpenSSL paths (macOS Intel + ARM). On Linux, libcrypto is
		// typically in the default search path, so no -L needed.
		for _, dir := range []string{
			"/opt/homebrew/opt/openssl/lib", // Homebrew ARM
			"/usr/local/opt/openssl/lib",    // Homebrew x86
		} {
			if _, err := os.Stat(dir); err == nil {
				clangArgs = append(clangArgs, "-L"+dir)
				// Also set rpath so the runtime linker finds libcrypto.dylib.
				clangArgs = append(clangArgs, "-Wl,-rpath,"+dir)
				break
			}
		}
	}
	// SQLite (db stdlib) — same pattern.
	if strings.Contains(ir, "@__kylix_db_") || strings.Contains(ir, "@sqlite3_") {
		clangArgs = append(clangArgs, "-lsqlite3")
		for _, dir := range []string{
			"/opt/homebrew/opt/sqlite/lib", // Homebrew ARM
			"/usr/local/opt/sqlite/lib",    // Homebrew x86
		} {
			if _, err := os.Stat(dir); err == nil {
				clangArgs = append(clangArgs, "-L"+dir)
				clangArgs = append(clangArgs, "-Wl,-rpath,"+dir)
				break
			}
		}
	}
	// libcurl (httpclient stdlib, v4.5.0) — same IR-symbol-scan pattern.
	if strings.Contains(ir, "@__kylix_httpclient_") || strings.Contains(ir, "@curl_easy_") {
		clangArgs = append(clangArgs, "-lcurl")
		for _, dir := range []string{
			"/opt/homebrew/opt/curl/lib", // Homebrew ARM
			"/usr/local/opt/curl/lib",    // Homebrew x86
		} {
			if _, err := os.Stat(dir); err == nil {
				clangArgs = append(clangArgs, "-L"+dir)
				clangArgs = append(clangArgs, "-Wl,-rpath,"+dir)
				break
			}
		}
	}
	clangCmd := exec.Command(llvmPaths.Clang, clangArgs...)
	if out, err := clangCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("clang link failed: %w\n%s", err, out)
	}

	return &CompileResult{
		IRFile:  irFile,
		ObjFile: objFile,
		BinFile: outBin,
	}, nil
}
