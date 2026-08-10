package main

import (
	"flag"
	"fmt"
	"io"
	"kylix/pkg/compiler"
	"kylix/pkg/llvmgen"
	"kylix/pkg/project"
	"os"
	"os/exec"
	"path/filepath"
)

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	keepGo := fs.Bool("keep", false, "Keep generated .go file (Go backend) or native binary (LLVM backend)")
	verbose := fs.Bool("v", false, "Verbose output")
	backend := fs.String("backend", "auto", "Compiler backend: auto (default: Go if available, else LLVM), go, or llvm")
	llvmOpt := fs.String("llvm-opt", "", "LLVM optimization level (0/1/2/3); only meaningful with --backend=llvm")
	llvmDebug := fs.Bool("g", false, "Emit DWARF debug info (LLVM backend)")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix run [options] [file.klx]

Compile and run the current project or a single Kylix file.

The backend defaults to auto: Go when a Go toolchain is on PATH (existing
behaviour), otherwise the LLVM backend (native binary, needs llc + clang).

LLVM EXAMPLES:
  kylix run --backend=llvm hello.klx      # native binary, no Go needed
  kylix run --backend=llvm --llvm-opt=2 hello.klx

OPTIONS:
`)
		fs.PrintDefaults()
		fmt.Println()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	runBackend := resolveRunBackend(*backend)

	// Single file mode
	if fs.NArg() > 0 {
		file := fs.Arg(0)

		if runBackend == "llvm" {
			if err := runWithLLVM(file, *keepGo, *llvmOpt, *llvmDebug, *verbose); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		opts := compiler.Options{
			Verbose:    *verbose,
			KeepGoFile: *keepGo,
		}
		result, err := compiler.RunFile(file, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printDiagnostics(result.Diagnostics)
		if !result.Success {
			os.Exit(1)
		}
		return
	}

	// Project mode
	cfg, err := project.Find(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if cfg == nil {
		fmt.Fprintf(os.Stderr, "Error: no kylix.toml found\n")
		os.Exit(1)
	}

	if runBackend == "llvm" {
		if err := runProjectWithLLVM(cfg, *keepGo, *llvmOpt, *llvmDebug, *verbose); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	mainFile := cfg.MainFilePath()
	outDir := cfg.OutputDir()
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	outFile := filepath.Join(outDir, cfg.Name+".go")
	opts := compiler.Options{
		OutputFile: outFile,
		Verbose:    *verbose,
		KeepGoFile: *keepGo,
	}

	// Find all .klx files
	klxFiles, err := cfg.FindAllKlxFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding source files: %v\n", err)
		os.Exit(1)
	}

	var result *compiler.Result
	if len(klxFiles) > 1 {
		result, err = compiler.CompileProject(klxFiles, opts)
	} else {
		result, err = compiler.RunFile(mainFile, opts)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	printDiagnostics(result.Diagnostics)
	if !result.Success {
		os.Exit(1)
	}

	// For multi-file mode, run the generated Go file
	if len(klxFiles) > 1 && result.OutputFile != "" {
		goModPath := filepath.Join(outDir, "go.mod")
		if _, err := os.Stat(goModPath); os.IsNotExist(err) {
			goMod := fmt.Sprintf("module %s\n\ngo 1.21\n", cfg.GoMod)
			os.WriteFile(goModPath, []byte(goMod), 0644)
		}
		cmd := exec.Command("go", "run", filepath.Base(result.OutputFile))
		cmd.Dir = outDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if runErr := cmd.Run(); runErr != nil {
			fmt.Fprintf(os.Stderr, "Runtime error: %v\n", runErr)
			os.Exit(1)
		}
		if !*keepGo {
			os.Remove(result.OutputFile)
		}
	}
}

// resolveRunBackend maps the --backend flag to a concrete backend. "auto"
// prefers the Go backend when a Go toolchain is on PATH (existing behaviour),
// and falls back to the LLVM backend otherwise — so `kylix run` works on a
// machine with llc + clang but no Go toolchain (v6.1.0 KylixRT).
func resolveRunBackend(s string) string {
	switch s {
	case "llvm", "go":
		return s
	}
	if _, err := exec.LookPath("go"); err == nil {
		return "go"
	}
	return "llvm"
}

// runWithLLVM compiles a single Kylix file to a native binary and runs it.
// The .ll/.o/binary are produced in a throwaway temp dir and removed after the
// run unless --keep is given (in which case the binary is copied next to the
// source file, mirroring `kylix build --backend=llvm` output location).
func runWithLLVM(srcFile string, keep bool, optLevel string, debug, verbose bool) error {
	llvmPaths, err := llvmgen.FindLLVM()
	if err != nil {
		return fmt.Errorf("no Go toolchain and %w\nHint: install Go, or install LLVM (brew install llvm / apt install llvm clang)", err)
	}

	tmp, err := os.MkdirTemp("", "kylix-run-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	tmpSrc := filepath.Join(tmp, filepath.Base(srcFile))
	if err := copyFile(tmpSrc, srcFile); err != nil {
		return err
	}

	bin := filepath.Join(tmp, stripExt(filepath.Base(srcFile)))
	result, err := llvmgen.CompileToNativeOpts(tmpSrc, bin, llvmPaths, llvmgen.CompileOpts{
		OptLevel:  optLevel,
		DebugInfo: debug,
	})
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("  run llvm: %s\n", result.BinFile)
	}

	cmd := exec.Command(result.BinFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if runErr := cmd.Run(); runErr != nil {
		return fmt.Errorf("runtime error: %v", runErr)
	}

	if keep {
		out := filepath.Join(filepath.Dir(srcFile), stripExt(filepath.Base(srcFile)))
		return copyFile(out, result.BinFile)
	}
	return nil
}

// runProjectWithLLVM compiles all .klx files of a project to one native binary
// and runs it. Mirrors runWithLLVM's temp-dir + cleanup semantics.
func runProjectWithLLVM(cfg *project.Config, keep bool, optLevel string, debug, verbose bool) error {
	llvmPaths, err := llvmgen.FindLLVM()
	if err != nil {
		return fmt.Errorf("no Go toolchain and %w\nHint: install Go, or install LLVM (brew install llvm / apt install llvm clang)", err)
	}

	klxFiles, err := cfg.FindAllKlxFiles()
	if err != nil {
		return err
	}

	tmp, err := os.MkdirTemp("", "kylix-run-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	tmpFiles := make([]string, 0, len(klxFiles))
	for _, f := range klxFiles {
		dst := filepath.Join(tmp, filepath.Base(f))
		if err := copyFile(dst, f); err != nil {
			return err
		}
		tmpFiles = append(tmpFiles, dst)
	}

	bin := filepath.Join(tmp, cfg.Name)
	var result *llvmgen.CompileResult
	if len(tmpFiles) > 1 {
		result, err = llvmgen.CompileFilesToNative(tmpFiles, bin, llvmPaths, llvmgen.CompileOpts{OptLevel: optLevel, DebugInfo: debug})
	} else {
		result, err = llvmgen.CompileToNativeOpts(tmpFiles[0], bin, llvmPaths, llvmgen.CompileOpts{OptLevel: optLevel, DebugInfo: debug})
	}
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("  run llvm: %s\n", result.BinFile)
	}

	cmd := exec.Command(result.BinFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if runErr := cmd.Run(); runErr != nil {
		return fmt.Errorf("runtime error: %v", runErr)
	}

	if keep {
		mainFile := cfg.MainFilePath()
		out := filepath.Join(filepath.Dir(mainFile), cfg.Name)
		return copyFile(out, result.BinFile)
	}
	return nil
}

// copyFile copies src to dst, preserving nothing but content (used to stage
// sources into the temp build dir and copy the final binary back with --keep).
func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
