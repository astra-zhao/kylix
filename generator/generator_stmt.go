// generator_stmt.go — Statement code generation.
package generator

import (
	"fmt"
	"kylix/ast"
	"kylix/token"
)

func (g *Generator) generateStatement(stmt ast.Statement) {
	// Emit a //line directive so Go compiler errors point to the Kylix source.
	if tok := stmtToken(stmt); tok.Line > 0 {
		g.writeLineDirective(tok.Line)
	}
	switch s := stmt.(type) {
	case *ast.VarDecl:
		g.generateVarDecl(s)
	case *ast.AssignmentStatement:
		g.generateAssignment(s)
	case *ast.ExpressionStatement:
		g.generateExpressionStatement(s)
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

// generateExpressionStatement handles expression statements, including special
// builtins like append/SetLength that require LHS rewriting.
func (g *Generator) generateExpressionStatement(s *ast.ExpressionStatement) {
	if call, ok := s.Expression.(*ast.CallExpression); ok {
		if ident, ok := call.Function.(*ast.Identifier); ok {
			if ident.Value == "append" && len(call.Arguments) >= 1 {
				// append(arr, elem) → arr = append(arr, elem)
				g.generateExpression(call.Arguments[0])
				g.write(" = append(")
				for i, arg := range call.Arguments {
					if i > 0 {
						g.write(", ")
					}
					g.generateExpression(arg)
				}
				g.write(")\n")
				return
			}
			if ident.Value == "SetLength" && len(call.Arguments) >= 2 {
				// SetLength(arr, n) → arr = __kylixSetLength(arr, int(n))
				// The helper grows (append zeros) or truncates as needed,
				// working correctly for nil/empty/short slices.
				g.needsSetLength = true
				g.generateExpression(call.Arguments[0])
				g.write(" = __kylixSetLength(")
				g.generateExpression(call.Arguments[0])
				g.write(", int(")
				g.generateExpression(call.Arguments[1])
				g.write("))\n")
				return
			}
		}
		g.generateExpression(s.Expression)
		g.write("\n")
		return
	}

	// Exit → return result (in functions) or return
	if ident, ok := s.Expression.(*ast.Identifier); ok && ident.Value == "Exit" {
		if g.inReturnFunc {
			g.write("return result\n")
		} else {
			g.write("return\n")
		}
		return
	}

	// Bare member access used as a procedure call: self.Method → self.Method()
	if _, ok := s.Expression.(*ast.MemberExpression); ok {
		g.generateExpression(s.Expression)
		g.write("()\n")
		return
	}

	g.generateExpression(s.Expression)
	g.write("\n")
}

func (g *Generator) generateVarDecl(decl *ast.VarDecl) {
	// Destructuring: var (a, b) := expr → a, b := expr
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

// generateLocalVarDecl generates a var declaration inside a function body (no initializer).
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

// generateLocalConstDecl generates a const declaration inside a function body.
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
	// Multi-return: result := (expr1, expr2) → return expr1, expr2
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

	// Multi-variable LHS: x, y := Pair() → x, y := Pair()
	if tuple, ok := stmt.Name.(*ast.TupleLiteral); ok {
		for i, elem := range tuple.Elements {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(elem)
		}
		g.write(" := ")
		g.generateExpression(stmt.Value)
		g.write("\n")
		return
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

// generateRepeatStatement generates a repeat...until loop as an infinite for with a break.
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
	// Pascal match → Go: introduce _v as a local, then use a tagless switch
	// with boolean cases. Tagless `switch { case cond: }` supports both
	// value equality (_v == p) and guards (when clause) uniformly.
	g.writeLine("{")
	g.indent++
	g.write("_v := ")
	g.generateExpression(stmt.Expression)
	g.writeLine("")
	g.write("_ = _v")
	g.writeLine("")
	g.writeLine("switch {")
	for _, branch := range stmt.Branches {
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
	g.writeLine("}")
	g.indent--
	g.writeLine("}")
}

// generateTryStatement converts Pascal try/except/finally to Go defer/recover.
func (g *Generator) generateTryStatement(stmt *ast.TryStatement) {
	hasOnClauses := len(stmt.OnClauses) > 0

	g.writeLine("func() {")
	g.indent++
	g.writeLine("defer func() {")
	g.indent++
	g.writeLine("if r := recover(); r != nil {")
	g.indent++

	if hasOnClauses {
		g.needsException = true
		// Type-switch over ON clauses: on E: ExceptionType do ...
		g.writeLine("switch e := r.(type) {")
		g.indent++
		for _, on := range stmt.OnClauses {
			g.writeIndent()
			g.write("case ")
			if on.Type != nil {
				g.write("*")
				g.generateTypeExpression(on.Type)
			} else {
				g.write("interface{}")
			}
			g.writeLine(":")
			g.indent++
			g.inExceptHandler = true
			g.reRaiseVar = "e"
			if on.Variable != "" {
				g.nameMap[on.Variable] = "e"
			}
			if on.Body != nil {
				for _, s := range on.Body.Statements {
					g.generateStatement(s)
				}
			}
			if on.Variable != "" {
				delete(g.nameMap, on.Variable)
			}
			g.inExceptHandler = false
			g.reRaiseVar = ""
			g.indent--
		}
		// Re-panic anything not matched by ON clauses.
		g.writeLine("default:")
		g.indent++
		g.writeLine("panic(r)")
		g.indent--
		g.indent--
		g.writeLine("}")
	} else if stmt.ExceptBlock != nil {
		// Plain except block — handle all exceptions.
		g.inExceptHandler = true
		g.reRaiseVar = "r"
		for _, s := range stmt.ExceptBlock.Statements {
			g.generateStatement(s)
		}
		g.inExceptHandler = false
		g.reRaiseVar = ""
	} else {
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
		// Bare raise inside except handler → re-panic the caught value.
		g.write(fmt.Sprintf("panic(%s)\n", g.reRaiseVar))
	} else {
		g.write(`panic(&Exception{Message: "exception"})` + "\n")
	}
}

func (g *Generator) generateReturnStatement(stmt *ast.ReturnStatement) {
	g.write("return")
	if stmt.Value != nil {
		// Multi-return tuple: return (expr1, expr2) → return expr1, expr2
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
		// inherited MethodName(args) → Go embedding handles dispatch via self.
		g.write("self.")
		g.generateExpression(stmt.Expr)
		g.write("\n")
	} else {
		// Bare inherited; → no-op, rely on Go struct embedding.
		g.writeLine("// inherited")
	}
}

// stmtToken extracts the leading token from a statement for //line directives.
func stmtToken(stmt ast.Statement) token.Token {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		return s.Token
	case *ast.ConstDecl:
		return s.Token
	case *ast.AssignmentStatement:
		return s.Token
	case *ast.ExpressionStatement:
		return s.Token
	case *ast.IfStatement:
		return s.Token
	case *ast.WhileStatement:
		return s.Token
	case *ast.ForStatement:
		return s.Token
	case *ast.ForEachStatement:
		return s.Token
	case *ast.RepeatStatement:
		return s.Token
	case *ast.CaseStatement:
		return s.Token
	case *ast.MatchStatement:
		return s.Token
	case *ast.TryStatement:
		return s.Token
	case *ast.ReturnStatement:
		return s.Token
	case *ast.RaiseStatement:
		return s.Token
	case *ast.BreakStatement:
		return s.Token
	case *ast.ContinueStatement:
		return s.Token
	case *ast.InheritedStatement:
		return s.Token
	case *ast.BlockStatement:
		return s.Token
	}
	return token.Token{}
}
