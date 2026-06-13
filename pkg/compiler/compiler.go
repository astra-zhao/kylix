package compiler

import (
	"fmt"
	"io/ioutil"
	"kylix/ast"
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

	// Semantic check: interface implementation validation
	if diags := checkInterfaces(program, sourceFile); len(diags) > 0 {
		result.Diagnostics = append(result.Diagnostics, diags...)
		result.Success = false
		return result, nil
	}

	// Generate Go code
	gen := generator.New()
	gen.SetSourceFile(sourceFile)
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

// CompileProject compiles multiple .klx files as a single project.
func CompileProject(files []string, opts Options) (*Result, error) {
	result := &Result{}

	if len(files) == 0 {
		return nil, fmt.Errorf("no source files provided")
	}

	programs := make([]*ast.Program, 0, len(files))
	fileMap := make(map[string]*ast.Program)

	for _, file := range files {
		source, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %v", file, err)
		}

		l := lexer.New(string(source))
		p := parser.New(l)
		program := p.ParseProgram()

		for _, errMsg := range p.Errors() {
			d := Diagnostic{
				File:    file,
				Level:   "error",
				Message: errMsg,
			}
			parseLocation(&d, errMsg)
			result.Diagnostics = append(result.Diagnostics, d)
		}

		if len(p.Errors()) > 0 {
			result.Success = false
			return result, nil
		}

		name := program.UnitName
		if name == "" {
			name = program.Name
		}
		if name != "" {
			fileMap[name] = program
		}
		programs = append(programs, program)
	}

	sorted, err := topoSort(programs, fileMap)
	if err != nil {
		return nil, fmt.Errorf("dependency error: %v", err)
	}

	// Semantic check: interface implementation validation across all files
	for i, prog := range sorted {
		if diags := checkInterfaces(prog, files[i]); len(diags) > 0 {
			result.Diagnostics = append(result.Diagnostics, diags...)
		}
	}
	if len(result.Diagnostics) > 0 {
		result.Success = false
		return result, nil
	}

	gen := generator.New()
	goCode := gen.GenerateMulti(sorted)
	result.GoCode = goCode

	outputFile := opts.OutputFile
	if outputFile == "" {
		outputFile = "main.go"
		if opts.WorkingDir != "" {
			outputFile = filepath.Join(opts.WorkingDir, outputFile)
		}
	}

	if err := ioutil.WriteFile(outputFile, []byte(goCode), 0644); err != nil {
		return nil, fmt.Errorf("cannot write %s: %v", outputFile, err)
	}
	result.OutputFile = outputFile
	result.Success = true

	return result, nil
}

func topoSort(programs []*ast.Program, fileMap map[string]*ast.Program) ([]*ast.Program, error) {
	visited := make(map[*ast.Program]bool)
	inStack := make(map[*ast.Program]bool)
	var sorted []*ast.Program

	var visit func(prog *ast.Program) error
	visit = func(prog *ast.Program) error {
		if visited[prog] {
			return nil
		}
		if inStack[prog] {
			return fmt.Errorf("circular dependency detected")
		}
		inStack[prog] = true

		for _, use := range prog.Uses {
			if dep, ok := fileMap[use]; ok {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}

		inStack[prog] = false
		visited[prog] = true
		sorted = append(sorted, prog)
		return nil
	}

	for _, prog := range programs {
		if err := visit(prog); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}

// checkInterfaces verifies that every class implementing an interface provides
// all required methods. Returns Kylix-layer diagnostics (not Go compiler errors).
func checkInterfaces(program *ast.Program, sourceFile string) []Diagnostic {
	// Build a map of interface name → required method names from this program.
	ifaceMap := make(map[string][]string)
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.InterfaceDecl:
			methods := make([]string, 0, len(d.Methods))
			for _, m := range d.Methods {
				methods = append(methods, m.Name)
			}
			ifaceMap[d.Name] = methods
		case *ast.TypeDecl:
			if iface, ok := d.Type.(*ast.InterfaceDecl); ok {
				methods := make([]string, 0, len(iface.Methods))
				for _, m := range iface.Methods {
					methods = append(methods, m.Name)
				}
				ifaceMap[d.Name] = methods
			}
		}
	}

	var diags []Diagnostic

	for _, decl := range program.Declarations {
		var classDecl *ast.ClassDecl
		switch d := decl.(type) {
		case *ast.ClassDecl:
			classDecl = d
		case *ast.TypeDecl:
			if cd, ok := d.Type.(*ast.ClassDecl); ok {
				classDecl = cd
			}
		}
		if classDecl == nil || len(classDecl.Interfaces) == 0 {
			continue
		}

		// Build the set of method names the class provides.
		implemented := make(map[string]bool)
		for _, m := range classDecl.Methods {
			// Strip "ClassName." prefix from top-level method definitions.
			name := m.Name
			if idx := strings.LastIndex(name, "."); idx >= 0 {
				name = name[idx+1:]
			}
			implemented[name] = true
		}

		for _, ifaceName := range classDecl.Interfaces {
			required, known := ifaceMap[ifaceName]
			if !known {
				// Interface defined in another unit — skip (can't validate cross-file yet).
				continue
			}
			for _, method := range required {
				if !implemented[method] {
					diags = append(diags, Diagnostic{
						File:    sourceFile,
						Line:    classDecl.Token.Line,
						Column:  classDecl.Token.Column,
						Level:   "error",
						Message: fmt.Sprintf("class %q implements %q but is missing method %q", classDecl.Name, ifaceName, method),
					})
				}
			}
		}
	}

	return diags
}
