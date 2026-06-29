// generator.go — Generator core: struct, constructors, Generate/GenerateMulti,
// write helpers, string interpolation, and import/exception pre-scanning.
package generator

import (
	"fmt"
	"kylix/ast"
	"strings"
)

// setLengthHelperSource is the Go source for the __kylixSetLength runtime helper,
// injected into the output when SetLength() is called by Kylix code.
var setLengthHelperSource = `
func __kylixSetLength[T any](s []T, n int) []T {
	if n <= len(s) {
		return s[:n]
	}
	return append(s, make([]T, n-len(s))...)
}
`

// Generator accumulates Go source code from a Kylix AST.
type Generator struct {
	output           strings.Builder
	indent           int
	sourceFile       string // current Kylix source file (for //line directives)
	program          *ast.Program
	variables        map[string]string // tracks variable types for codegen hints
	inFunction       bool
	inReturnFunc     bool                            // true when current function has a return value (Exit → return result)
	inExceptHandler  bool                            // true when inside a recover() block for bare raise
	reRaiseVar       string                          // Go variable holding the recovered value for re-raise
	nameMap          map[string]string               // temporary name substitutions (e.g., E→e in on clause)
	imports          map[string]bool                 // Go imports needed by the output
	needsException   bool                            // whether Exception type must be emitted
	needsSetLength   bool                            // whether the __kylixSetLength helper is needed
	exceptionTypes   map[string]bool                 // exception type names from on clauses
	multiReturn      bool                            // current function has multiple return values
	multiReturnN     int                             // number of return values in current function
	classTypes       map[string]bool                 // user-defined class type names
	classIsBase      map[string]bool                 // true if class is a parent (→ interface{} in type exprs)
	classFields      map[string][]string             // class name → ordered field names (for constructor mapping)
	classTypeParams  map[string][]*ast.TypeParameter // class name → generic type parameters
	bootRoutes       []bootRoute                     // KylixBoot annotation-generated routes
	bootComponents   []bootComponent                 // KylixBoot service/component singletons
	bootInjects      []bootInjectField               // KylixBoot [Inject] field assignments
	validationFields map[string][]validationField    // class name → fields with validation attrs
	validatedOrder   []string                        // deterministic order of validated classes
	ormEntities      map[string]*ormEntity           // ORM [Entity] metadata
	ormEntitiesOrder []string                        // deterministic order for ORM emission
	ormRepositories  []ormRepository                 // ORM [Repository] classes
	userFuncs        map[string]bool                 // user-defined function names (override built-in mapping)
	usedModules      map[string]bool                 // modules imported via `uses` clause
}

func New() *Generator {
	return &Generator{
		variables:        make(map[string]string),
		nameMap:          make(map[string]string),
		imports:          make(map[string]bool),
		exceptionTypes:   make(map[string]bool),
		classTypes:       make(map[string]bool),
		classIsBase:      make(map[string]bool),
		classFields:      make(map[string][]string),
		classTypeParams:  make(map[string][]*ast.TypeParameter),
		validationFields: make(map[string][]validationField),
		ormEntities:      make(map[string]*ormEntity),
		userFuncs:        make(map[string]bool),
		usedModules:      make(map[string]bool),
	}
}

// SetSourceFile sets the Kylix source filename used in //line directives.
func (g *Generator) SetSourceFile(f string) { g.sourceFile = f }

// CollectClassTypes is the exported pre-scan pass for cross-package use.
func (g *Generator) CollectClassTypes(p *ast.Program) { g.collectClassTypes(p) }

// ScanImports is the exported pre-scan pass for cross-package use.
func (g *Generator) ScanImports(p *ast.Program) { g.scanImports(p) }

// ScanForException is the exported pre-scan pass for cross-package use.
func (g *Generator) ScanForException(p *ast.Program) { g.scanForException(p) }

// ScanBootAnnotations is the exported pre-scan pass for KylixBoot annotation auto-wiring.
func (g *Generator) ScanBootAnnotations(p *ast.Program) { g.scanBootAnnotations(p) }

// ScanValidationAnnotations is the exported pre-scan pass for field validation annotations.
func (g *Generator) ScanValidationAnnotations(p *ast.Program) { g.scanValidationAnnotations(p) }

// ScanORMAnnotations is the exported pre-scan pass for [Entity]/[Repository]/[Query] annotations.
func (g *Generator) ScanORMAnnotations(p *ast.Program) { g.scanORMAnnotations(p) }

// WriteExceptionTypes emits the shared exception type declarations.
func (g *Generator) WriteExceptionTypes() { g.writeExceptionTypes() }

