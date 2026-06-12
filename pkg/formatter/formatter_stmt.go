// formatter_stmt.go — Statement formatting.
package formatter

import "kylix/ast"

func (f *Formatter) formatStatement(stmt interface{}) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		f.formatVarDecl(s)
	case *ast.TypeDecl:
		f.formatTypeDecl(s)
	case *ast.ConstDecl:
		f.formatConstDecl(s)
	case *ast.FunctionDecl:
		f.formatFunctionDecl(s)
	case *ast.ClassDecl:
		f.formatClassDecl(s)
	case *ast.InterfaceDecl:
		f.formatInterfaceDecl(s)
	case *ast.PropertyDecl:
		f.formatPropertyDecl(s)
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
	case *ast.InheritedStatement:
		f.formatInheritedStatement(s)
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
		f.formatExpression(branch.Pattern)
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

	if len(stmt.OnClauses) > 0 {
		f.writeIndent()
		f.write("except\n")
		f.indent++
		for _, on := range stmt.OnClauses {
			f.writeIndent()
			f.write("on " + on.Variable)
			if on.Type != nil {
				f.write(": ")
				f.formatType(on.Type)
			}
			f.write(" do\n")
			f.indent++
			if on.Body != nil {
				f.formatBlock(on.Body)
			}
			f.write(";\n")
			f.indent--
		}
		if stmt.ExceptBlock != nil {
			f.writeIndent()
			f.write("else\n")
			f.indent++
			f.formatBlock(stmt.ExceptBlock)
			f.write(";\n")
			f.indent--
		}
		f.indent--
	} else if stmt.ExceptBlock != nil {
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
	f.write("raise")
	if stmt.Exception != nil {
		f.write(" ")
		f.formatExpression(stmt.Exception)
	}
	f.write(";\n")
}

func (f *Formatter) formatInheritedStatement(stmt *ast.InheritedStatement) {
	f.writeIndent()
	f.write("inherited")
	if stmt.Expr != nil {
		f.write(" ")
		f.formatExpression(stmt.Expr)
	}
	f.write(";\n")
}
