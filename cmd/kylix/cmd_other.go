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
)

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
