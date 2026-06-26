package compiler

import (
	"fmt"

	"kylix/ast"
	"kylix/generator"
	"kylix/lexer"
	"kylix/parser"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Diagnostic represents a single error or warning from compilation
type Diagnostic struct {
	File    string
	Line    int
	Column  int
	Level   string // "error" or "warning"
	Code    string // "KLX001", "KLX002", etc.
	Message string
	Source  string // the source line where the issue occurred
	Hint    string // optional fix suggestion
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
	OutputFile        string
	Verbose           bool
	KeepGoFile        bool // don't delete the intermediate .go file after running
	WorkingDir        string
	CacheDir          string   // directory for incremental build cache; "" disables caching
	PackageSearchDirs []string // extra directories containing .klx unit files (from packages/)
}

// CompileFile compiles a single .klx file to Go
func CompileFile(sourceFile string, opts Options) (*Result, error) {
	result := &Result{}

	// Read source
	source, err := os.ReadFile(sourceFile)
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
			Code:    ErrParseGeneric,
			Message: errMsg,
		}
		parseLocation(&d, errMsg)
		result.Diagnostics = append(result.Diagnostics, d)
	}

	// Parser errors are fatal — the AST is incomplete so semantic checks
	// would produce noisy false positives. Stop here.
	if len(result.Diagnostics) > 0 {
		result.Success = false
		return result, nil
	}

	// Semantic checks — run ALL of them before deciding success/failure so
	// the user sees as many errors as possible in one compilation.

	// Interface implementation validation
	result.Diagnostics = append(result.Diagnostics, checkInterfaces(program, sourceFile)...)
	result.Diagnostics = append(result.Diagnostics, checkBootAnnotations(program, sourceFile)...)
	result.Diagnostics = append(result.Diagnostics, checkValidationAnnotations(program, sourceFile)...)
	result.Diagnostics = append(result.Diagnostics, checkSecurityAnnotations(program, sourceFile)...)
	result.Diagnostics = append(result.Diagnostics, checkORMAnnotations(program, sourceFile)...)

	// Type checker: undeclared vars, arity, obvious type mismatches
	for _, td := range TypeCheck(program, sourceFile) {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			File:    td.File,
			Line:    td.Line,
			Column:  td.Column,
			Level:   "error",
			Code:    td.Code,
			Message: td.Message,
			Hint:    td.Hint,
		})
	}

	// If any semantic errors were found, stop before code generation.
	if len(result.Diagnostics) > 0 {
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
	if err := os.WriteFile(outputFile, []byte(goCode), 0644); err != nil {
		return nil, fmt.Errorf("cannot write %s: %v", outputFile, err)
	}
	result.OutputFile = outputFile
	result.Success = true

	return result, nil
}

