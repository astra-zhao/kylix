// multifile.go — multi-file (`uses X` / `unit X`) support for the LLVM backend.
//
// The LLVM backend's core codegen (Generate/emitProgram) only ever knew how
// to lower a single *ast.Program. When a Kylix program is split across a main
// program file and one or more `unit` files (referenced via `uses`), each
// unit's function/type/const declarations need to be visible to the main
// program's codegen pass — otherwise calls like `Cube(3)` lower to a `call`
// against a symbol that was never `define`d, and llc rejects the module with
// "use of undefined value".
//
// The Go backend solves this with generator.GenerateMulti, which walks every
// parsed *ast.Program and merges their declarations before emitting a single
// Go source file. This file provides the LLVM-backend equivalent: parse each
// file into its own *ast.Program, then flatten all of their Declarations into
// one synthetic *ast.Program (keyed off the non-unit "main" program's Name/
// Statements) before handing it to the existing single-program Generate path.
//
// Declaration order across files does not matter: emitProgram pre-registers
// every FunctionDecl's signature into g.funcSigs (and const values into
// g.constants) in a first pass before emitting any function bodies, so a
// function defined in a unit file that is call-ordered "before" its
// declaration in the merged list still resolves correctly (see
// codegen.go's emitProgram forward-reference pre-scan).
package llvmgen

import (
	"fmt"
	"os"

	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
)

// ParseFile reads and parses a single .klx file into an *ast.Program.
func ParseFile(path string) (*ast.Program, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	l := lexer.New(string(src))
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		return nil, fmt.Errorf("parse errors in %s: %v", path, errs)
	}
	return prog, nil
}

// MergePrograms flattens the declarations of every parsed file (main program
// + unit files, in any order) into a single *ast.Program suitable for the
// existing single-program Generate/emitProgram path.
//
// Exactly one program in the input must be a non-unit program (IsUnit ==
// false) — that program's Name and top-level Statements become the merged
// program's entry point. All other programs are treated as unit
// dependencies: only their Declarations are folded in (a unit has no
// Statements to run). Returns an error if zero or more than one non-unit
// program is found, since the LLVM backend (like the Go backend) only
// supports a single executable entry point per build.
func MergePrograms(programs []*ast.Program) (*ast.Program, error) {
	var main *ast.Program
	var decls []ast.Node

	for _, p := range programs {
		if !p.IsUnit {
			if main != nil {
				return nil, fmt.Errorf("multiple non-unit programs found (%q and %q); exactly one main program is required", main.Name, p.Name)
			}
			main = p
		}
	}
	if main == nil {
		return nil, fmt.Errorf("no main program found among %d file(s) — every file is a `unit`; exactly one must be a `program`", len(programs))
	}

	// Collect Uses across all files (main + units) so stdlib/module-aware
	// codegen paths (e.g. the stdlib dispatch in stdlib.go) see the full set.
	usesSeen := make(map[string]bool)
	var uses []string
	addUses := func(list []string) {
		for _, u := range list {
			if !usesSeen[u] {
				usesSeen[u] = true
				uses = append(uses, u)
			}
		}
	}

	for _, p := range programs {
		decls = append(decls, p.Declarations...)
		addUses(p.Uses)
	}

	merged := &ast.Program{
		Name:         main.Name,
		NameToken:    main.NameToken,
		IsUnit:       false,
		Uses:         uses,
		Declarations: decls,
		Statements:   main.Statements,
	}
	return merged, nil
}

// CompileFilesToNative parses multiple .klx files (a main program plus any
// `unit` dependencies it `uses`), merges their declarations, and compiles the
// result to a native binary via the standard LLVM pipeline. outBin/srcFile
// naming for the emitted .ll/.o/binary follows the first file in the list
// (conventionally the main program), matching the single-file CLI's naming.
func CompileFilesToNative(files []string, outBin string, llvmPaths *LLVMPaths, opts CompileOpts) (*CompileResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no source files provided")
	}
	if len(files) == 1 {
		return CompileToNativeOpts(files[0], outBin, llvmPaths, opts)
	}

	programs := make([]*ast.Program, 0, len(files))
	for _, f := range files {
		prog, err := ParseFile(f)
		if err != nil {
			return nil, err
		}
		programs = append(programs, prog)
	}

	merged, err := MergePrograms(programs)
	if err != nil {
		return nil, err
	}

	// Name the emitted .ll/.o/binary after the main (non-unit) source file,
	// not necessarily files[0] — callers may list units before the program.
	mainFile := files[0]
	for i, p := range programs {
		if !p.IsUnit {
			mainFile = files[i]
			break
		}
	}

	return compileASTWithOpts(merged, mainFile, outBin, llvmPaths, opts)
}
