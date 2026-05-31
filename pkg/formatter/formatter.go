package formatter

import (
	"fmt"
	"kylix/ast"
	"strings"
)

// Formatter formats Kylix source code
type Formatter struct {
	indent    int
	indentStr string
	output    strings.Builder
}

// New creates a new formatter with default settings
func New() *Formatter {
	return &Formatter{
		indent:    0,
		indentStr: "  ", // 2 spaces
	}
}

// Format formats an AST and returns formatted source code
func (f *Formatter) Format(program *ast.Program) string {
	f.output.Reset()
	f.indent = 0

	// Program declaration
	if program.Name != "" {
		f.writeLine("program " + program.Name + ";")
		f.writeLine("")
	}

	// Uses clause
	if len(program.Uses) > 0 {
		f.writeLine("uses")
		for i, unit := range program.Uses {
			if i < len(program.Uses)-1 {
				f.writeLine("  " + unit + ",")
			} else {
				f.writeLine("  " + unit + ";")
			}
		}
		f.writeLine("")
	}

	// Declarations
	for i, decl := range program.Declarations {
		f.formatDeclaration(decl)
		if i < len(program.Declarations)-1 {
			f.writeLine("")
		}
	}

	if len(program.Declarations) > 0 && len(program.Statements) > 0 {
		f.writeLine("")
	}

	// Main block
	if len(program.Statements) > 0 {
		f.writeLine("begin")
		f.indent++
		for _, stmt := range program.Statements {
			f.formatStatement(stmt)
		}
		f.indent--
		f.writeLine("end.")
	}

	return f.output.String()
}

func (f *Formatter) write(s string) {
	f.output.WriteString(s)
}

func (f *Formatter) writeLine(s string) {
	for i := 0; i < f.indent; i++ {
		f.output.WriteString(f.indentStr)
	}
	f.output.WriteString(s)
	f.output.WriteString("\n")
}

func (f *Formatter) writeIndent() {
	for i := 0; i < f.indent; i++ {
		f.output.WriteString(f.indentStr)
	}
}

func (f *Formatter) formatDeclaration(decl interface{}) {
	switch d := decl.(type) {
	case *ast.TypeDecl:
		f.formatTypeDecl(d)
	case *ast.VarDecl:
		f.formatVarDecl(d)
	case *ast.ConstDecl:
		f.formatConstDecl(d)
	case *ast.FunctionDecl:
		f.formatFunctionDecl(d)
	case *ast.ClassDecl:
		f.formatClassDecl(d)
	case *ast.InterfaceDecl:
		f.formatInterfaceDecl(d)
	case *ast.PropertyDecl:
		f.formatPropertyDecl(d)
	}
}

func (f *Formatter) formatTypeDecl(decl *ast.TypeDecl) {
	f.writeIndent()
	f.write("type " + decl.Name + " = ")
	f.formatType(decl.Type)
	f.write(";\n")
}

func (f *Formatter) formatVarDecl(decl *ast.VarDecl) {
	if len(decl.Names) == 0 {
		return
	}

	f.writeIndent()
	f.write(strings.Join(decl.Names, ", "))

	if decl.Type != nil {
		f.write(": ")
		f.formatType(decl.Type)
	}

	if decl.Value != nil {
		if decl.Inferred {
			f.write(" := ")
		} else {
			f.write(" = ")
		}
		f.formatExpression(decl.Value)
	}

	f.write(";\n")
}

func (f *Formatter) formatConstDecl(decl *ast.ConstDecl) {
	f.writeIndent()
	f.write(decl.Name + " = ")
	f.formatExpression(decl.Value)
	f.write(";\n")
}

func (f *Formatter) formatFunctionDecl(decl *ast.FunctionDecl) {
	f.writeIndent()

	if decl.IsAsync {
		f.write("async ")
	}

	if decl.ReturnType != nil {
		f.write("function ")
	} else {
		f.write("procedure ")
	}

	f.write(decl.Name)

	// Parameters
	if len(decl.Parameters) > 0 {
		f.write("(")
		for i, param := range decl.Parameters {
			if i > 0 {
				f.write("; ")
			}
			f.write(param.Name + ": ")
			f.formatType(param.Type)
		}
		f.write(")")
	}

	// Return type
	if decl.ReturnType != nil {
		f.write(": ")
		f.formatType(decl.ReturnType)
	}

	f.write(";\n")

	// Body
	if decl.Body != nil {
		f.formatBlock(decl.Body)
	}

	f.write(";\n\n")
}

