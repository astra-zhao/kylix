package generator

import (
	"fmt"
	"kylix/ast"
	"strings"
)

type Generator struct {
	output          strings.Builder
	indent          int
	program         *ast.Program
	variables       map[string]string // track variable types
	inFunction      bool
	inReturnFunc    bool   // true if current function has a return value (Exit → return result)
	inExceptHandler bool             // tracks if we're inside recover() for bare raise
	reRaiseVar      string           // variable name holding the recovered value for re-raise
	nameMap         map[string]string // temporary name substitutions (e.g., E→e in on clause)
	imports         map[string]bool  // track which imports are needed
	needsException  bool             // whether Exception type needs to be generated
	exceptionTypes  map[string]bool  // exception type names referenced in on clauses
	multiReturn     bool             // whether current function has multiple return values
	multiReturnN    int              // number of return values in current function
	classTypes      map[string]bool  // track which user-defined types are classes
	classIsBase     map[string]bool  // true if class is parent of another class (→ interface{} in type expressions)
	classFields     map[string][]string // class name → ordered field names (for constructor mapping)
}

func New() *Generator {
	return &Generator{
		variables:      make(map[string]string),
		nameMap:        make(map[string]string),
		imports:        make(map[string]bool),
		exceptionTypes: make(map[string]bool),
		classTypes:     make(map[string]bool),
		classIsBase:    make(map[string]bool),
	classFields:    make(map[string][]string),
	}
}

func (g *Generator) Generate(program *ast.Program) string {
	g.program = program

	// First pass: scan for needed imports and exception usage
	g.collectClassTypes(program)
	g.scanImports(program)
	g.scanForException(program)

	// Generate package and imports
	g.writeLine("package main")
	g.writeLine("")
	if len(g.imports) > 0 {
		g.writeLine("import (")
		g.indent++
		for imp := range g.imports {
			g.writeLine(fmt.Sprintf(`"%s"`, imp))
		}
		g.indent--
		g.writeLine(")")
		g.writeLine("")
	}

	// Generate type declarations
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

	// Generate runtime Exception type if needed
	if g.needsException {
		g.writeLine("// Kylix runtime: Exception type for try/except/raise")
		g.writeLine("type Exception struct {")
		g.indent++
		g.writeLine("Message string")
		g.indent--
		g.writeLine("}")
		g.writeLine("")
		g.writeLine("func (e *Exception) Error() string { return e.Message }")
		g.writeLine("")

		// Generate sub-types for all exception types referenced in on clauses
		for excType := range g.exceptionTypes {
			if excType != "Exception" {
				g.writeLine(fmt.Sprintf("// %s is a sub-type of Exception", excType))
				g.writeLine(fmt.Sprintf("type %s struct {", excType))
				g.indent++
				g.writeLine("Exception")
				g.indent--
				g.writeLine("}")
				g.writeLine("")
			}
		}
	}

	// Generate global variables and constants
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.VarDecl:
			g.generateGlobalVarDecl(d)
		case *ast.ConstDecl:
			g.generateConstDecl(d)
		}
	}

	// Generate functions
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			g.generateFunctionDecl(d)
		}
	}

	// Generate main function if there are top-level statements (skip for unit files)
	if !program.IsUnit && len(program.Statements) > 0 {
		g.writeLine("func main() {")
		g.indent++
		for _, stmt := range program.Statements {
			g.generateStatement(stmt)
		}
		g.indent--
		g.writeLine("}")
	}

	return g.output.String()
}

// GenerateMulti compiles multiple Kylix source files into a single Go package.
func (g *Generator) GenerateMulti(programs []*ast.Program) string {
	// Pre-scan: collect all class types from all programs
	for _, prog := range programs {
		g.collectClassTypes(prog)
	}

	for _, prog := range programs {
		g.scanImports(prog)
		g.scanForException(prog)
	}

	g.writeLine("package main")
	g.writeLine("")
	if len(g.imports) > 0 {
		g.writeLine("import (")
		g.indent++
		for imp := range g.imports {
			g.writeLine(fmt.Sprintf(`"%s"`, imp))
		}
		g.indent--
		g.writeLine(")")
		g.writeLine("")
	}

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

	if g.needsException {
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
				g.writeLine(fmt.Sprintf("type %s struct {", excType))
				g.indent++
				g.writeLine("Exception")
				g.indent--
				g.writeLine("}")
				g.writeLine("")
			}
		}
	}

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
			for _, stmt := range prog.Statements {
				g.generateStatement(stmt)
			}
			g.indent--
			g.writeLine("}")
		}
	}

	return g.output.String()
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

// writeInterpolation generates a fmt.Sprintf call for string interpolation.
// e.g., 'Hello, ${name}!' → fmt.Sprintf("Hello, %v!", name)
func (g *Generator) writeInterpolation(interp *ast.StringInterpolation) {
	var formatParts []string
	var exprParts []string

	for _, part := range interp.Parts {
		switch p := part.(type) {
		case *ast.StringLiteral:
			formatParts = append(formatParts, p.Value)
		default:
			formatParts = append(formatParts, "%v")
			// Capture the expression output into a buffer
			oldOutput := g.output
			g.output = strings.Builder{}
			g.generateExpression(p)
			exprParts = append(exprParts, g.output.String())
			g.output = oldOutput
		}
	}

	if len(exprParts) == 0 {
		// No interpolation, just a plain string
		g.write(fmt.Sprintf(`"%s"`, strings.Join(formatParts, "")))
	} else {
		g.imports["fmt"] = true
		g.write("fmt.Sprintf(")
		g.write(fmt.Sprintf(`"%s"`, strings.Join(formatParts, "")))
		for _, arg := range exprParts {
			g.write(", ")
			g.write(arg)
		}
		g.write(")")
	}
}

