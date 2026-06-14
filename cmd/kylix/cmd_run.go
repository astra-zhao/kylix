package main

import (
	"flag"
	"fmt"
	"kylix/pkg/compiler"
	"kylix/pkg/project"
	"os"
	"os/exec"
	"path/filepath"
)

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	keepGo := fs.Bool("keep", false, "Keep generated .go file")
	verbose := fs.Bool("v", false, "Verbose output")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix run [options] [file.klx]

Compile and run the current project or a single Kylix file.

OPTIONS:
`)
		fs.PrintDefaults()
		fmt.Println()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Single file mode
	if fs.NArg() > 0 {
		file := fs.Arg(0)
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