func (f *Formatter) formatClassDecl(decl *ast.ClassDecl) {
	f.writeIndent()
	f.write("class " + decl.Name)

	if decl.Parent != "" {
		f.write(" inherits " + decl.Parent)
	}

	if len(decl.Interfaces) > 0 {
		f.write(" implements " + strings.Join(decl.Interfaces, ", "))
	}

	f.write("\n")

	f.indent++

	// Fields
	if len(decl.Fields) > 0 {
		f.writeIndent()
		f.write("var\n")
		f.indent++
		for _, field := range decl.Fields {
			f.formatVarDecl(field)
		}
		f.indent--
		f.writeLine("")
	}

	// Methods
	for _, method := range decl.Methods {
		f.formatFunctionDecl(method)
	}

	f.indent--
	f.writeLine("end;")
}

func (f *Formatter) formatInterfaceDecl(decl *ast.InterfaceDecl) {
	f.writeIndent()
	f.write("interface " + decl.Name)

	if len(decl.Parents) > 0 {
		f.write(" inherits " + strings.Join(decl.Parents, ", "))
	}

	f.write("\n")

	f.indent++
	for _, method := range decl.Methods {
		f.formatFunctionDecl(method)
	}
	f.indent--

	f.writeLine("end;")
}

func (f *Formatter) formatPropertyDecl(decl *ast.PropertyDecl) {
	f.writeIndent()
	f.write("property " + decl.Name)

	if decl.Type != nil {
		f.write(": ")
		f.formatType(decl.Type)
	}

	if decl.Getter != "" {
		f.write(" read " + decl.Getter)
	}

	if decl.Setter != "" {
		f.write(" write " + decl.Setter)
	}

	if decl.Default != nil {
		f.write(" default ")
		f.formatExpression(decl.Default)
	}

	f.write(";\n")
}

func (f *Formatter) formatBlock(block *ast.BlockStatement) {
	f.writeIndent()
	f.write("begin\n")
	f.indent++

	for _, stmt := range block.Statements {
		f.formatStatement(stmt)
	}

	f.indent--
	f.writeIndent()
	f.write("end")
}

func (f *Formatter) formatStatement(stmt interface{}) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		f.formatVarDecl(s)
	case *ast.AssignmentStatement:
		f.formatAssignment(s)
	case *ast.IfStatement:
		f.formatIfStatement(s)
	case *ast.WhileStatement:
		f.formatWhileStatement(s)
	case *ast.ForStatement:
		f.formatForStatement(s)
	case *ast.ForEachStatement:
		f.formatForEachStatement(s)
	case *ast.RepeatStatement:
		f.formatRepeatStatement(s)
	case *ast.CaseStatement:
		f.formatCaseStatement(s)
	case *ast.MatchStatement:
		f.formatMatchStatement(s)
	case *ast.TryStatement:
		f.formatTryStatement(s)
	case *ast.RaiseStatement:
		f.formatRaiseStatement(s)
	case *ast.ReturnStatement:
		f.formatReturnStatement(s)
	case *ast.BreakStatement:
		f.writeLine("break;")
	case *ast.ContinueStatement:
		f.writeLine("continue;")
	case *ast.ExpressionStatement:
		f.writeIndent()
		f.formatExpression(s.Expression)
		f.write(";\n")
	case *ast.BlockStatement:
		f.formatBlock(s)
		f.write(";\n")
	}
}

func (f *Formatter) formatAssignment(stmt *ast.AssignmentStatement) {
	f.writeIndent()
	f.formatExpression(stmt.Name)
	f.write(" := ")
	f.formatExpression(stmt.Value)
	f.write(";\n")
}