func (g *Generator) generateTypeDecl(decl *ast.TypeDecl) {
	// Handle class/interface declarations that are wrapped in TypeDecl
	if classDecl, ok := decl.Type.(*ast.ClassDecl); ok {
		classDecl.Name = decl.Name
		g.generateClassDecl(classDecl)
		return
	}
	if ifaceDecl, ok := decl.Type.(*ast.InterfaceDecl); ok {
		ifaceDecl.Name = decl.Name
		g.generateInterfaceDecl(ifaceDecl)
		return
	}
	if variantDecl, ok := decl.Type.(*ast.VariantType); ok {
		g.generateVariantType(decl.Name, variantDecl)
		return
	}
	if enumDecl, ok := decl.Type.(*ast.EnumType); ok {
		g.generateEnumType(decl.Name, enumDecl)
		return
	}

	g.write(fmt.Sprintf("type %s ", decl.Name))
	g.generateTypeExpression(decl.Type)
	g.write("\n\n")
}

func (g *Generator) generateClassDecl(decl *ast.ClassDecl) {
	// Record as class type
	g.classTypes[decl.Name] = true

	// All classes generate as plain structs with parent embedding.
	// Polymorphism is handled via interface{} at the type-expression level
	// (for fields/params typed as base classes like TNode/TStatement/TExpression).
	g.write("type ")
	g.write(decl.Name)
	g.generateTypeParams(decl.TypeParams)
	g.writeLine(" struct {")
	g.indent++
	if decl.Parent != "" {
		g.writeLine(decl.Parent)
	}
	for _, field := range decl.Fields {
		for _, name := range field.Names {
			g.write(name + " ")
			if field.Type != nil {
				g.generateTypeExpression(field.Type)
			} else {
				g.write("interface{}")
			}
			g.write("\n")
		}
	}
	g.indent--
	g.writeLine("}")
	g.writeLine("")

	for _, method := range decl.Methods {
		g.generateClassMethod(decl.Name, method)
	}

	for _, prop := range decl.Properties {
		g.generatePropertyAccessors(decl.Name, prop)
	}
}

