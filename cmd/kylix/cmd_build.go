package main

import (
	"flag"
	"fmt"
	"kylix/pkg/compiler"
	"kylix/pkg/llvmgen"
	"kylix/pkg/project"
	"os"
	"os/exec"
	"path/filepath"
)

func cmdBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	output := fs.String("o", "", "Output file (for single file compilation)")
	verbose := fs.Bool("v", false, "Verbose output")
	target := fs.String("target", "", "Cross-compile target: os/arch (e.g. linux/amd64, windows/amd64, darwin/arm64)")
	wasm := fs.Bool("wasm", false, "Compile to WebAssembly (.wasm) — uses GOOS=js GOARCH=wasm (browser)")
	wasi := fs.Bool("wasi", false, "Compile to WASI WebAssembly (.wasm) — uses GOOS=wasip1 GOARCH=wasm (server-side)")
	tinygo := fs.Bool("tinygo", false, "Use TinyGo for WASM/WASI build (smaller output, requires tinygo installed)")
	backend := fs.String("backend", "go", "Compiler backend: go (default) or llvm (experimental)")
	llvmOpt := fs.String("llvm-opt", "", "LLVM optimization level (0/1/2/3); only meaningful with --backend=llvm")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix build [options] [file.klx]

Build the current project or a single Kylix file.

WASM/WASI EXAMPLES:
  kylix build --wasm main.klx            # Browser WASM via Go (~3 MB)
  kylix build --wasm --tinygo main.klx   # Browser WASM via TinyGo (~30 KB)
  kylix build --wasi main.klx            # WASI (Wasmtime/Cloudflare Workers)
  kylix build --wasi --tinygo main.klx   # WASI via TinyGo (smaller)

LLVM BACKEND (EXPERIMENTAL):
  kylix build --backend=llvm main.klx    # Native binary via LLVM IR