func (f *Formatter) formatIfStatement(stmt *ast.IfStatement) {
	f.writeIndent()
	f.write("if ")
	f.formatExpression(stmt.Condition)
	f.write(" then\n")

	f.indent++
	f.formatBlock(stmt.Consequence)
	f.write(";\n")
	f.indent--

	if stmt.Alternative != nil {
		f.writeIndent()
		f.write("else\n")
		f.indent++
		f.formatBlock(stmt.Alternative)
		f.write(";\n")
		f.indent--
	}
}

func (f *Formatter) formatWhileStatement(stmt *ast.WhileStatement) {
	f.writeIndent()
	f.write("while ")
	f.formatExpression(stmt.Condition)
	f.write(" do\n")

	f.indent++
	f.formatBlock(stmt.Body)
	f.write(";\n")
	f.indent--
}

func (f *Formatter) formatForStatement(stmt *ast.ForStatement) {
	f.writeIndent()
	f.write("for " + stmt.Variable + " := ")
	f.formatExpression(stmt.From)

	if stmt.DownTo {
		f.write(" downto ")
	} else {
		f.write(" to ")
	}

	f.formatExpression(stmt.To)
	f.write(" do\n")

	f.indent++
	f.formatBlock(stmt.Body)
	f.write(";\n")
	f.indent--
}

func (f *Formatter) formatForEachStatement(stmt *ast.ForEachStatement) {
	f.writeIndent()
	f.write("for " + stmt.Variable + " in ")
	f.formatExpression(stmt.Iterable)
	f.write(" do\n")

	f.indent++
	f.formatBlock(stmt.Body)
	f.write(";\n")
	f.indent--
}

func (f *Formatter) formatRepeatStatement(stmt *ast.RepeatStatement) {
	f.writeIndent()
	f.write("repeat\n")

	f.indent++
	f.formatBlock(stmt.Body)
	f.write(";\n")
	f.indent--

	f.writeIndent()
	f.write("until ")
	f.formatExpression(stmt.Condition)
	f.write(";\n")
}

func (f *Formatter) formatCaseStatement(stmt *ast.CaseStatement) {
	f.writeIndent()
	f.write("case ")
	f.formatExpression(stmt.Expression)
	f.write(" of\n")

	f.indent++
	for _, branch := range stmt.Branches {
		f.writeIndent()
		for i, val := range branch.Values {
			if i > 0 {
				f.write(", ")
			}
			f.formatExpression(val)
		}
		f.write(":\n")

		f.indent++
		f.formatBlock(branch.Body)
		f.write(";\n")
		f.indent--
	}

	if stmt.ElseBranch != nil {
		f.writeIndent()
		f.write("else\n")
		f.indent++
		f.formatBlock(stmt.ElseBranch)
		f.write(";\n")
		f.indent--
	}

	f.indent--
	f.writeIndent()
	f.write("end;\n")
}

func (f *Formatter) formatMatchStatement(stmt *ast.MatchStatement) {
	f.writeIndent()
	f.write("match ")
	f.formatExpression(stmt.Expression)
	f.write("\n")

	f.indent++
	for i, branch := range stmt.Branches {
		f.writeIndent()

		// Format pattern
		f.formatExpression(branch.Pattern)

		// Format guard (when clause)
		if branch.When != nil {
			f.write(" when ")
			f.formatExpression(branch.When)
		}

		f.write(" =>\n")
		f.indent++
		f.formatBlock(branch.Body)
		if i < len(stmt.Branches)-1 {
			f.write(",\n")
		} else {
			f.write("\n")
		}
		f.indent--
	}

	f.indent--
	f.writeIndent()
	f.write("end;\n")
}

func (f *Formatter) formatTryStatement(stmt *ast.TryStatement) {
	f.writeIndent()
	f.write("try\n")
	f.indent++
	f.formatBlock(stmt.Body)
	f.write(";\n")
	f.indent--

	if stmt.ExceptBlock != nil {
		f.writeIndent()
		f.write("except\n")
		f.indent++
		f.formatBlock(stmt.ExceptBlock)
		f.write(";\n")
		f.indent--
	}

	if stmt.FinallyBlock != nil {
		f.writeIndent()
		f.write("finally\n")
		f.indent++
		f.formatBlock(stmt.FinallyBlock)
		f.write(";\n")
		f.indent--
	}

	f.writeIndent()
	f.write("end;\n")
}

