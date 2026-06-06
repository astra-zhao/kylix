package main

import (
	"flag"
	"fmt"
	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/compiler"
	"kylix/pkg/formatter"
	"kylix/pkg/lsp"
	"kylix/pkg/project"
	"kylix/pkg/repl"
	"os"
	"os/exec"
	"path/filepath"
)

const Version = "1.1.1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "build":
		cmdBuild(os.Args[2:])
	case "run":
		cmdRun(os.Args[2:])
	case "check":
		cmdCheck(os.Args[2:])
	case "new":
		cmdNew(os.Args[2:])
	case "fmt":
		cmdFmt(os.Args[2:])
	case "repl":
		cmdRepl(os.Args[2:])
	case "lsp":
		cmdLsp(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("Kylix %s\n", Version)
	case "help", "-h", "--help":
		printUsage()
	default:
		// For backward compatibility, treat as a file to compile
		if cmd[0] != '-' && (len(cmd) > 4 && cmd[len(cmd)-4:] == ".klx") {
			legacyCompile(os.Args[1:])
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
			printUsage()
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Printf(`Kylix %s - Modern Pascal Compiler

USAGE:
    kylix <command> [options] [arguments]

COMMANDS:
    new     Create a new Kylix project
    build   Compile the project or a single file
    run     Compile and run the project or file
    check   Check syntax without generating code
    fmt     Format Kylix source files
    repl    Start interactive REPL
    lsp     Start Language Server Protocol server (for editors)
    version Show version information
    help    Show this help message

EXAMPLES:
    kylix new myapp          Create a new project
    kylix build              Build current project
    kylix build main.klx     Build a single file
    kylix run                Run current project
    kylix run main.klx       Run a single file
    kylix check              Check all project files
    kylix fmt                Format all project files

For more information on a command, run:
    kylix <command> --help
`, Version)
}

func cmdBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	output := fs.String("o", "", "Output file (for single file compilation)")
	verbose := fs.Bool("v", false, "Verbose output")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix build [options] [file.klx]

Build the current project or a single Kylix file.

OPTIONS:
`)
		fs.PrintDefaults()
		fmt.Println()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Single file or multi-file mode
	if fs.NArg() > 0 {
		opts := compiler.Options{
			OutputFile: *output,
			Verbose:    *verbose,
		}

		if fs.NArg() == 1 {
			file := fs.Arg(0)
			result, err := compiler.CompileFile(file, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			printDiagnostics(result.Diagnostics)
			if !result.Success {
				os.Exit(1)
			}
			fmt.Printf("✓ Compiled %s → %s\n", file, result.OutputFile)
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
		OutputFile: outFile,
		Verbose:    *verbose,
	}

	result, err := compiler.CompileFile(mainFile, opts)
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

	fmt.Printf("✓ Built %s → %s\n", cfg.Name, outFile)
}

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

func cmdCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix check [files...]

Check Kylix files for syntax errors without generating code.
If no files are specified, checks all .klx files in the project.
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	var files []string

	if fs.NArg() > 0 {
		files = fs.Args()
	} else {
		cfg, err := project.Find(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if cfg == nil {
			fmt.Fprintf(os.Stderr, "Error: no kylix.toml found\n")
			os.Exit(1)
		}
		files, err = cfg.FindAllKlxFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
			os.Exit(1)
		}
	}

	hasErrors := false
	for _, file := range files {
		result, err := compiler.CheckFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			hasErrors = true
			continue
		}
		if len(result.Diagnostics) > 0 {
			hasErrors = true
			printDiagnostics(result.Diagnostics)
		} else {
			fmt.Printf("✓ %s\n", file)
		}
	}

	if hasErrors {
		os.Exit(1)
	}
	fmt.Printf("\nAll files OK\n")
}

func cmdNew(args []string) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix new <project-name>

Create a new Kylix project with a template structure.

The project will include:
  - kylix.toml (project configuration)
  - main.klx (entry point)
  - build/ (output directory)
  - .gitignore
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: project name required\n")
		fs.Usage()
		os.Exit(1)
	}

	name := fs.Arg(0)

	_, err := project.Init(name, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created project '%s' in %s/\n", name, name)
	fmt.Printf("  Run 'cd %s && kylix run' to get started\n", name)
}

func printDiagnostics(diags []compiler.Diagnostic) {
	for _, d := range diags {
		loc := ""
		if d.Line > 0 {
			loc = fmt.Sprintf(":%d", d.Line)
			if d.Column > 0 {
				loc += fmt.Sprintf(":%d", d.Column)
			}
		}

		symbol := "✗"
		if d.Level == "warning" {
			symbol = "⚠"
		}

		fmt.Fprintf(os.Stderr, "%s %s%s: %s\n", symbol, d.File, loc, d.Message)

		if d.Source != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", d.Source)
			if d.Column > 0 {
				fmt.Fprintf(os.Stderr, "  %s^\n", repeatSpace(d.Column-1))
			}
		}
	}
}

func repeatSpace(n int) string {
	s := make([]byte, n)
	for i := range s {
		s[i] = ' '
	}
	return string(s)
}

func cmdRepl(args []string) {
	fmt.Println("Starting Kylix REPL...")
	if err := repl.Start(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "REPL error: %v\n", err)
		os.Exit(1)
	}
}

func cmdLsp(args []string) {
	server := lsp.New(os.Stdin, os.Stdout)
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "LSP error: %v\n", err)
		os.Exit(1)
	}
}

func cmdFmt(args []string) {
	fs := flag.NewFlagSet("fmt", flag.ExitOnError)
	overwrite := fs.Bool("w", false, "Write formatted output back to source files")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix fmt [options] [files...]

Format Kylix source files. If no files are specified, formats all .klx files in the project.

Options:
  -w    Write formatted output back to source files (default: print to stdout)

Examples:
  kylix fmt              # Preview formatted output
  kylix fmt -w           # Format and save all project files
  kylix fmt -w file.klx  # Format specific file
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	var files []string

	if fs.NArg() > 0 {
		files = fs.Args()
	} else {
		cfg, err := project.Find(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if cfg == nil {
			fmt.Fprintf(os.Stderr, "Error: no kylix.toml found\n")
			os.Exit(1)
		}
		files, err = cfg.FindAllKlxFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
			os.Exit(1)
		}
	}

	formattedCount := 0
	for _, file := range files {
		// Parse the file
		result, err := compiler.CheckFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		if len(result.Diagnostics) > 0 {
			fmt.Fprintf(os.Stderr, "⚠ %s has errors, skipping\n", file)
			printDiagnostics(result.Diagnostics)
			continue
		}

		// Parse to AST
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			continue
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			fmt.Fprintf(os.Stderr, "⚠ %s has parse errors, skipping\n", file)
			continue
		}

		// Format
		formatterInst := formatter.New()
		formatted := formatterInst.Format(program)

		if *overwrite {
			// Write back to file
			err = os.WriteFile(file, []byte(formatted), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", file, err)
				continue
			}
			fmt.Fprintf(os.Stderr, "✓ %s\n", file)
			formattedCount++
		} else {
			// Print to stdout
			fmt.Println(formatted)
		}
	}

	if *overwrite && formattedCount > 0 {
		fmt.Fprintf(os.Stderr, "\nFormatted %d file(s)\n", formattedCount)
	}
}

// legacyCompile handles backward-compatible single-file compilation
func legacyCompile(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: no file specified\n")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("legacy", flag.ExitOnError)
	output := fs.String("o", "", "Output file")
	run := fs.Bool("run", false, "Run after compilation")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	file := fs.Arg(0)
	if *run {
		opts := compiler.Options{
			Verbose: false,
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
	} else {
		opts := compiler.Options{
			OutputFile: *output,
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
		fmt.Printf("✓ Compiled %s → %s\n", file, result.OutputFile)
	}
}