// CheckFile only parses the file and reports diagnostics, no code generation
func CheckFile(sourceFile string) (*Result, error) {
	result := &Result{}

	source, err := os.ReadFile(sourceFile)
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

// CheckProject performs a full project-level check across multiple .klx files.
// It runs:
//  1. Per-file syntax check
//  2. Cross-file dependency resolution (uses clauses)
//  3. Interface implementation validation
//  4. Type checking with cross-unit symbol visibility
//
// Errors are collected from ALL files and returned together. Code is NOT generated.
func CheckProject(files []string) (*Result, error) {
	result := &Result{}

	if len(files) == 0 {
		return nil, fmt.Errorf("no source files provided")
	}

	type parsed struct {
		file    string
		program *ast.Program
		lines   []string
	}

	programs := make([]parsed, 0, len(files))
	fileMap := make(map[string]*ast.Program) // unitName → program	// Pass 1: parse all files and collect syntax diagnostics.
	for _, file := range files {
		source, err := os.ReadFile(file)
		if err != nil {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				File:    file,
				Level:   "error",
				Code:    ErrCannotRead,
				Message: fmt.Sprintf("cannot read: %v", err),
			})
			continue
		}
		lines := strings.Split(string(source), "\n")
		l := lexer.New(string(source))
		p := parser.New(l)
		program := p.ParseProgram()

		for _, errMsg := range p.Errors() {
			d := Diagnostic{File: file, Level: "error", Code: ErrParseGeneric, Message: errMsg}
			parseLocation(&d, errMsg)
			if d.Line > 0 && d.Line <= len(lines) {
				d.Source = lines[d.Line-1]
			}
			result.Diagnostics = append(result.Diagnostics, d)
		}

		// Even with parse errors keep the (partial) AST so later checks can run.
		programs = append(programs, parsed{file: file, program: program, lines: lines})

		name := program.UnitName
		if name == "" {
			name = program.Name
		}
		if name != "" {
			fileMap[name] = program
		}
	}

	// If parser found no usable programs, stop here.
	if len(programs) == 0 {
		result.Success = false
		return result, nil
	}

	// Pass 2: cross-file dependency resolution — verify each `uses X` resolves.
	for _, p := range programs {
		for _, used := range p.program.Uses {
			if _, found := fileMap[used]; !found {
				// Could be a stdlib unit (sysutil, datetime, etc.) — skip those.
				if isStdlibUnit(used) {
					continue
				}
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					File:    p.file,
					Level:   "error",
					Code:    ErrUndeclared,
					Message: fmt.Sprintf("uses clause references unknown unit '%s'", used),
					Hint:    fmt.Sprintf("ensure %s.klx is in the project, or it's a stdlib module", used),
				})
			}
		}
	}

	// Pass 3: per-file semantic checks (interfaces + types).
	// We pass a merged "visible" symbol set so cross-unit symbols don't trigger
	// false-positive "undeclared" errors.
	mergedSyms := make(map[string]bool)
	for _, p := range programs {
		for _, decl := range p.program.Declarations {
			switch d := decl.(type) {
			case *ast.FunctionDecl:
				mergedSyms[d.Name] = true
			case *ast.VarDecl:
				for _, n := range d.Names {
					mergedSyms[n] = true
				}
			case *ast.ConstDecl:
				mergedSyms[d.Name] = true
			case *ast.TypeDecl:
				mergedSyms[d.Name] = true
			case *ast.ClassDecl:
				mergedSyms[d.Name] = true
			case *ast.InterfaceDecl:
				mergedSyms[d.Name] = true
			}
		}
	}

	for _, p := range programs {
		// Interface implementation validation (single-file scope is fine here).
		result.Diagnostics = append(result.Diagnostics, checkInterfaces(p.program, p.file)...)

		// Type checker with cross-unit symbol context.
		for _, td := range typeCheckWithExternals(p.program, p.file, mergedSyms) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				File:    td.File,
				Line:    td.Line,
				Column:  td.Column,
				Level:   "error",
				Code:    td.Code,
				Message: td.Message,
				Hint:    td.Hint,
			})
		}
	}

	result.Success = len(result.Diagnostics) == 0
	return result, nil
}

// isStdlibUnit reports whether a uses-clause name refers to a known stdlib module
// rather than a project-local unit. Stdlib modules don't need .klx files in the project.
func isStdlibUnit(name string) bool {
	stdlib := map[string]bool{
		"web": true, "container": true, "config": true, "middleware": true,
		"validation": true, "orm": true, "template": true, "autoconfig": true,
		"sysutil": true, "jsonutil": true, "datetime": true, "regex": true,
		"strutil": true, "mathutil": true, // Phase 1 Kylix stdlib
	}
	return stdlib[name]
}