func (f *Formatter) formatReturnStatement(stmt *ast.ReturnStatement) {
	f.writeIndent()
	f.write("return")
	if stmt.Value != nil {
		f.write(" ")
		f.formatExpression(stmt.Value)
	}
	f.write(";\n")
}

func (f *Formatter) formatRaiseStatement(stmt *ast.RaiseStatement) {
	f.writeIndent()
	f.write("raise ")
	f.formatExpression(stmt.Exception)
	f.write(";\n")
}

func (f *Formatter) formatType(typ interface{}) {
	switch t := typ.(type) {
	case *ast.Identifier:
		f.write(t.Value)
	case *ast.ArrayType:
		f.write("array")
		if t.Size != nil {
			f.write("[")
			f.formatExpression(t.Size)
			f.write("]")
		}
		f.write(" of ")
		f.formatType(t.ElementType)
	case *ast.RecordType:
		f.write("record\n")
		f.indent++
		for _, field := range t.Fields {
			f.formatVarDecl(field)
		}
		f.indent--
		f.writeIndent()
		f.write("end")
	case *ast.GenericType:
		f.write(t.Base + "<")
		for i, param := range t.TypeParams {
			if i > 0 {
				f.write(", ")
			}
			f.formatType(param)
		}
		f.write(">")
	}
}

func (f *Formatter) formatExpression(expr interface{}) {
	switch e := expr.(type) {
	case *ast.Identifier:
		f.write(e.Value)
	case *ast.IntegerLiteral:
		f.write(fmt.Sprintf("%d", e.Value))
	case *ast.FloatLiteral:
		f.write(fmt.Sprintf("%g", e.Value))
	case *ast.StringLiteral:
		f.write("'" + e.Value + "'")
	case *ast.StringInterpolation:
		f.write("$'")
		for _, part := range e.Parts {
			if strPart, ok := part.(*ast.StringLiteral); ok {
				f.write(strPart.Value)
			} else {
				f.write("{")
				f.formatExpression(part)
				f.write("}")
			}
		}
		f.write("'")
	case *ast.BooleanLiteral:
		if e.Value {
			f.write("true")
		} else {
			f.write("false")
		}
	case *ast.NilLiteral:
		f.write("nil")
	case *ast.ArrayLiteral:
		f.write("[")
		for i, elem := range e.Elements {
			if i > 0 {
				f.write(", ")
			}
			f.formatExpression(elem)
		}
		f.write("]")
	case *ast.PrefixExpression:
		f.write("(" + e.Operator + " ")
		f.formatExpression(e.Right)
		f.write(")")
	case *ast.InfixExpression:
		f.write("(")
		f.formatExpression(e.Left)
		f.write(" " + e.Operator + " ")
		f.formatExpression(e.Right)
		f.write(")")
	case *ast.CallExpression:
		f.formatExpression(e.Function)
		f.write("(")
		for i, arg := range e.Arguments {
			if i > 0 {
				f.write(", ")
			}
			f.formatExpression(arg)
		}
		f.write(")")
	case *ast.MemberExpression:
		f.formatExpression(e.Object)
		f.write("." + e.Member)
	case *ast.IndexExpression:
		f.formatExpression(e.Left)
		f.write("[")
		f.formatExpression(e.Index)
		f.write("]")
	case *ast.LambdaExpression:
		f.write("(")
		for i, param := range e.Parameters {
			if i > 0 {
				f.write("; ")
			}
			f.write(param.Name + ": ")
			f.formatType(param.Type)
		}
		f.write(") -> ")
		// Handle both expression and block body
		switch body := e.Body.(type) {
		case ast.Expression:
			f.formatExpression(body)
		case *ast.BlockStatement:
			f.write("\n")
			f.indent++
			f.formatBlock(body)
			f.indent--
			f.writeIndent()
		}
	case *ast.AwaitExpression:
		f.write("await ")
		f.formatExpression(e.Expression)
	case *ast.IsExpression:
		f.formatExpression(e.Expression)
		f.write(" is ")
		f.formatType(e.TargetType)
	case *ast.TypeCastExpression:
		f.formatExpression(e.Expression)
		f.write(" as ")
		f.formatType(e.TargetType)
	}
}