// BuildOutput assembles the final Go source from pre-scanned import state and
// a slice of body fragments (one per source file, from GenerateBody).
// Exception types (if needed) are inserted before the body fragments.
func (g *Generator) BuildOutput(bodies []string) string {
	var out strings.Builder
	out.WriteString("package main\n\n")
	if len(g.imports) > 0 {
		out.WriteString("import (\n")
		for imp := range g.imports {
			out.WriteString(fmt.Sprintf("\t%q\n", imp))
		}
		out.WriteString(")\n\n")
	}
	// Exception runtime types (captured by a temporary output snapshot).
	if g.needsException {
		snap := g.output.Len()
		g.writeExceptionTypes()
		full := g.output.String()
		out.WriteString(full[snap:])
	}
	for _, b := range bodies {
		out.WriteString(b)
	}
	// Runtime helpers (e.g. __kylixSetLength) detected during GenerateBody.
	if g.needsSetLength {
		out.WriteString(setLengthHelperSource)
	}
	return out.String()
}

// Generate compiles a single Kylix program to Go source.
func (g *Generator) Generate(program *ast.Program) string {
	g.program = program

	// Pre-scan passes collect metadata before any output is written.
	g.collectClassTypes(program)
	g.scanImports(program)
	g.scanBootAnnotations(program)
	g.scanValidationAnnotations(program)
	g.scanORMAnnotations(program)
	g.scanForException(program)

	g.writeLine("package main")
	g.writeLine("")
	g.writeImports()

	// Type declarations first (classes, interfaces, type aliases).
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			g.generateTypeDecl(d)
		case *ast.ClassDecl:
			g.generateClassDecl(d)
		case *ast.InterfaceDecl:
			g.generateInterfaceDecl(d)
		}
	}

	g.writeExceptionTypes()

	// Global variables and constants.
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.VarDecl:
			g.generateGlobalVarDecl(d)
		case *ast.ConstDecl:
			g.generateConstDecl(d)
		}
	}

	// Functions.
	for _, decl := range program.Declarations {
		if d, ok := decl.(*ast.FunctionDecl); ok {
			g.generateFunctionDecl(d)
		}
	}

	// main() from top-level statements (unit files have no main).
	if !program.IsUnit && len(program.Statements) > 0 {
		g.writeLine("func main() {")
		g.indent++
		g.emitBootAutoWiring()
		for _, stmt := range program.Statements {
			g.generateStatement(stmt)
		}
		g.indent--
		g.writeLine("}")
	}

	g.writeRuntimeHelpers()
	return g.output.String()
}

// writeRuntimeHelpers appends generated runtime helper functions that were
// found to be needed during code generation (e.g. __kylixSetLength).
// Go allows top-level functions in any order, so emitting these at the end is safe.
func (g *Generator) writeRuntimeHelpers() {
	if g.needsSetLength {
		g.writeLine("")
		g.writeLine("// __kylixSetLength resizes a slice to length n, growing with zero values")
		g.writeLine("// or truncating as needed. Backs the Kylix SetLength built-in.")
		g.writeLine("func __kylixSetLength[T any](s []T, n int) []T {")
		g.indent++
		g.writeLine("if n <= len(s) {")
		g.indent++
		g.writeLine("return s[:n]")
		g.indent--
		g.writeLine("}")
		g.writeLine("return append(s, make([]T, n-len(s))...)")
		g.indent--
		g.writeLine("}")
	}
}

// GenerateMulti compiles multiple Kylix source files into a single Go package.
func (g *Generator) GenerateMulti(programs []*ast.Program) string {
	for _, prog := range programs {
		g.collectClassTypes(prog)
	}
	for _, prog := range programs {
		g.scanImports(prog)
		g.scanBootAnnotations(prog)
		g.scanValidationAnnotations(prog)
		g.scanORMAnnotations(prog)
		g.scanForException(prog)
	}

	g.writeLine("package main")
	g.writeLine("")
	g.writeImports()

	for _, prog := range programs {
		for _, decl := range prog.Declarations {
			switch d := decl.(type) {
			case *ast.TypeDecl:
				g.generateTypeDecl(d)
			case *ast.ClassDecl:
				g.generateClassDecl(d)
			case *ast.InterfaceDecl:
				g.generateInterfaceDecl(d)
			}
		}
	}

	g.writeExceptionTypes()

	for _, prog := range programs {
		for _, decl := range prog.Declarations {
			switch d := decl.(type) {
			case *ast.VarDecl:
				g.generateGlobalVarDecl(d)
			case *ast.ConstDecl:
				g.generateConstDecl(d)
			}
		}
	}

	for _, prog := range programs {
		for _, decl := range prog.Declarations {
			if d, ok := decl.(*ast.FunctionDecl); ok {
				g.generateFunctionDecl(d)
			}
		}
	}

	for _, prog := range programs {
		if !prog.IsUnit && len(prog.Statements) > 0 {
			g.writeLine("func main() {")
			g.indent++
			g.emitBootAutoWiring()
			for _, stmt := range prog.Statements {
				g.generateStatement(stmt)
			}
			g.indent--
			g.writeLine("}")
		}
	}

	g.writeRuntimeHelpers()
	return g.output.String()
}

