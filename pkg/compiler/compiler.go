package compiler

import (
	"fmt"
	"io/ioutil"
	"kylix/generator"
	"kylix/lexer"
	"kylix/parser"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Diagnostic represents a single error or warning from compilation
type Diagnostic struct {
	File    string
	Line    int
	Column  int
	Level   string // "error" or "warning"
	Message string
	Source  string // the source line where the issue occurred
}

// Result holds the outcome of a compilation
type Result struct {
	Success     bool
	GoCode      string
	OutputFile  string
	Diagnostics []Diagnostic
}

// Options controls compilation behavior
type Options struct {
	OutputFile  string
	Verbose     bool
	KeepGoFile  bool // don't delete the intermediate .go file after running
	WorkingDir  string
}

// CompileFile compiles a single .klx file to Go
func CompileFile(sourceFile string, opts Options) (*Result, error) {
	result := &Result{}

	// Read source
	source, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %v", sourceFile, err)
	}

	// Lex + Parse
	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	// Collect parser errors as diagnostics
	for _, errMsg := range p.Errors() {
		d := Diagnostic{
			File:    sourceFile,
			Level:   "error",
			Message: errMsg,
		}
		// Try to extract line/column from error message
		parseLocation(&d, errMsg)
		result.Diagnostics = append(result.Diagnostics, d)
	}

	if len(result.Diagnostics) > 0 {
		result.Success = false
		return result, nil
	}

	// Generate Go code
	gen := generator.New()
	goCode := gen.Generate(program)
	result.GoCode = goCode

	// Determine output path
	outputFile := opts.OutputFile
	if outputFile == "" {
		base := filepath.Base(sourceFile)
		name := base[:len(base)-len(filepath.Ext(base))]
		outputFile = name + ".go"
		if opts.WorkingDir != "" {
			outputFile = filepath.Join(opts.WorkingDir, outputFile)
		}
	}

	// Write Go file
	if err := ioutil.WriteFile(outputFile, []byte(goCode), 0644); err != nil {
		return nil, fmt.Errorf("cannot write %s: %v", outputFile, err)
	}
	result.OutputFile = outputFile
	result.Success = true

	return result, nil
}

// CheckFile only parses the file and reports diagnostics, no code generation
func CheckFile(sourceFile string) (*Result, error) {
	result := &Result{}

	source, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %v", sourceFile, err)
	}

	lines := strings.Split(string(source), "\n")

	l := lexer.New(string(source))
	p := parser.New(l)
	_ = p.ParseProgram()

	for _, errMsg := range p.Errors() {
		d := Diagnostic{
			File:    sourceFile,
			Level:   "error",
			Message: errMsg,
		}
		parseLocation(&d, errMsg)
		// Attach source line for context
		if d.Line > 0 && d.Line <= len(lines) {
			d.Source = lines[d.Line-1]
		}
		result.Diagnostics = append(result.Diagnostics, d)
	}

	result.Success = len(result.Diagnostics) == 0
	return result, nil
}

// RunFile compiles and immediately runs the generated Go code
func RunFile(sourceFile string, opts Options) (*Result, error) {
	result, err := CompileFile(sourceFile, opts)
	if err != nil {
		return result, err
	}
	if !result.Success {
		return result, nil
	}

	// Run with `go run`
	cmd := exec.Command("go", "run", result.OutputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	runErr := cmd.Run()

	// Clean up intermediate .go file unless asked to keep
	if !opts.KeepGoFile && result.OutputFile != "" {
		os.Remove(result.OutputFile)
	}

	if runErr != nil {
		return result, fmt.Errorf("runtime error: %v", runErr)
	}

	return result, nil
}

// parseLocation tries to extract line and column from an error message
// like "expected next token to be X, got Y instead (line 5, column 14)"
func parseLocation(d *Diagnostic, msg string) {
	var line, col int
	// Look for "(line N, column M)" pattern
	if idx := strings.Index(msg, "(line "); idx >= 0 {
		_, err := fmt.Sscanf(msg[idx:], "(line %d, column %d)", &line, &col)
		if err == nil {
			d.Line = line
			d.Column = col
		}
	}
}
