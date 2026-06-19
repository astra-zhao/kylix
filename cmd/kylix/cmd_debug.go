package main

import (
	"flag"
	"fmt"
	"kylix/pkg/compiler"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func cmdDebug(args []string) {
	fs := flag.NewFlagSet("debug", flag.ExitOnError)
	headless := fs.Bool("headless", false, "Run dlv in headless mode (for IDE attach)")
	port := fs.String("port", "2345", "Port for headless mode")
	keepGo := fs.Bool("keep", false, "Keep generated .go file after exit")
	fs.Usage = func() {
		fmt.Printf(`USAGE: kylix debug [options] <file.klx>

Compile a Kylix file and launch the Delve debugger (dlv).
Generated Go code includes //line directives so breakpoints set on
the .klx source map back to the right line during stepping.

REQUIREMENTS:
  - dlv must be installed (https://github.com/go-delve/delve)

EXAMPLES:
  kylix debug main.klx                  # interactive REPL debugger
  kylix debug --headless --port=2345 main.klx
                                        # listen for IDE attach (VS Code, etc.)

DEBUGGER COMMANDS (interactive mode):
  break <file.klx>:<line>   set breakpoint
  continue / c              run until next breakpoint
  next / n                  step over
  step / s                  step into
  print <var> / p <var>     evaluate expression
  vars                      show local variables
  exit                      quit debugger

OPTIONS:
`)
		fs.PrintDefaults()
		fmt.Println()
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Error: exactly one .klx file required")
		fs.Usage()
		os.Exit(1)
	}
	klxFile := fs.Arg(0)
	if !strings.HasSuffix(klxFile, ".klx") {
		fmt.Fprintf(os.Stderr, "Error: %s is not a .klx file\n", klxFile)
		os.Exit(1)
	}

	// Verify dlv is installed
	if _, err := exec.LookPath("dlv"); err != nil {
		fmt.Fprintln(os.Stderr, "Error: 'dlv' not found in PATH")
		fmt.Fprintln(os.Stderr, "Install with:")
		fmt.Fprintln(os.Stderr, "  go install github.com/go-delve/delve/cmd/dlv@latest")
		os.Exit(1)
	}

	// Compile to Go (preserve //line directives so breakpoints map back to .klx)
	tmpDir, err := os.MkdirTemp("", "kylix-debug-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !*keepGo {
		defer os.RemoveAll(tmpDir)
	}

	goFile := filepath.Join(tmpDir, "main.go")
	opts := compiler.Options{
		OutputFile: goFile,
		Verbose:    false,
	}
	result, err := compiler.CompileFile(klxFile, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compile error: %v\n", err)
		os.Exit(1)
	}
	if !result.Success {
		printDiagnostics(result.Diagnostics)
		os.Exit(1)
	}

	// Build a debug binary (-gcflags="all=-N -l" disables optimization & inlining)
	binary := filepath.Join(tmpDir, "kylix-debug-bin")
	buildCmd := exec.Command("go", "build",
		"-gcflags=all=-N -l",
		"-o", binary,
		goFile,
	)
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Build error: %v\n", err)
		os.Exit(1)
	}

	// Construct dlv invocation
	var dlvArgs []string
	if *headless {
		dlvArgs = []string{
			"exec", binary,
			"--headless",
			"--listen=:" + *port,
			"--api-version=2",
			"--accept-multiclient",
		}
		fmt.Printf("Launching dlv in headless mode on port %s...\n", *port)
		fmt.Printf("Connect with: dlv connect localhost:%s\n", *port)
	} else {
		dlvArgs = []string{"exec", binary}
		fmt.Printf("Launching dlv for %s\n", klxFile)
		fmt.Printf("Set a breakpoint with: break %s:<line>\n", klxFile)
	}

	dlv := exec.Command("dlv", dlvArgs...)
	dlv.Stdin = os.Stdin
	dlv.Stdout = os.Stdout
	dlv.Stderr = os.Stderr
	if *keepGo {
		fmt.Printf("(generated Go file kept at %s)\n", goFile)
	}
	if err := dlv.Run(); err != nil {
		// dlv exiting with non-zero is normal (user issued 'exit' etc.)
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "dlv error: %v\n", err)
		os.Exit(1)
	}
}