// generateClassMethod generates a single method of a class, including
// the var result declaration and local var/const declarations.
func (g *Generator) generateClassMethod(className string, method *ast.FunctionDecl) {
	hasReturnType := method.ReturnType != nil || len(method.ReturnTypes) > 0

	g.write(fmt.Sprintf("func (self *%s) %s", className, method.Name))
	g.generateFunctionSignature(method)
	g.writeLine(" {")
	g.indent++

	if hasReturnType {
		if method.ReturnType != nil {
			g.write("var result ")
			g.generateTypeExpression(method.ReturnType)
			g.write("\n")
		} else if len(method.ReturnTypes) == 1 {
			g.write("var result ")
			g.generateTypeExpression(method.ReturnTypes[0])
			g.write("\n")
		}
	}

	// Generate local var/const declarations
	for _, local := range method.LocalDecls {
		switch d := local.(type) {
		case *ast.VarDecl:
			g.generateLocalVarDecl(d)
		case *ast.ConstDecl:
			g.generateLocalConstDecl(d)
		}
	}

	if method.Body != nil {
		g.inFunction = true
		g.inReturnFunc = hasReturnType
		for _, stmt := range method.Body.Statements {
			g.generateStatement(stmt)
		}
		g.inFunction = false
		g.inReturnFunc = false
	}

	// Suppress "declared and not used" for local vars in class methods
	for _, local := range method.LocalDecls {
		if vd, ok := local.(*ast.VarDecl); ok && len(vd.Names) == 1 {
			g.write(fmt.Sprintf("_ = %s\n", vd.Names[0]))
		}
	}

	if hasReturnType {
		g.writeLine("return result")
	}

	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

// generatePropertyAccessors generates getter and setter methods for a property.
// Pascal property: property GetName: String read Name;
// Generates: func (self *ClassName) GetName() string { return self.Name }
func (g *Generator) generatePropertyAccessors(className string, prop *ast.PropertyDecl) {
	if prop.Getter != "" {
		// Generate getter: func (self *ClassName) PropName() Type { return self.FieldName }
		g.write(fmt.Sprintf("func (self *%s) %s() ", className, prop.Name))
		if prop.Type != nil {
			g.generateTypeExpression(prop.Type)
		} else {
			g.write("interface{}")
		}
		g.writeLine(" {")
		g.indent++
		g.write(fmt.Sprintf("return self.%s\n", prop.Getter))
		g.indent--
		g.writeLine("}")
		g.writeLine("")
	}

	if prop.Setter != "" {
		// Generate setter: func (self *ClassName) SetPropName(v Type) { self.FieldName = v }
		g.write(fmt.Sprintf("func (self *%s) Set%s(v ", className, prop.Name))
		if prop.Type != nil {
			g.generateTypeExpression(prop.Type)
		} else {
			g.write("interface{}")
		}
		g.writeLine(") {")
		g.indent++
		g.write(fmt.Sprintf("self.%s = v\n", prop.Setter))
		g.indent--
		g.writeLine("}")
		g.writeLine("")
	}
}

func (g *Generator) generateInterfaceDecl(decl *ast.InterfaceDecl) {
	g.writeLine(fmt.Sprintf("type %s interface {", decl.Name))
	g.indent++

	// Embed parent interfaces
	for _, parent := range decl.Parents {
		g.writeLine(parent)
	}

	// Add method signatures
	for _, method := range decl.Methods {
		g.write(method.Name)
		g.generateFunctionSignature(method)
		g.write("\n")
	}

	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

// generateVariantType generates a discriminated union: an interface + concrete types
func (g *Generator) generateVariantType(name string, variant *ast.VariantType) {
	g.writeLine(fmt.Sprintf("type %s interface {", name))
	g.indent++
	g.writeLine(fmt.Sprintf("is%s()", name))
	g.indent--
	g.writeLine("}")
	g.writeLine("")

	for _, c := range variant.Cases {
		structName := fmt.Sprintf("%s_%s", name, c.Name)
		g.writeLine(fmt.Sprintf("type %s struct {", structName))
		g.indent++
		if c.Type != nil {
			g.write("Value ")
			g.generateTypeExpression(c.Type)
			g.write("\n")
		}
		g.indent--
		g.writeLine("}")
		g.writeLine("")
		g.writeLine(fmt.Sprintf("func (s *%s) is%s() {}", structName, name))
		g.writeLine("")
	}
}

// generateEnumType generates Go const + iota for enum types
func (g *Generator) generateEnumType(name string, enum *ast.EnumType) {
	if len(enum.Names) == 0 {
		return
	}
	g.writeLine("const (")
	g.indent++
	for i, n := range enum.Names {
		if i == 0 {
			g.writeLine(fmt.Sprintf("%s %s = iota", n, name))
		} else {
			g.writeLine(n)
		}
	}
	g.indent--
	g.writeLine(")")
	g.writeLine("")
	g.writeLine(fmt.Sprintf("type %s int", name))
	g.writeLine("")
}

func (g *Generator) generateGlobalVarDecl(decl *ast.VarDecl) {
	g.write("var ")
	for i, name := range decl.Names {
		if i > 0 {
			g.write(", ")
		}
		g.write(name)
	}

	if decl.Type != nil {
		g.write(" ")
		g.generateTypeExpression(decl.Type)
	}

	if decl.Value != nil {
		g.write(" = ")
		g.generateExpression(decl.Value)
	} else if _, isMap := decl.Type.(*ast.MapType); isMap {
		g.write(" = ")
		g.generateTypeExpression(decl.Type)
		g.write("{}")
	}

	g.write("\n")
}

func (g *Generator) generateConstDecl(decl *ast.ConstDecl) {
	g.write(fmt.Sprintf("const %s", decl.Name))
	if decl.Type != nil {
		g.write(" ")
		g.generateTypeExpression(decl.Type)
	}
	if decl.Value != nil {
		g.write(" = ")
		g.generateExpression(decl.Value)
	}
	g.write("\n")
}

func (g *Generator) generateFunctionDecl(decl *ast.FunctionDecl) {
	hasReturnType := decl.ReturnType != nil || len(decl.ReturnTypes) > 0
	hasMultiReturn := len(decl.ReturnTypes) > 1
	g.multiReturn = hasMultiReturn
	g.multiReturnN = len(decl.ReturnTypes)

	if decl.IsAsync {
		// Async function: generate func(...) <-chan ReturnType { ... }
		g.write(fmt.Sprintf("func %s", decl.Name))
		g.generateTypeParams(decl.TypeParams)
		g.generateAsyncSignature(decl)
		g.writeLine(" {")
		g.indent++

		// Create channel
		if hasReturnType {
			g.write("ch := make(chan ")
			g.generateTypeExpression(decl.ReturnType)
			g.writeLine(", 1)")
		} else {
			g.writeLine("ch := make(chan bool, 1)")
		}

		// Launch goroutine
		g.writeLine("go func() {")
		g.indent++

		if hasReturnType {
			g.write("var result ")
			g.generateTypeExpression(decl.ReturnType)
			g.write("\n")
		}

		if decl.Body != nil {
			g.inFunction = true
			for _, stmt := range decl.Body.Statements {
				g.generateStatement(stmt)
			}
			g.inFunction = false
		}

		// Send result to channel
		if hasReturnType {
			g.writeLine("ch <- result")
		} else {
			g.writeLine("ch <- true")
		}

		g.indent--
		g.writeLine("}()")

		g.writeLine("return ch")
		g.indent--
		g.writeLine("}")
		g.writeLine("")
		return
	}

	// Regular (non-async) function
	// Check if this is a method definition: ClassName.MethodName
	if idx := strings.Index(decl.Name, "."); idx >= 0 {
		className := decl.Name[:idx]
		methodName := decl.Name[idx+1:]
		g.write(fmt.Sprintf("func (self *%s) %s", className, methodName))
	} else {
		g.write(fmt.Sprintf("func %s", decl.Name))
	}
	g.generateTypeParams(decl.TypeParams)
	g.generateFunctionSignature(decl)
	g.writeLine(" {")
	g.indent++

	if hasReturnType {
		if hasMultiReturn {
			// Multi-return: no result variable, use direct returns
			// result := (expr1, expr2) is handled as return in the body
		} else if decl.ReturnType != nil {
			g.write("var result ")
			g.generateTypeExpression(decl.ReturnType)
			g.write("\n")
		} else {
			// Single return from ReturnTypes list
			g.write("var result ")
			g.generateTypeExpression(decl.ReturnTypes[0])
			g.write("\n")
		}
	}

	// Generate local var/const declarations
	for _, local := range decl.LocalDecls {
		switch d := local.(type) {
		case *ast.VarDecl:
			g.generateLocalVarDecl(d)
		case *ast.ConstDecl:
			g.generateLocalConstDecl(d)
		}
	}

	if decl.Body != nil {
		g.inFunction = true
		g.inReturnFunc = hasReturnType
		for _, stmt := range decl.Body.Statements {
			g.generateStatement(stmt)
		}
		g.inFunction = false
		g.inReturnFunc = false
	}

	// Suppress "declared and not used" for local vars
	for _, local := range decl.LocalDecls {
		if vd, ok := local.(*ast.VarDecl); ok && len(vd.Names) == 1 {
			g.write(fmt.Sprintf("_ = %s\n", vd.Names[0]))
		}
	}

	if hasReturnType {
		if hasMultiReturn {
			// Multi-return: each path has explicit result := (x, y) → return x, y
			// No fallback return — Go compiler will catch uncovered paths
		} else {
			g.write("return result\n")
		}
	}
	g.multiReturn = false
	g.multiReturnN = 0

	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

func (g *Generator) generateAsyncSignature(decl *ast.FunctionDecl) {
	g.write("(")
	for i, param := range decl.Parameters {
		if i > 0 {
			g.write(", ")
		}
		g.write(param.Name)
		if param.Type != nil {
			g.write(" ")
			g.generateTypeExpression(param.Type)
		}
	}
	g.write(")")

	// Async functions return <-chan type
	if decl.ReturnType != nil {
		g.write(" <-chan ")
		g.generateTypeExpression(decl.ReturnType)
	}
}

func (g *Generator) generateTypeParams(params []*ast.TypeParameter) {
	if len(params) == 0 {
		return
	}
	g.write("[")
	for i, tp := range params {
		if i > 0 {
			g.write(", ")
		}
		g.write(tp.Name)
		g.write(" ")
		if tp.Constraint != nil {
			g.generateTypeExpression(tp.Constraint)
		} else {
			g.write("interface{}")
		}
	}
	g.write("]")
}

func (g *Generator) generateFunctionSignature(decl *ast.FunctionDecl) {
	g.write("(")
	for i, param := range decl.Parameters {
		if i > 0 {
			g.write(", ")
		}
		g.write(param.Name)
		if param.Type != nil {
			g.write(" ")
			g.generateTypeExpression(param.Type)
		}
	}
	g.write(")")

	if len(decl.ReturnTypes) > 1 {
		g.write(" ")
		g.generateMultiReturnType(decl.ReturnTypes)
	} else if decl.ReturnType != nil {
		g.write(" ")
		g.generateTypeExpression(decl.ReturnType)
	}
}

func (g *Generator) generateMultiReturnType(types []ast.Expression) {
	g.write("(")
	for i, t := range types {
		if i > 0 {
			g.write(", ")
		}
		g.generateTypeExpression(t)
	}
	g.write(")")
}

func (g *Generator) generateTypeExpression(expr ast.Expression) {
	switch t := expr.(type) {
	case *ast.Identifier:
		typeName := t.Value
		if g.classTypes[typeName] {
			// Base classes (parents of other classes) → interface{} for polymorphism.
			// Concrete classes → *ClassName (pointer to struct).
			if g.classIsBase[typeName] {
				g.write("interface{}")
			} else {
				g.write("*" + typeName)
			}
		} else {
			g.write(g.mapType(typeName))
		}
	case *ast.ArrayType:
		if t.Dynamic {
			g.write("[]")
		} else {
			g.write("[")
			if t.Size != nil {
				g.generateExpression(t.Size)
			}
			g.write("]")
		}
		if t.ElementType != nil {
			g.generateTypeExpression(t.ElementType)
		} else {
			g.write("interface{}")
		}
	case *ast.MapType:
		g.write("map[")
		if t.KeyType != nil {
			g.generateTypeExpression(t.KeyType)
		} else {
			g.write("string")
		}
		g.write("]")
		if t.ValueType != nil {
			g.generateTypeExpression(t.ValueType)
		} else {
			g.write("interface{}")
		}
	case *ast.GenericType:
		// Go 1.18+ uses square brackets for generics: TPair[int64, string]
		g.write(g.mapType(t.Base))
		if len(t.TypeParams) > 0 {
			g.write("[")
			for i, param := range t.TypeParams {
				if i > 0 {
					g.write(", ")
				}
				g.generateTypeExpression(param)
			}
			g.write("]")
		}
	case *ast.RecordType:
		g.write("struct {\n")
		g.indent++
		for _, field := range t.Fields {
			for _, name := range field.Names {
				g.write(name + " ")
				if field.Type != nil {
					g.generateTypeExpression(field.Type)
				}
				g.write("\n")
			}
		}
		g.indent--
		g.write("}")
	default:
		g.write("interface{}")
	}
}

// generateTypeExpressionForCast emits a type expression suitable for a Go type
// assertion. Unlike generateTypeExpression, base class types map to *ClassName
// (the concrete struct pointer) instead of interface{}, because you can't
// type-assert to interface{}.
func (g *Generator) generateTypeExpressionForCast(expr ast.Expression) {
	if ident, ok := expr.(*ast.Identifier); ok {
		typeName := ident.Value
		if g.classTypes[typeName] {
			g.write("*" + typeName)
			return
		}
	}
	g.generateTypeExpression(expr)
}

func (g *Generator) mapType(kylixType string) string {
	typeMap := map[string]string{
		"Integer":  "int64",
		"Real":     "float64",
		"Boolean":  "bool",
		"String":   "string",
		"Char":     "byte",
		"Byte":     "byte",
		"Word":     "uint16",
		"Cardinal": "uint32",
		"LongInt":  "int64",
		"Double":   "float64",
		"Extended": "float64",
	}

	if goType, ok := typeMap[kylixType]; ok {
		return goType
	}
	return kylixType
}

func (g *Generator) mapBuiltinFunction(name string) string {
	builtinMap := map[string]string{
		"WriteLn":  "fmt.Println",
		"Write":    "fmt.Print",
		"ReadLn":   "fmt.Scanln",
		"Read":     "fmt.Scan",
		"IntToStr": "fmt.Sprintf",
		"StrToInt": "strconv.ParseInt",
		"Length":   "len",
		"Copy":     "copy",
		"Concat":   "fmt.Sprintf",
		"Pos":      "strings.Index",
		"Delete":   "delete",
		"Insert":   "insert",
		"UpperCase": "strings.ToUpper",
		"LowerCase": "strings.ToLower",
		"Trim":     "strings.TrimSpace",
		"Inc":      "++",
		"Dec":      "--",
		"Succ":     "++",
		"Pred":     "--",
		"Ord":      "int",
		"Chr":      "string",
		"Sqr":      "math.Pow",
		"Sqrt":     "math.Sqrt",
		"Abs":      "math.Abs",
		"Sin":      "math.Sin",
		"Cos":      "math.Cos",
		"Tan":      "math.Tan",
		"Ln":       "math.Log",
		"Exp":      "math.Exp",
		"Round":    "math.Round",
		"Trunc":    "math.Trunc",
		"Frac":     "math.Mod",
		"Random":   "rand.Float64",
		"Randomize": "rand.Seed",
		"Halt":     "os.Exit",
		"Exit":     "return",
		"append":   "append",
		"SetLength": "SetLength",
	}

	if goFunc, ok := builtinMap[name]; ok {
		// Track which imports are needed
		if strings.HasPrefix(goFunc, "fmt.") {
			g.imports["fmt"] = true
		} else if strings.HasPrefix(goFunc, "strings.") {
			g.imports["strings"] = true
		} else if strings.HasPrefix(goFunc, "math.") {
			g.imports["math"] = true
		} else if strings.HasPrefix(goFunc, "strconv.") {
			g.imports["strconv"] = true
		} else if strings.HasPrefix(goFunc, "os.") {
			g.imports["os"] = true
		} else if strings.HasPrefix(goFunc, "rand.") {
			g.imports["math/rand"] = true
		}
		return goFunc
	}
	return name
}

// scanForException pre-scans the program for exception-related constructs
// to determine if the Exception type needs to be generated.
func (g *Generator) scanForException(program *ast.Program) {
	// Check all top-level statements for try/raise
	for _, stmt := range program.Statements {
		g.scanStatementForException(stmt)
	}
	// Check function bodies
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
		// Collect exception type names from ON clauses
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

// collectClassTypes scans a program and records all class type names and base class relationships.
func (g *Generator) collectClassTypes(program *ast.Program) {
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.ClassDecl:
			g.classTypes[d.Name] = true
			if d.Parent != "" {
				g.classIsBase[d.Parent] = true
			}
			// Collect field names for constructor mapping
			for _, field := range d.Fields {
				for _, name := range field.Names {
					g.classFields[d.Name] = append(g.classFields[d.Name], name)
				}
			}
		case *ast.TypeDecl:
			if cd, ok := d.Type.(*ast.ClassDecl); ok {
				g.classTypes[d.Name] = true
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

func (g *Generator) scanImports(program *ast.Program) {
	// Map uses clause stdlib modules to Go imports
	for _, module := range program.Uses {
		switch module {
		case "web":
			g.imports["kylix/stdlib"] = true
		case "container":
			g.imports["kylix/stdlib"] = true
		case "config":
			g.imports["kylix/stdlib"] = true
		case "middleware":
			g.imports["kylix/stdlib"] = true
		case "validation":
			g.imports["kylix/stdlib"] = true
		case "orm":
			g.imports["kylix/stdlib"] = true
		case "template":
			g.imports["kylix/stdlib"] = true
		case "autoconfig":
			g.imports["kylix/stdlib"] = true
		case "sysutil":
			g.imports["kylix/stdlib"] = true
		case "jsonutil":
			g.imports["kylix/stdlib"] = true
		case "datetime":
			g.imports["kylix/stdlib"] = true
		case "regex":
			g.imports["kylix/stdlib"] = true
		}
	}

	// Scan all statements for builtin function calls
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
		// Check if this is a builtin function
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
				g.imports["os"] = true
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

func (g *Generator) generateStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		g.generateVarDecl(s)
	case *ast.AssignmentStatement:
		g.generateAssignment(s)
	case *ast.ExpressionStatement:
		// Handle special builtins: append(arr, elem) → arr = append(arr, elem)
		if call, ok := s.Expression.(*ast.CallExpression); ok {
			if ident, ok := call.Function.(*ast.Identifier); ok {
				if ident.Value == "append" && len(call.Arguments) >= 1 {
					g.generateExpression(call.Arguments[0])
					g.write(" = append(")
					for i, arg := range call.Arguments {
						if i > 0 {
							g.write(", ")
						}
						g.generateExpression(arg)
					}
					g.write(")\n")
					break
				}
				if ident.Value == "SetLength" && len(call.Arguments) >= 2 {
					g.generateExpression(call.Arguments[0])
					g.write(" = ")
					g.generateExpression(call.Arguments[0])
					g.write("[:")
					g.generateExpression(call.Arguments[1])
					g.write("]")
					g.write("\n")
					break
				}
			}
			// Generate standalone expression as a statement
			g.generateExpression(s.Expression)
			g.write("\n")
		} else {
			// Handle Exit: generates "return result" for functions with return values, "return" otherwise
			if ident, ok := s.Expression.(*ast.Identifier); ok && ident.Value == "Exit" {
				if g.inReturnFunc {
					g.write("return result\n")
				} else {
					g.write("return\n")
				}
				break
			}
			// Handle bare member access as procedure call: self.Method → self.Method()
			if _, ok := s.Expression.(*ast.MemberExpression); ok {
				g.generateExpression(s.Expression)
				g.write("()\n")
				break
			}
			g.generateExpression(s.Expression)
			g.write("\n")
		}
	case *ast.IfStatement:
		g.generateIfStatement(s)
	case *ast.WhileStatement:
		g.generateWhileStatement(s)
	case *ast.ForStatement:
		g.generateForStatement(s)
	case *ast.ForEachStatement:
		g.generateForEachStatement(s)
	case *ast.RepeatStatement:
		g.generateRepeatStatement(s)
	case *ast.CaseStatement:
		g.generateCaseStatement(s)
	case *ast.MatchStatement:
		g.generateMatchStatement(s)
	case *ast.TryStatement:
		g.generateTryStatement(s)
	case *ast.RaiseStatement:
		g.generateRaiseStatement(s)
	case *ast.ReturnStatement:
		g.generateReturnStatement(s)
	case *ast.BreakStatement:
		g.writeLine("break")
	case *ast.ContinueStatement:
		g.writeLine("continue")
	case *ast.InheritedStatement:
		g.generateInheritedStatement(s)
	case *ast.BlockStatement:
		for _, st := range s.Statements {
			g.generateStatement(st)
		}
	}
}

func (g *Generator) generateVarDecl(decl *ast.VarDecl) {
	// Handle destructuring: var (a, b) := expr → a, b := expr
	if len(decl.Names) > 1 && decl.Inferred {
		for i, name := range decl.Names {
			if i > 0 {
				g.write(", ")
			}
			g.write(name)
		}
		g.write(" := ")
		g.generateExpression(decl.Value)
		g.write("\n")
		return
	}

	for _, name := range decl.Names {
		if decl.Inferred {
			g.write(fmt.Sprintf("%s := ", name))
			g.generateExpression(decl.Value)
		} else {
			g.write(fmt.Sprintf("var %s", name))
			if decl.Type != nil {
				g.write(" ")
				g.generateTypeExpression(decl.Type)
			}
			if decl.Value != nil {
				g.write(" = ")
				g.generateExpression(decl.Value)
			}
		}
		g.write("\n")
	}
}

// generateLocalVarDecl generates a local variable declaration inside a function body.
func (g *Generator) generateLocalVarDecl(decl *ast.VarDecl) {
	for _, name := range decl.Names {
		g.write("var " + name + " ")
		if decl.Type != nil {
			g.generateTypeExpression(decl.Type)
		} else {
			g.write("interface{}")
		}
		g.write("\n")
	}
}

// generateLocalConstDecl generates a local constant declaration inside a function body.
func (g *Generator) generateLocalConstDecl(decl *ast.ConstDecl) {
	g.write("const " + decl.Name)
	if decl.Type != nil {
		g.write(" ")
		g.generateTypeExpression(decl.Type)
	}
	g.write(" = ")
	if decl.Value != nil {
		g.generateExpression(decl.Value)
	}
	g.write("\n")
}

func (g *Generator) generateAssignment(stmt *ast.AssignmentStatement) {
	// Handle multi-return: result := (expr1, expr2) → return expr1, expr2
	if g.multiReturn {
		if ident, ok := stmt.Name.(*ast.Identifier); ok && ident.Value == "result" {
			g.write("return ")
			if tuple, ok := stmt.Value.(*ast.TupleLiteral); ok {
				for i, elem := range tuple.Elements {
					if i > 0 {
						g.write(", ")
					}
					g.generateExpression(elem)
				}
			} else {
				g.generateExpression(stmt.Value)
			}
			g.write("\n")
			return
		}
	}
	g.generateExpression(stmt.Name)
	g.write(" = ")
	g.generateExpression(stmt.Value)
	g.write("\n")
}

func (g *Generator) generateIfStatement(stmt *ast.IfStatement) {
	g.write("if ")
	g.generateExpression(stmt.Condition)
	g.writeLine(" {")
	g.indent++
	if stmt.Consequence != nil {
		for _, s := range stmt.Consequence.Statements {
			g.generateStatement(s)
		}
	}
	g.indent--
	g.write("}")

	if stmt.Alternative != nil {
		g.writeLine(" else {")
		g.indent++
		for _, s := range stmt.Alternative.Statements {
			g.generateStatement(s)
		}
		g.indent--
		g.writeLine("}")
	} else {
		g.writeLine("")
	}
}

func (g *Generator) generateWhileStatement(stmt *ast.WhileStatement) {
	g.write("for ")
	g.generateExpression(stmt.Condition)
	g.writeLine(" {")
	g.indent++
	if stmt.Body != nil {
		for _, s := range stmt.Body.Statements {
			g.generateStatement(s)
		}
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateForStatement(stmt *ast.ForStatement) {
	op := "<="
	if stmt.DownTo {
		op = ">="
	}

	g.write(fmt.Sprintf("for %s = ", stmt.Variable))
	g.generateExpression(stmt.From)
	g.write(fmt.Sprintf("; %s %s ", stmt.Variable, op))
	g.generateExpression(stmt.To)
	g.write(fmt.Sprintf("; %s", stmt.Variable))
	if stmt.DownTo {
		g.write("--")
	} else {
		g.write("++")
	}
	g.write(" {\n")
	g.indent++
	if stmt.Body != nil {
		for _, s := range stmt.Body.Statements {
			g.generateStatement(s)
		}
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateForEachStatement(stmt *ast.ForEachStatement) {
	g.write(fmt.Sprintf("for _, %s := range ", stmt.Variable))
	g.generateExpression(stmt.Iterable)
	g.writeLine(" {")
	g.indent++
	if stmt.Body != nil {
		for _, s := range stmt.Body.Statements {
			g.generateStatement(s)
		}
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateRepeatStatement(stmt *ast.RepeatStatement) {
	g.writeLine("for {")
	g.indent++
	if stmt.Body != nil {
		for _, s := range stmt.Body.Statements {
			g.generateStatement(s)
		}
	}
	g.write("if ")
	g.generateExpression(stmt.Condition)
	g.writeLine(" {")
	g.indent++
	g.writeLine("break")
	g.indent--
	g.writeLine("}")
	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateCaseStatement(stmt *ast.CaseStatement) {
	g.write("switch ")
	g.generateExpression(stmt.Expression)
	g.writeLine(" {")
	g.indent++
	for _, branch := range stmt.Branches {
		g.write("case ")
		for i, val := range branch.Values {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(val)
		}
		g.writeLine(":")
		g.indent++
		if branch.Body != nil {
			for _, s := range branch.Body.Statements {
				g.generateStatement(s)
			}
		}
		g.indent--
	}
	if stmt.ElseBranch != nil {
		g.writeLine("default:")
		g.indent++
		for _, s := range stmt.ElseBranch.Statements {
			g.generateStatement(s)
		}
		g.indent--
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateMatchStatement(stmt *ast.MatchStatement) {
	// Convert match to switch with guards
	g.write("switch _v := ")
	g.generateExpression(stmt.Expression)
	g.writeLine(" {")
	g.indent++
	for _, branch := range stmt.Branches {
		// Check if this is a wildcard branch (underscore = catch-all)
		wildcard := false
		if ident, ok := branch.Pattern.(*ast.Identifier); ok && ident.Value == "_" {
			wildcard = true
		}

		if wildcard {
			g.writeLine("default:")
		} else if branch.Pattern == nil && branch.When != nil {
			// Guard-only branch: when condition =>
			g.write("case ")
			g.generateExpression(branch.When)
			g.writeLine(":")
		} else {
			g.write("case ")
			// Generate pattern comparisons
			if len(branch.AdditionalPatterns) > 0 {
				// Multi-pattern: _v == p1 || _v == p2 || ...
				for i, p := range append([]ast.Expression{branch.Pattern}, branch.AdditionalPatterns...) {
					if i > 0 {
						g.write(" || ")
					}
					g.write("_v == ")
					g.generateExpression(p)
				}
			} else {
				g.write("_v == ")
				g.generateExpression(branch.Pattern)
			}
			if branch.When != nil {
				g.write(" && ")
				g.generateExpression(branch.When)
			}
			g.writeLine(":")
		}
		g.indent++
		if branch.Body != nil {
			for _, s := range branch.Body.Statements {
				g.generateStatement(s)
			}
		}
		g.indent--
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateTryStatement(stmt *ast.TryStatement) {
	// Go doesn't have try/catch, use defer/recover pattern
	hasOnClauses := len(stmt.OnClauses) > 0

	g.writeLine("func() {")
	g.indent++
	g.writeLine("defer func() {")
	g.indent++
	g.writeLine("if r := recover(); r != nil {")
	g.indent++

	if hasOnClauses {
		g.needsException = true
		// Generate type switch for ON clauses
		g.writeLine("switch e := r.(type) {")
		g.indent++
		for _, on := range stmt.OnClauses {
			g.writeIndent()
			g.write("case ")
			if on.Type != nil {
				// Exception types are panicked as pointers (&Exception{})
				g.write("*")
				g.generateTypeExpression(on.Type)
			} else {
				g.write("interface{}")
			}
			g.writeLine(":")
			g.indent++
			g.inExceptHandler = true
			g.reRaiseVar = "e"
			// Map Pascal ON clause variable to Go type switch variable
			if on.Variable != "" {
				g.nameMap[on.Variable] = "e"
			}
			if on.Body != nil {
				for _, s := range on.Body.Statements {
					g.generateStatement(s)
				}
			}
			// Clear name mapping
			if on.Variable != "" {
				delete(g.nameMap, on.Variable)
			}
			g.inExceptHandler = false
			g.reRaiseVar = ""
			g.indent--
		}
		// Default: re-panic unhandled exceptions
		g.writeLine("default:")
		g.indent++
		g.writeLine("panic(r)")
		g.indent--
		g.indent--
		g.writeLine("}")
	} else if stmt.ExceptBlock != nil {
		// Plain except block (no ON clauses) — handle all exceptions
		g.inExceptHandler = true
		g.reRaiseVar = "r"
		for _, s := range stmt.ExceptBlock.Statements {
			g.generateStatement(s)
		}
		g.inExceptHandler = false
		g.reRaiseVar = ""
	} else {
		// No except handler — just re-panic
		g.writeLine("panic(r)")
	}
	g.indent--
	g.writeLine("}")
	g.indent--
	g.writeLine("}()")

	if stmt.Body != nil {
		for _, s := range stmt.Body.Statements {
			g.generateStatement(s)
		}
	}

	g.indent--
	g.writeLine("}()")

	if stmt.FinallyBlock != nil {
		// Finally block is tricky: wrap remaining code in a defer
		g.writeLine("// finally block")
		for _, s := range stmt.FinallyBlock.Statements {
			g.generateStatement(s)
		}
	}
}

func (g *Generator) generateRaiseStatement(stmt *ast.RaiseStatement) {
	g.needsException = true
	if stmt.Exception != nil {
		g.write("panic(")
		g.generateExpression(stmt.Exception)
		g.write(")\n")
	} else if g.inExceptHandler && g.reRaiseVar != "" {
		// bare raise inside except handler -> re-panic
		g.write(fmt.Sprintf("panic(%s)\n", g.reRaiseVar))
	} else {
		g.write(`panic(&Exception{Message: "exception"})` + "\n")
	}
}

func (g *Generator) generateReturnStatement(stmt *ast.ReturnStatement) {
	g.write("return")
	if stmt.Value != nil {
		// Handle multi-return tuple: return (expr1, expr2)
		if tuple, ok := stmt.Value.(*ast.TupleLiteral); ok {
			g.write(" ")
			for i, elem := range tuple.Elements {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(elem)
			}
		} else {
			g.write(" ")
			g.generateExpression(stmt.Value)
		}
	}
	g.write("\n")
}

func (g *Generator) generateInheritedStatement(stmt *ast.InheritedStatement) {
	if stmt.Expr != nil {
		// inherited MethodName(args) -> call on self (Go embedding handles dispatch)
		g.write("self.")
		g.generateExpression(stmt.Expr)
		g.write("\n")
	} else {
		// bare inherited; -> no-op in Go, rely on embedding
		g.writeLine("// inherited")
	}
}

func (g *Generator) generateExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		// Check name substitution map (for ON clause variables, etc.)
		if mapped, ok := g.nameMap[e.Value]; ok {
			g.write(mapped)
		} else {
			g.write(g.mapBuiltinFunction(e.Value))
		}
	case *ast.IntegerLiteral:
		g.write(fmt.Sprintf("%d", e.Value))
	case *ast.FloatLiteral:
		g.write(fmt.Sprintf("%f", e.Value))
	case *ast.StringLiteral:
		escaped := strings.ReplaceAll(e.Value, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		g.write(fmt.Sprintf(`"%s"`, escaped))
	case *ast.StringInterpolation:
		g.writeInterpolation(e)
	case *ast.BooleanLiteral:
		if e.Value {
			g.write("true")
		} else {
			g.write("false")
		}
	case *ast.NilLiteral:
		g.write("nil")
	case *ast.ArrayLiteral:
		// Empty array literal: use nil (assignable to any slice type)
		if len(e.Elements) == 0 {
			g.write("nil")
		} else {
			g.write("[]interface{}{")
			for i, elem := range e.Elements {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(elem)
			}
			g.write("}")
		}
	case *ast.PrefixExpression:
		op := e.Operator
		if op == "not" {
			op = "!"
		}
		g.write("(")
		g.write(op)
		g.generateExpression(e.Right)
		g.write(")")
	case *ast.InfixExpression:
		g.write("(")
		g.generateExpression(e.Left)
		g.write(" ")
		op := e.Operator
		switch op {
		case "and":
			op = "&&"
		case "or":
			op = "||"
		case "xor":
			op = "^"
		case "div":
			op = "/"
		case "mod":
			op = "%"
		case "<>":
			op = "!="
		case "=":
			op = "=="
		}
		g.write(op)
		g.write(" ")
		g.generateExpression(e.Right)
		g.write(")")
	case *ast.CallExpression:
		// Handle constructor pattern: ClassName.Create(args) → &ClassName{Field: arg, ...}
		if member, ok := e.Function.(*ast.MemberExpression); ok && member.Member == "Create" {
			if ident, ok := member.Object.(*ast.Identifier); ok {
				typeName := ident.Value
				g.write("&")
				g.write(typeName)
				g.write("{")
				fields := g.classFields[typeName]
				for i, arg := range e.Arguments {
					if i > 0 {
						g.write(", ")
					}
					if i < len(fields) {
						g.write(fields[i] + ": ")
					}
					g.generateExpression(arg)
				}
				g.write("}")
				break
			}
		}
		// Handle Ord() — use int(s[0]) with empty string guard
		if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "Ord" && len(e.Arguments) == 1 {
			g.write("func() int { if len(")
			g.generateExpression(e.Arguments[0])
			g.write(") == 0 { return 0 }; return int(")
			g.generateExpression(e.Arguments[0])
			g.write("[0]) }()")
			break
		}
		// Handle Length() — cast to int64
		if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "Length" && len(e.Arguments) == 1 {
			g.write("int64(len(")
			g.generateExpression(e.Arguments[0])
			g.write("))")
			break
		}
		// Handle IntToStr() — use fmt.Sprintf
		if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "IntToStr" && len(e.Arguments) == 1 {
			g.write("fmt.Sprintf(\"%d\", ")
			g.generateExpression(e.Arguments[0])
			g.write(")")
			break
		}
		// Handle StrToInt64()
		if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "StrToInt64" && len(e.Arguments) == 1 {
			g.imports["strconv"] = true
			g.write("func() int64 { v, _ := strconv.ParseInt(")
			g.generateExpression(e.Arguments[0])
			g.write(", 10, 64); return v }()")
			break
		}
		// Handle StrToFloat()
		if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "StrToFloat" && len(e.Arguments) == 1 {
			g.imports["strconv"] = true
			g.write("func() float64 { v, _ := strconv.ParseFloat(")
			g.generateExpression(e.Arguments[0])
			g.write(", 64); return v }()")
			break
		}
			// Handle ReadFile() — reads a file and returns its content as string
		if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "ReadFile" && len(e.Arguments) == 1 {
			g.write("func() string { data, _ := os.ReadFile(")
			g.generateExpression(e.Arguments[0])
			g.write("); return string(data) }()")
			g.imports["os"] = true
			break
		}
		g.generateExpression(e.Function)
		g.write("(")
		for i, arg := range e.Arguments {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(arg)
		}
		g.write(")")
	case *ast.MemberExpression:
		// Handle constructor without parens: ClassName.Create → &ClassName{}
		if e.Member == "Create" {
			if ident, ok := e.Object.(*ast.Identifier); ok {
				g.write("&")
				g.write(ident.Value)
				g.write("{}")
				break
			}
		}
		g.generateExpression(e.Object)
		g.write(".")
		g.write(e.Member)
	case *ast.IndexExpression:
		g.generateExpression(e.Left)
		g.write("[")
		g.generateExpression(e.Index)
		g.write("]")
	case *ast.SliceExpression:
		g.generateExpression(e.Left)
		g.write("[")
		if e.Low != nil {
			g.generateExpression(e.Low)
		}
		g.write(":")
		if e.High != nil {
			g.generateExpression(e.High)
		}
		g.write("]")
	case *ast.LambdaExpression:
		g.write("func(")
		for i, param := range e.Parameters {
			if i > 0 {
				g.write(", ")
			}
			g.write(param.Name)
			if param.Type != nil {
				g.write(" ")
				g.generateTypeExpression(param.Type)
			}
		}
		g.write(") ")
		switch body := e.Body.(type) {
		case *ast.BlockStatement:
			g.writeLine("{")
			g.indent++
			for _, s := range body.Statements {
				g.generateStatement(s)
			}
			g.indent--
			g.write("}")
		case ast.Expression:
			g.writeLine("{")
			g.indent++
			g.write("return ")
			g.generateExpression(body)
			g.write("\n")
			g.indent--
			g.write("}")
		}
	case *ast.AwaitExpression:
		// Go doesn't have await, use goroutines/channels or context
		g.write("<-")
		g.generateExpression(e.Expression)
	case *ast.IsExpression:
		// Type assertion check — use concrete type for class types
		g.write("func() bool { _, ok := ")
		g.generateExpression(e.Expression)
		g.write(".(")
		g.generateTypeExpressionForCast(e.TargetType)
		g.write("); return ok }()")
	case *ast.TypeCastExpression:
		// Type assertion — use concrete type for class types
		g.generateExpression(e.Expression)
		g.write(".(")
		g.generateTypeExpressionForCast(e.TargetType)
		g.write(")")
	case *ast.TupleLiteral:
		// In expression context, a tuple literal doesn't make sense in Go
		// but we need to handle it gracefully (shouldn't normally reach here)
		for i, elem := range e.Elements {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(elem)
		}
	case *ast.MapType:
		// Map type used as a value: generate map[K]V{}
		g.write("map[")
		if e.KeyType != nil {
			g.generateTypeExpression(e.KeyType)
		} else {
			g.write("string")
		}
		g.write("]")
		if e.ValueType != nil {
			g.generateTypeExpression(e.ValueType)
		} else {
			g.write("interface{}")
		}
		g.write("{}")
	}
}
