package main

import (
	"flag"
	"fmt"
	"kylix/pkg/compiler"
	"os"
)

const Version = "1.5.0"

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
	case "add":
		cmdAdd(os.Args[2:])
	case "install":
		cmdInstall(os.Args[2:])
	case "remove", "rm":
		cmdRemove(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("Kylix %s\n", Version)
	case "help", "-h", "--help":
		printUsage()
	default:
		if cmd[0] != '-' && len(cmd) > 4 && cmd[len(cmd)-4:] == ".klx" {
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
