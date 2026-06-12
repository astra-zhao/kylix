// formatter.go — Formatter core: struct, Format entry point, declaration formatting.
package formatter

import (
	"kylix/ast"
	"kylix/token"
	"strings"
)

// Operator precedence levels (higher = binds tighter).
const (
	precLowest  = 0
	precOr      = 1 // or, xor
	precAnd     = 2 // and
	precEquals  = 3 // =, <>
	precCompare = 4 // <, >, <=, >=
	precAdd     = 5 // +, -
	precMul     = 6 // *, /, div, mod
	precPrefix  = 7 // not, unary -
)

// operatorPrecedence returns the precedence of a binary operator.
func operatorPrecedence(op string) int {
	switch op {
	case "or", "xor":
		return precOr
	case "and":
		return precAnd
	case "=", "<>":
		return precEquals
	case "<", ">", "<=", ">=":
		return precCompare
	case "+", "-":
		return precAdd
	case "*", "/", "div", "mod":
		return precMul
	default:
		return precLowest
	}
}

// Formatter formats Kylix source code from an AST.
type Formatter struct {
	indent    int
	indentStr string
	output    strings.Builder
}

// New creates a formatter with 2-space indentation.
func New() *Formatter {
	return &Formatter{indentStr: "  "}
}

// Format formats an AST program and returns the formatted source.
func (f *Formatter) Format(program *ast.Program) string {
	f.output.Reset()
	f.indent = 0

	if program.Name != "" {
		f.writeLine("program " + program.Name + ";")
		f.writeLine("")
	}

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

	f.formatDeclarations(program.Declarations)

	if len(program.Declarations) > 0 && len(program.Statements) > 0 {
		f.writeLine("")
	}

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

// formatDeclarations groups consecutive declarations by kind and adds section keywords.
func (f *Formatter) formatDeclarations(decls []ast.Node) {
	if len(decls) == 0 {
		return
	}

	type declSection int
	const (
		sectionNone      declSection = iota
		sectionTypeGroup             // type
		sectionConst                 // const
		sectionVar                   // var
		sectionFunc
		sectionClass
		sectionInterface
		sectionProperty
	)

	currentSection := sectionNone

	for i, decl := range decls {
		var newSection declSection
		switch decl.(type) {
		case *ast.TypeDecl:
			newSection = sectionTypeGroup
		case *ast.ConstDecl:
			newSection = sectionConst
		case *ast.VarDecl:
			newSection = sectionVar
		case *ast.FunctionDecl:
			newSection = sectionFunc
		case *ast.ClassDecl:
			newSection = sectionClass
		case *ast.InterfaceDecl:
			newSection = sectionInterface
		case *ast.PropertyDecl:
			newSection = sectionProperty
		default:
			newSection = sectionNone
		}

		if newSection != currentSection {
			if currentSection != sectionNone {
				f.writeLine("")
			}
			switch newSection {
			case sectionTypeGroup:
				f.writeLine("type")
				f.indent++
			case sectionConst:
				f.writeLine("const")
				f.indent++
			case sectionVar:
				f.writeLine("var")
				f.indent++
			}
			currentSection = newSection
		}

		f.formatDeclaration(decl)

		if newSection == sectionFunc || newSection == sectionClass || newSection == sectionInterface {
			if i < len(decls)-1 {
				f.writeLine("")
			}
		}
	}

	if currentSection == sectionTypeGroup || currentSection == sectionConst || currentSection == sectionVar {
		f.indent--
	}
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
	f.write(decl.Name)
	if decl.Type != nil {
		f.write(": ")
		f.formatType(decl.Type)
	}
	f.write(" = ")
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

	if len(decl.TypeParams) > 0 {
		f.write("<")
		for i, tp := range decl.TypeParams {
			if i > 0 {
				f.write(", ")
			}
			f.write(tp.Name)
			if tp.Constraint != nil {
				f.write(": ")
				f.formatType(tp.Constraint)
			}
		}
		f.write(">")
	}

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

	if decl.ReturnType != nil {
		f.write(": ")
		f.formatType(decl.ReturnType)
	}
	f.write(";\n")

	if decl.Body != nil {
		f.formatBlock(decl.Body)
	}
	f.write(";\n\n")
}

func (f *Formatter) formatClassDecl(decl *ast.ClassDecl) {
	f.writeIndent()
	f.write("class " + decl.Name)

	if len(decl.TypeParams) > 0 {
		f.write("<")
		for i, tp := range decl.TypeParams {
			if i > 0 {
				f.write(", ")
			}
			f.write(tp.Name)
			if tp.Constraint != nil {
				f.write(": ")
				f.formatType(tp.Constraint)
			}
		}
		f.write(">")
	}

	if decl.Parent != "" {
		f.write(" inherits " + decl.Parent)
	}
	if len(decl.Interfaces) > 0 {
		f.write(" implements " + strings.Join(decl.Interfaces, ", "))
	}
	f.write("\n")
	f.indent++

	if decl.Visibility != token.PUBLIC && decl.Visibility != "" {
		f.writeIndent()
		f.write(string(decl.Visibility) + "\n")
	}

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

	for _, prop := range decl.Properties {
		f.formatPropertyDecl(prop)
	}
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