// GenerateBody generates only the declarations and statements for one program
// (no "package main", no "import" block). Used by incremental compilation:
// each unit is generated independently, then the results are assembled by
// CompileProjectIncremental which adds the shared header once.
//
// The caller must have already run the global pre-scan passes
// (collectClassTypes / scanImports / ScanBootAnnotations / scanForException) on ALL programs so
// cross-unit type references and KylixBoot routes resolve correctly.
func (g *Generator) GenerateBody(program *ast.Program) string {
	// Snapshot current output length so we can extract only what we add.
	start := g.output.Len()

	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			g.generateTypeDecl(d)
		case *ast.ClassDecl:
			g.generateClassDecl(d)
		case *ast.InterfaceDecl:
			g.generateInterfaceDecl(d)
		}
	}

	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.VarDecl:
			g.generateGlobalVarDecl(d)
		case *ast.ConstDecl:
			g.generateConstDecl(d)
		}
	}

	for _, decl := range program.Declarations {
		if d, ok := decl.(*ast.FunctionDecl); ok {
			g.generateFunctionDecl(d)
		}
	}

	if !program.IsUnit && len(program.Statements) > 0 {
		g.writeLine("func main() {")
		g.indent++
		g.emitBootAutoWiring()
		for _, stmt := range program.Statements {
			g.generateStatement(stmt)
		}
		g.indent--
		g.writeLine("}")
	}

	full := g.output.String()
	return full[start:]
}

// writeImports emits the import block if any imports are needed.
func (g *Generator) writeImports() {
	if len(g.imports) == 0 {
		return
	}
	g.writeLine("import (")
	g.indent++
	for imp := range g.imports {
		g.writeLine(fmt.Sprintf(`"%s"`, imp))
	}
	g.indent--
	g.writeLine(")")
	g.writeLine("")
}

// writeExceptionTypes emits the runtime Exception struct and sub-types when needed.
func (g *Generator) writeExceptionTypes() {
	if !g.needsException {
		return
	}
	g.writeLine("// Kylix runtime exception base type")
	g.writeLine("type Exception struct {")
	g.indent++
	g.writeLine("Message string")
	g.indent--
	g.writeLine("}")
	g.writeLine("")
	g.writeLine("func (e *Exception) Error() string { return e.Message }")
	g.writeLine("")

	for excType := range g.exceptionTypes {
		if excType != "Exception" {
			g.writeLine(fmt.Sprintf("type %s struct { Exception }", excType))
			g.writeLine("")
		}
	}
}

func (g *Generator) write(s string) {
	g.output.WriteString(s)
}

func (g *Generator) writeIndent() {
	for i := 0; i < g.indent; i++ {
		g.output.WriteString("\t")
	}
}

func (g *Generator) writeLine(s string) {
	for i := 0; i < g.indent; i++ {
		g.output.WriteString("\t")
	}
	g.output.WriteString(s)
	g.output.WriteString("\n")
}

// writeLineDirective emits a Go //line directive so the Go compiler maps
// errors back to the original Kylix source file and line number.
func (g *Generator) writeLineDirective(line int) {
	if g.sourceFile == "" || line <= 0 {
		return
	}
	g.output.WriteString(fmt.Sprintf("//line %s:%d\n", g.sourceFile, line))
}

// writeInterpolation emits a fmt.Sprintf call for string interpolation.
// e.g., `Hello, ${name}!` → fmt.Sprintf("Hello, %v!", name)
func (g *Generator) writeInterpolation(interp *ast.StringInterpolation) {
	var formatParts []string
	var exprParts []string

	for _, part := range interp.Parts {
		switch p := part.(type) {
		case *ast.StringLiteral:
			formatParts = append(formatParts, p.Value)
		default:
			formatParts = append(formatParts, "%v")
			oldOutput := g.output
			g.output = strings.Builder{}
			g.generateExpression(p)
			exprParts = append(exprParts, g.output.String())
			g.output = oldOutput
		}
	}

	if len(exprParts) == 0 {
		g.write(fmt.Sprintf(`"%s"`, strings.Join(formatParts, "")))
	} else {
		g.imports["fmt"] = true
		g.write("fmt.Sprintf(")
		g.write(fmt.Sprintf(`"%s"`, strings.Join(formatParts, "")))
		for _, arg := range exprParts {
			g.write(", " + arg)
		}
		g.write(")")
	}
}