OPTIONS:
`)
		fs.PrintDefaults()
		fmt.Println()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Mutual exclusion checks
	if *wasm && *wasi {
		fmt.Fprintln(os.Stderr, "Error: --wasm and --wasi are mutually exclusive")
		os.Exit(1)
	}
	if (*wasm || *wasi) && *target != "" {
		fmt.Fprintln(os.Stderr, "Error: --wasm/--wasi and --target are mutually exclusive")
		os.Exit(1)
	}
	if *tinygo && !*wasm && !*wasi {
		fmt.Fprintln(os.Stderr, "Error: --tinygo requires --wasm or --wasi")
		os.Exit(1)
	}

	// Parse --target into GOOS/GOARCH
	var targetGOOS, targetGOARCH string
	if *target != "" {
		parts := splitTarget(*target)
		if parts == nil {
			fmt.Fprintf(os.Stderr, "Error: invalid --target %q, expected os/arch (e.g. linux/amd64)\n", *target)
			os.Exit(1)
		}
		targetGOOS, targetGOARCH = parts[0], parts[1]
	}

	// Single file or multi-file mode
	if fs.NArg() > 0 {
		wd, _ := os.Getwd()
		opts := compiler.Options{
			OutputFile:        *output,
			Verbose:           *verbose,
			CacheDir:          wd,
			PackageSearchDirs: packageDirsFromWd(wd),
		}

		if fs.NArg() == 1 {
			file := fs.Arg(0)

			// LLVM backend shortcut — bypass Go codegen entirely
			if *backend == "llvm" {
				if err := buildWithLLVM(file, *output, *llvmOpt); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			}

			result, err := compiler.CompileFile(file, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			printDiagnostics(result.Diagnostics)
			if !result.Success {
				os.Exit(1)
			}
			if *wasm || *wasi {
				binOut := *output
				if binOut == "" {
					binOut = stripExt(file) + ".wasm"
				}
				if err := goBuildWasmTarget(result.OutputFile, binOut, *wasi, *tinygo); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				flavor := "Go"
				if *tinygo {
					flavor = "TinyGo"
				}
				target := "wasm"
				if *wasi {
					target = "wasi"
				}
				fmt.Printf("✓ Built %s → %s [%s via %s]\n", file, binOut, target, flavor)
			} else if targetGOOS != "" {
				binOut := *output
				if binOut == "" {
					binOut = stripExt(file)
					if targetGOOS == "windows" {
						binOut += ".exe"
					}
				}
				if err := goBuild(result.OutputFile, binOut, targetGOOS, targetGOARCH); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("✓ Built %s → %s [%s/%s]\n", file, binOut, targetGOOS, targetGOARCH)
			} else {
				fmt.Printf("✓ Compiled %s → %s\n", file, result.OutputFile)
			}
			return
		}

		// Multi-file mode
		files := fs.Args()
		result, err := compiler.CompileProject(files, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printDiagnostics(result.Diagnostics)
		if !result.Success {
			os.Exit(1)
		}
		fmt.Printf("✓ Compiled %d files → %s\n", len(files), result.OutputFile)
		return
	}

	// Project mode
	cfg, err := project.Find(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if cfg == nil {
		fmt.Fprintf(os.Stderr, "Error: no kylix.toml found in current directory or parents\n")
		fmt.Fprintf(os.Stderr, "Run 'kylix new <name>' to create a project\n")
		os.Exit(1)
	}

	mainFile := cfg.MainFilePath()
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: main file not found: %s\n", mainFile)
		os.Exit(1)
	}

	// Ensure output directory exists
	outDir := cfg.OutputDir()
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Compile
	outFile := filepath.Join(outDir, cfg.Name+".go")
	opts := compiler.Options{
		OutputFile:        outFile,
		Verbose:           *verbose,
		CacheDir:          cfg.ProjectDir(),
		PackageSearchDirs: packageDirsFromWd(cfg.ProjectDir()),
	}

	// If the project has multiple .klx files, use CompileProject (incremental cache).
	// Otherwise fall back to CompileFile for single-file projects.
	allFiles, _ := cfg.FindAllKlxFiles()
	var result *compiler.Result
	if len(allFiles) > 1 {
		result, err = compiler.CompileProject(allFiles, opts)
	} else {
		result, err = compiler.CompileFile(mainFile, opts)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	printDiagnostics(result.Diagnostics)
	if !result.Success {
		os.Exit(1)
	}

	// Generate go.mod if needed
	goModPath := filepath.Join(outDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		goMod := fmt.Sprintf("module %s\n\ngo 1.21\n", cfg.GoMod)
		if err := os.WriteFile(goModPath, []byte(goMod), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not create go.mod: %v\n", err)
		}
	}

	if *wasm || *wasi {
		binOut := filepath.Join(outDir, cfg.Name+".wasm")
		if err := goBuildWasmTarget(outFile, binOut, *wasi, *tinygo); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		flavor := "Go"
		if *tinygo {
			flavor = "TinyGo"
		}
		tgt := "wasm"
		if *wasi {
			tgt = "wasi"
		}
		fmt.Printf("✓ Built %s → %s [%s via %s]\n", cfg.Name, binOut, tgt, flavor)
	} else if targetGOOS != "" {
		binOut := filepath.Join(outDir, cfg.Name)
		if targetGOOS == "windows" {
			binOut += ".exe"
		}
		if err := goBuild(outFile, binOut, targetGOOS, targetGOARCH); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Built %s → %s [%s/%s]\n", cfg.Name, binOut, targetGOOS, targetGOARCH)
	} else {
		fmt.Printf("✓ Built %s → %s\n", cfg.Name, outFile)
	}
}

// splitTarget parses "os/arch" into [os, arch], returns nil on bad input.
func splitTarget(t string) []string {
	for i, c := range t {
		if c == '/' {
			return []string{t[:i], t[i+1:]}
		}
	}
	return nil
}

// goBuild compiles a .go file to a native binary with optional cross-compilation.
func goBuild(goFile, outBin, goos, goarch string) error {
	cmd := exec.Command("go", "build", "-o", outBin, goFile)
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// goBuildWasmTarget compiles to WebAssembly for browser (GOOS=js) or WASI (GOOS=wasip1).
// useTinyGo selects TinyGo instead of the standard Go toolchain.
func goBuildWasmTarget(goFile, outBin string, useWasi, useTinyGo bool) error {
	if useTinyGo {
		if _, err := exec.LookPath("tinygo"); err != nil {
			return fmt.Errorf("tinygo not found in PATH; install from https://tinygo.org/getting-started/install/")
		}
		tinygoTarget := "wasm"
		if useWasi {
			tinygoTarget = "wasi"
		}
		cmd := exec.Command("tinygo", "build", "-o", outBin, "-target="+tinygoTarget, goFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	// Standard Go toolchain
	goos := "js"
	if useWasi {
		goos = "wasip1"
	}
	cmd := exec.Command("go", "build", "-o", outBin, goFile)
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// goBuildWasm is kept for backward compatibility.
func goBuildWasm(goFile, outBin string, useTinyGo bool) error {
	return goBuildWasmTarget(goFile, outBin, false, useTinyGo)
}

// stripExt removes the file extension.
func stripExt(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[:i]
		}
		if name[i] == '/' || name[i] == os.PathSeparator {
			break
		}
	}
	return name
}

// buildWithLLVM compiles a Kylix file to native binary via LLVM IR.
func buildWithLLVM(srcFile, outBin, optLevel string) error {
	llvmPaths, err := llvmgen.FindLLVM()
	if err != nil {
		return fmt.Errorf("LLVM toolchain not found: %w\nHint: brew install llvm (macOS) or apt install llvm clang (Linux)", err)
	}

	result, err := llvmgen.CompileToNativeOpts(srcFile, outBin, llvmPaths, llvmgen.CompileOpts{
		OptLevel: optLevel,
	})
	if err != nil {
		return err
	}

	optInfo := ""
	if optLevel != "" {
		optInfo = " -O" + optLevel
	}
	fmt.Printf("✓ Built %s → %s [llvm%s]\n", srcFile, result.BinFile, optInfo)
	fmt.Printf("  IR:  %s\n", result.IRFile)
	fmt.Printf("  Obj: %s\n", result.ObjFile)
	return nil
}
