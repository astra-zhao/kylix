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
}

// FindLLVM looks for llc and clang in common install locations.
func FindLLVM() (*LLVMPaths, error) {
	searchDirs := []string{
		"/opt/homebrew/opt/llvm/bin", // Homebrew ARM
		"/usr/local/opt/llvm/bin",    // Homebrew x86
		"/usr/bin",                    // Linux system
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

	if llc == "" {
		return nil, fmt.Errorf("llc not found; install LLVM (brew install llvm or apt install llvm)")
	}
	if clang == "" {
		return nil, fmt.Errorf("clang not found; install clang (brew install llvm or apt install clang)")
	}

	return &LLVMPaths{LLC: llc, Clang: clang}, nil
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

	return CompileASTToNative(prog, srcFile, outBin, llvmPaths)
}

// CompileASTToNative compiles an already-parsed AST to a native binary.
func CompileASTToNative(prog *ast.Program, srcFile, outBin string, llvmPaths *LLVMPaths) (*CompileResult, error) {
	// Generate LLVM IR
	ir, err := Generate(prog)
	if err != nil {
		return nil, fmt.Errorf("LLVM IR generation: %w", err)
	}

	// Write .ll file
	base := strings.TrimSuffix(srcFile, filepath.Ext(srcFile))
	irFile := base + ".ll"
	if err := os.WriteFile(irFile, []byte(ir), 0644); err != nil {
		return nil, fmt.Errorf("write IR: %w", err)
	}

	// llc: .ll → .o
	objFile := base + ".o"
	llcCmd := exec.Command(llvmPaths.LLC,
		"-filetype=obj",
		"-o", objFile,
		irFile,
	)
	if out, err := llcCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("llc failed: %w\n%s", err, out)
	}

	// Determine output binary name
	if outBin == "" {
		outBin = base
	}

	// clang: .o → native binary
	clangCmd := exec.Command(llvmPaths.Clang,
		"-o", outBin,
		objFile,
	)
	if out, err := clangCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("clang link failed: %w\n%s", err, out)
	}

	return &CompileResult{
		IRFile:  irFile,
		ObjFile: objFile,
		BinFile: outBin,
	}, nil
}