// ── Pre-scan passes ──────────────────────────────────────────────────────────

// collectClassTypes records all class names and parent–child relationships
// so generateTypeExpression can decide interface{} vs *ClassName.
func (g *Generator) collectClassTypes(program *ast.Program) {
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			// Track user-defined functions so mapBuiltinFunction doesn't
			// rewrite calls to them as Go built-ins (e.g. Abs → math.Abs).
			g.userFuncs[d.Name] = true
		case *ast.ClassDecl:
			g.classTypes[d.Name] = true
			g.classTypeParams[d.Name] = d.TypeParams
			if d.Parent != "" {
				g.classIsBase[d.Parent] = true
			}
			for _, field := range d.Fields {
				for _, name := range field.Names {
					g.classFields[d.Name] = append(g.classFields[d.Name], name)
				}
			}
		case *ast.TypeDecl:
			if cd, ok := d.Type.(*ast.ClassDecl); ok {
				g.classTypes[d.Name] = true
				g.classTypeParams[d.Name] = cd.TypeParams
				if cd.Parent != "" {
					g.classIsBase[cd.Parent] = true
				}
				for _, field := range cd.Fields {
					for _, name := range field.Names {
						g.classFields[d.Name] = append(g.classFields[d.Name], name)
					}
				}
			}
		}
	}
}

// scanForException walks the program looking for try/raise so we know whether
// the Exception runtime type needs to be emitted.
func (g *Generator) scanForException(program *ast.Program) {
	for _, stmt := range program.Statements {
		g.scanStatementForException(stmt)
	}
	for _, decl := range program.Declarations {
		if fn, ok := decl.(*ast.FunctionDecl); ok && fn.Body != nil {
			for _, stmt := range fn.Body.Statements {
				g.scanStatementForException(stmt)
			}
		}
		if class, ok := decl.(*ast.ClassDecl); ok {
			for _, method := range class.Methods {
				if method.Body != nil {
					for _, stmt := range method.Body.Statements {
						g.scanStatementForException(stmt)
					}
				}
			}
		}
	}
}

func (g *Generator) scanStatementForException(stmt ast.Statement) {
	if stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.TryStatement:
		g.needsException = true
		for _, on := range s.OnClauses {
			if on.Type != nil {
				if ident, ok := on.Type.(*ast.Identifier); ok {
					g.exceptionTypes[ident.Value] = true
				}
			}
		}
	case *ast.RaiseStatement:
		g.needsException = true
	case *ast.IfStatement:
		if s.Consequence != nil {
			for _, st := range s.Consequence.Statements {
				g.scanStatementForException(st)
			}
		}
		if s.Alternative != nil {
			for _, st := range s.Alternative.Statements {
				g.scanStatementForException(st)
			}
		}
	case *ast.WhileStatement:
		if s.Body != nil {
			for _, st := range s.Body.Statements {
				g.scanStatementForException(st)
			}
		}
	case *ast.ForStatement:
		if s.Body != nil {
			for _, st := range s.Body.Statements {
				g.scanStatementForException(st)
			}
		}
	case *ast.BlockStatement:
		for _, st := range s.Statements {
			g.scanStatementForException(st)
		}
	}
}

// scanImports maps uses clause modules and built-in function calls to Go imports.
func (g *Generator) scanImports(program *ast.Program) {
	// uses clause → stdlib package + record used modules
	for _, module := range program.Uses {
		g.usedModules[module] = true
		switch module {
		case "web", "container", "config", "middleware", "validation",
			"orm", "template", "autoconfig", "sysutil", "jsonutil",
			"datetime", "regex", "strutil", "mathutil", "arrayutil",
			"collections", "httpclient", "iter", "stringbuilder",
			"resulttype", "wasi", "boot", "net", "crypto", "encoding", "jwt", "db":
			g.imports["kylix/stdlib"] = true
		}
	}

	for _, stmt := range program.Statements {
		g.scanStatementForImports(stmt)
	}
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			if d.Body != nil {
				for _, stmt := range d.Body.Statements {
					g.scanStatementForImports(stmt)
				}
			}
		case *ast.ClassDecl:
			for _, method := range d.Methods {
				if method.Body != nil {
					for _, stmt := range method.Body.Statements {
						g.scanStatementForImports(stmt)
					}
				}
			}
		case *ast.TypeDecl:
			// Class methods often live inside TypeDecl.Type (a ClassDecl).
			// Scan them too so WriteLn etc. trigger fmt import.
			if cd, ok := d.Type.(*ast.ClassDecl); ok {
				for _, method := range cd.Methods {
					if method.Body != nil {
						for _, stmt := range method.Body.Statements {
							g.scanStatementForImports(stmt)
						}
					}
				}
			}
		}
	}
}

