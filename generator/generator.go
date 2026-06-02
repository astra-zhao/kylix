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
	inExceptHandler bool   // tracks if we're inside recover() for bare raise
	reRaiseVar      string // variable name holding the recovered value for re-raise
	nameMap         map[string]string // temporary name substitutions (e.g., E→e in on clause)
	imports         map[string]bool // track which imports are needed
}

func New() *Generator {
	return &Generator{
		variables: make(map[string]string),
		nameMap:   make(map[string]string),
		imports:   make(map[string]bool),
	}
}

func (g *Generator) Generate(program *ast.Program) string {
	g.program = program

	// First pass: scan for needed imports
	g.scanImports(program)

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

	// Generate main function if there are top-level statements
	if len(program.Statements) > 0 {
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

	g.write(fmt.Sprintf("type %s ", decl.Name))
	g.generateTypeExpression(decl.Type)
	g.write("\n\n")
}

func (g *Generator) generateClassDecl(decl *ast.ClassDecl) {
	// Generate struct
	g.write("type ")
	g.write(decl.Name)
	g.generateTypeParams(decl.TypeParams)
	g.writeLine(" struct {")
	g.indent++

	// Add parent embedding
	if decl.Parent != "" {
		g.writeLine(decl.Parent)
	}

	// Add fields
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

	// Generate methods
	for _, method := range decl.Methods {
		g.write(fmt.Sprintf("func (self *%s", decl.Name))
		g.generateTypeParams(decl.TypeParams)
		g.write(fmt.Sprintf(") %s", method.Name))
		g.generateFunctionSignature(method)
		g.writeLine(" {")
		g.indent++
		if method.Body != nil {
			g.inFunction = true
			for _, stmt := range method.Body.Statements {
				g.generateStatement(stmt)
			}
			g.inFunction = false
		}
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
	hasReturnType := decl.ReturnType != nil

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
	g.write(fmt.Sprintf("func %s", decl.Name))
	g.generateTypeParams(decl.TypeParams)
	g.generateFunctionSignature(decl)
	g.writeLine(" {")
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

	if hasReturnType {
		g.write("return result\n")
	}

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

	if decl.ReturnType != nil {
		g.write(" ")
		g.generateTypeExpression(decl.ReturnType)
	}
}

func (g *Generator) generateTypeExpression(expr ast.Expression) {
	switch t := expr.(type) {
	case *ast.Identifier:
		g.write(g.mapType(t.Value))
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
	}
}

func (g *Generator) generateStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		g.generateVarDecl(s)
	case *ast.AssignmentStatement:
		g.generateAssignment(s)
	case *ast.ExpressionStatement:
		g.generateExpression(s.Expression)
		g.write("\n")
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

func (g *Generator) generateAssignment(stmt *ast.AssignmentStatement) {
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

	g.write(fmt.Sprintf("for %s := ", stmt.Variable))
	g.generateExpression(stmt.From)
	g.write(fmt.Sprintf("; %s %s ", stmt.Variable, op))
	g.generateExpression(stmt.To)
	g.write(fmt.Sprintf("; %s", stmt.Variable))
	if stmt.DownTo {
		g.write("--")
	} else {
		g.write("++")
	}
	g.writeLine(") {")
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
		} else {
			g.write("case _v == ")
			g.generateExpression(branch.Pattern)
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
		// Generate type switch for ON clauses
		g.writeLine("switch e := r.(type) {")
		g.indent++
		for _, on := range stmt.OnClauses {
			g.writeIndent()
			g.write("case ")
			if on.Type != nil {
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
	if stmt.Exception != nil {
		g.write("panic(")
		g.generateExpression(stmt.Exception)
		g.write(")\n")
	} else if g.inExceptHandler && g.reRaiseVar != "" {
		// bare raise inside except handler -> re-panic
		g.write(fmt.Sprintf("panic(%s)\n", g.reRaiseVar))
	} else {
		g.write(`panic(errors.New("exception"))` + "\n")
		g.imports["errors"] = true
	}
}

func (g *Generator) generateReturnStatement(stmt *ast.ReturnStatement) {
	g.write("return")
	if stmt.Value != nil {
		g.write(" ")
		g.generateExpression(stmt.Value)
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
		g.write(fmt.Sprintf(`"%s"`, e.Value))
	case *ast.BooleanLiteral:
		if e.Value {
			g.write("true")
		} else {
			g.write("false")
		}
	case *ast.NilLiteral:
		g.write("nil")
	case *ast.ArrayLiteral:
		g.write("[]interface{}{")
		for i, elem := range e.Elements {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(elem)
		}
		g.write("}")
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
		// Handle constructor pattern: ClassName.Create(args) → &ClassName{...}
		if member, ok := e.Function.(*ast.MemberExpression); ok && member.Member == "Create" {
			if ident, ok := member.Object.(*ast.Identifier); ok {
				// Constructor call: generate as &TypeName{field: arg}
				g.write("&")
				g.write(ident.Value)
				g.write("{")
				// For constructors with positional args, we can't know field names
				// Use a placeholder approach with positional initialization
				for i, arg := range e.Arguments {
					if i > 0 {
						g.write(", ")
					}
					g.generateExpression(arg)
				}
				g.write("}")
				break
			}
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
		g.generateExpression(e.Object)
		g.write(".")
		g.write(e.Member)
	case *ast.IndexExpression:
		g.generateExpression(e.Left)
		g.write("[")
		g.generateExpression(e.Index)
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
		// Type assertion check
		g.write("func() bool { _, ok := ")
		g.generateExpression(e.Expression)
		g.write(".(")
		g.generateTypeExpression(e.TargetType)
		g.write("); return ok }()")
	case *ast.TypeCastExpression:
		// Type assertion
		g.generateExpression(e.Expression)
		g.write(".(")
		g.generateTypeExpression(e.TargetType)
		g.write(")")
	}
}