// typeCheckWithExternals runs the type checker but pre-seeds the symbol scope
// with names from other compilation units, preventing false-positive "undeclared"
// errors for cross-unit symbols. Strict mode for project-level checking.
func typeCheckWithExternals(program *ast.Program, sourceFile string, externals map[string]bool) []TypeDiagnostic {
	c := &checker{
		file:                sourceFile,
		strictFunctionCalls: true, // project-mode: report undeclared function calls
		funcs:               make(map[string]*ast.FunctionDecl),
		types:               make(map[string]string),
		aliases:             make(map[string]string),
		genericConstraints:  make(map[string]*GenericTypeInfo),
		interfaces:          make(map[string]map[string]*ast.FunctionDecl),
		classImpls:          make(map[string][]string),
		classParent:         make(map[string]string),
		classMethods:        make(map[string]map[string]*ast.FunctionDecl),
	}
	c.collectDeclarations(program)
	// Inject external symbols as "known" so undeclared checks pass for cross-unit refs.
	for name := range externals {
		if _, exists := c.types[name]; !exists {
			c.types[name] = "external"
		}
	}
	c.validateAliases(program, sourceFile)
	c.checkProgram(program)
	return c.diags
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
// When opts.CacheDir is non-empty, unchanged files reuse their cached Go body
// (incremental compilation).
// When opts.PackageSearchDirs is non-empty, .klx unit files found in those
// directories are automatically added to the compilation (packages/).
func CompileProject(files []string, opts Options) (*Result, error) {
	result := &Result{}

	if len(files) == 0 {
		return nil, fmt.Errorf("no source files provided")
	}

	// Append .klx unit files from package search directories.
	for _, dir := range opts.PackageSearchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".klx") {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
	}

	// Optional build cache.
	var cache *BuildCache
	if opts.CacheDir != "" {
		cache = NewBuildCache(opts.CacheDir)
	}

	programs := make([]*ast.Program, 0, len(files))
	fileMap := make(map[string]*ast.Program)
	// Track which files need fresh codegen vs cached body.
	cachedBodies := make(map[string]string) // absPath → cached Go body
	needsRegen := make([]bool, len(files))

	// Parallel parse: each file is lexed+parsed independently in a goroutine.
	type parseResult struct {
		index     int
		file      string
		absFile   string
		program   *ast.Program
		err       error
		parseErrs []string
		cached    bool
	}

	results := make([]parseResult, len(files))
	var wg sync.WaitGroup

	for i, file := range files {
		absFile, _ := filepath.Abs(file)

		// Check cache before parsing.
		isCached := false
		if cache != nil {
			if entry := cache.Load(absFile); entry != nil {
				cachedBodies[absFile] = entry.GoCode
				needsRegen[i] = false
				isCached = true
				if opts.Verbose {
					fmt.Printf("  cached: %s\n", file)
				}
			} else {
				needsRegen[i] = true
			}
		} else {
			needsRegen[i] = true
		}

		wg.Add(1)
		go func(idx int, f, abs string, cached bool) {
			defer wg.Done()
			pr := parseResult{index: idx, file: f, absFile: abs, cached: cached}

			source, err := os.ReadFile(f)
			if err != nil {
				pr.err = fmt.Errorf("cannot read %s: %v", f, err)
				results[idx] = pr
				return
			}

			l := lexer.New(string(source))
			p := parser.New(l)
			pr.program = p.ParseProgram()
			pr.parseErrs = p.Errors()
			results[idx] = pr
		}(i, file, absFile, isCached)
	}
	wg.Wait()

	// Collect results in original order.
	for _, pr := range results {
		if pr.err != nil {
			return nil, pr.err
		}
		for _, errMsg := range pr.parseErrs {
			d := Diagnostic{File: pr.file, Level: "error", Code: ErrParseGeneric, Message: errMsg}
			parseLocation(&d, errMsg)
			result.Diagnostics = append(result.Diagnostics, d)
		}
		if len(pr.parseErrs) > 0 {
			result.Success = false
			if cache != nil {
				cache.Invalidate(pr.absFile)
			}
			return result, nil
		}

		name := pr.program.UnitName
		if name == "" {
			name = pr.program.Name
		}
		if name != "" {
			fileMap[name] = pr.program
		}
		programs = append(programs, pr.program)
	}

	sorted, sortedFiles, err := topoSortWithFiles(programs, fileMap, files)
	if err != nil {
		return nil, fmt.Errorf("dependency error: %v", err)
	}

	// Semantic checks on all files — run all of them before deciding success.
	result.Diagnostics = append(result.Diagnostics, CheckBootAnnotations(sorted, sortedFiles)...)
	result.Diagnostics = append(result.Diagnostics, CheckValidationAnnotations(sorted, sortedFiles)...)
	result.Diagnostics = append(result.Diagnostics, CheckSecurityAnnotations(sorted, sortedFiles)...)
	result.Diagnostics = append(result.Diagnostics, CheckORMAnnotations(sorted, sortedFiles)...)
	for i, prog := range sorted {
		result.Diagnostics = append(result.Diagnostics, checkInterfaces(prog, sortedFiles[i])...)
		for _, td := range TypeCheck(prog, sortedFiles[i]) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				File:    td.File,
				Line:    td.Line,
				Column:  td.Column,
				Level:   "error",
				Code:    td.Code,
				Message: td.Message,
				Hint:    td.Hint,
			})
		}
	}
	if len(result.Diagnostics) > 0 {
		result.Success = false
		return result, nil
	}

	// Code generation — incremental where possible.
	gen := generator.New()

	// Global pre-scan must see all programs for correct cross-unit type refs.
	for _, prog := range sorted {
		gen.CollectClassTypes(prog)
		gen.ScanImports(prog)
		gen.ScanBootAnnotations(prog)
		gen.ScanValidationAnnotations(prog)
		gen.ScanORMAnnotations(prog)
		gen.ScanForException(prog)
	}
	// (exception types are emitted inside BuildOutput)

	var bodies []string
	for i, prog := range sorted {
		absFile, _ := filepath.Abs(sortedFiles[i])
		gen.SetSourceFile(sortedFiles[i])

		var body string
		if !needsRegen[absFile2index(files, sortedFiles[i])] {
			if cached, ok := cachedBodies[absFile]; ok {
				body = cached
				if opts.Verbose {
					fmt.Printf("  reuse:  %s\n", sortedFiles[i])
				}
			}
		}
		if body == "" {
			body = gen.GenerateBody(prog)
			if cache != nil {
				cache.Store(absFile, body)
			}
			if opts.Verbose {
				fmt.Printf("  compile: %s\n", sortedFiles[i])
			}
		}
		bodies = append(bodies, body)
	}

	goCode := gen.BuildOutput(bodies)
	result.GoCode = goCode

	outputFile := opts.OutputFile
	if outputFile == "" {
		outputFile = "main.go"
		if opts.WorkingDir != "" {
			outputFile = filepath.Join(opts.WorkingDir, outputFile)
		}
	}

	if err := os.WriteFile(outputFile, []byte(goCode), 0644); err != nil {
		return nil, fmt.Errorf("cannot write %s: %v", outputFile, err)
	}
	result.OutputFile = outputFile
	result.Success = true

	return result, nil
}