func (g *Generator) scanStatementForImports(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		g.scanExpressionForImports(s.Expression)
	case *ast.AssignmentStatement:
		g.scanExpressionForImports(s.Value)
	case *ast.VarDecl:
		if s.Value != nil {
			g.scanExpressionForImports(s.Value)
		}
	case *ast.IfStatement:
		g.scanExpressionForImports(s.Condition)
		if s.Consequence != nil {
			for _, st := range s.Consequence.Statements {
				g.scanStatementForImports(st)
			}
		}
		if s.Alternative != nil {
			for _, st := range s.Alternative.Statements {
				g.scanStatementForImports(st)
			}
		}
	case *ast.WhileStatement:
		g.scanExpressionForImports(s.Condition)
		if s.Body != nil {
			for _, st := range s.Body.Statements {
				g.scanStatementForImports(st)
			}
		}
	case *ast.ForStatement:
		if s.Body != nil {
			for _, st := range s.Body.Statements {
				g.scanStatementForImports(st)
			}
		}
	case *ast.ForEachStatement:
		g.scanExpressionForImports(s.Iterable)
		if s.Body != nil {
			for _, st := range s.Body.Statements {
				g.scanStatementForImports(st)
			}
		}
	case *ast.TryStatement:
		if s.Body != nil {
			for _, st := range s.Body.Statements {
				g.scanStatementForImports(st)
			}
		}
		for _, on := range s.OnClauses {
			if on.Body != nil {
				for _, st := range on.Body.Statements {
					g.scanStatementForImports(st)
				}
			}
		}
		if s.ExceptBlock != nil {
			for _, st := range s.ExceptBlock.Statements {
				g.scanStatementForImports(st)
			}
		}
		if s.FinallyBlock != nil {
			for _, st := range s.FinallyBlock.Statements {
				g.scanStatementForImports(st)
			}
		}
	case *ast.BlockStatement:
		for _, st := range s.Statements {
			g.scanStatementForImports(st)
		}
	case *ast.MatchStatement:
		for _, branch := range s.Branches {
			if branch.Body != nil {
				for _, st := range branch.Body.Statements {
					g.scanStatementForImports(st)
				}
			}
			g.scanExpressionForImports(branch.Pattern)
			if branch.When != nil {
				g.scanExpressionForImports(branch.When)
			}
		}
	}
}

func (g *Generator) scanExpressionForImports(expr ast.Expression) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		g.mapBuiltinFunction(e.Value)
	case *ast.CallExpression:
		g.scanExpressionForImports(e.Function)
		for _, arg := range e.Arguments {
			g.scanExpressionForImports(arg)
		}
		if ident, ok := e.Function.(*ast.Identifier); ok {
			if ident.Value == "StrToInt64" || ident.Value == "StrToFloat" {
				g.imports["strconv"] = true
			}
			if ident.Value == "ReadFile" {
				// Only add `os` import when ReadFile uses the legacy inline
				// codegen (no `uses sysutil`). When sysutil is imported,
				// ReadFile is dispatched to stdlib.ReadFile instead.
				if !g.usedModules["sysutil"] {
					g.imports["os"] = true
				}
			}
		}
	case *ast.InfixExpression:
		g.scanExpressionForImports(e.Left)
		g.scanExpressionForImports(e.Right)
	case *ast.PrefixExpression:
		g.scanExpressionForImports(e.Right)
	case *ast.MemberExpression:
		g.scanExpressionForImports(e.Object)
	case *ast.IndexExpression:
		g.scanExpressionForImports(e.Left)
		g.scanExpressionForImports(e.Index)
	case *ast.LambdaExpression:
		switch body := e.Body.(type) {
		case *ast.BlockStatement:
			for _, stmt := range body.Statements {
				g.scanStatementForImports(stmt)
			}
		case ast.Expression:
			g.scanExpressionForImports(body)
		}
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			g.scanExpressionForImports(part)
		}
	}
}