// absFile2index returns the index of file in files slice (-1 if not found).
func absFile2index(files []string, target string) int {
	for i, f := range files {
		if f == target {
			return i
		}
		if abs, _ := filepath.Abs(f); abs == target {
			return i
		}
	}
	return 0
}

func topoSortWithFiles(programs []*ast.Program, fileMap map[string]*ast.Program, files []string) ([]*ast.Program, []string, error) {
	// Build reverse map: program pointer → original file path
	progFile := make(map[*ast.Program]string)
	for i, prog := range programs {
		if i < len(files) {
			progFile[prog] = files[i]
		}
	}

	visited := make(map[*ast.Program]bool)
	inStack := make(map[*ast.Program]bool)
	var sorted []*ast.Program
	var sortedFiles []string

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
		sortedFiles = append(sortedFiles, progFile[prog])
		return nil
	}

	for _, prog := range programs {
		if err := visit(prog); err != nil {
			return nil, nil, err
		}
	}

	return sorted, sortedFiles, nil
}

// CheckInterfaces is the exported version for use by LSP and other packages.
func CheckInterfaces(program *ast.Program, sourceFile string) []Diagnostic {
	return checkInterfaces(program, sourceFile)
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
						Code:    ErrMissingMethod,
						Message: fmt.Sprintf("class %q implements %q but is missing method %q", classDecl.Name, ifaceName, method),
						Hint:    fmt.Sprintf("add 'procedure/function %s' to class %s", method, classDecl.Name),
					})
				}
			}
		}
	}

	return diags
}
