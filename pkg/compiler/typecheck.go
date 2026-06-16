// typecheck.go — Kylix MVP type checker.
//
// Performs three classes of checks before code generation:
//  1. Undeclared variable/function references
//  2. Function call argument count mismatches
//  3. Obvious type assignment incompatibilities (string literal → Integer var, etc.)
//
// The checker is intentionally conservative: when it cannot determine the type
// with certainty it stays silent rather than producing false positives.
package compiler

import (
	"fmt"
	"kylix/ast"
	"kylix/token"
	"strings"
)

// TypeDiagnostic is a single type-checker finding.
type TypeDiagnostic struct {
	File    string
	Line    int
	Column  int
	Code    string
	Message string
	Hint    string
}

// TypeCheck runs the MVP type checker on a single program and returns any
// findings. sourceFile is used for diagnostic messages only.
func TypeCheck(program *ast.Program, sourceFile string) []TypeDiagnostic {
	c := &checker{
		file:    sourceFile,
		funcs:   make(map[string]*ast.FunctionDecl),
		types:   make(map[string]string), // name → declared type string
		aliases: make(map[string]string), // type alias name → underlying type
	}
	c.collectDeclarations(program)
	c.validateAliases(program, sourceFile)
	c.checkProgram(program)
	return c.diags
}

// ── checker ───────────────────────────────────────────────────────────────────

type checker struct {
	file    string
	diags   []TypeDiagnostic
	funcs   map[string]*ast.FunctionDecl // globally declared functions
	types   map[string]string            // globally declared variable types
	aliases map[string]string            // type alias name → underlying type name
}

func (c *checker) diag(tok token.Token, code string, format string, args ...interface{}) {
	c.diags = append(c.diags, TypeDiagnostic{
		File:    c.file,
		Line:    tok.Line,
		Column:  tok.Column,
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	})
}

// ── pass 1: collect top-level declarations ────────────────────────────────────

func (c *checker) collectDeclarations(prog *ast.Program) {
	for _, decl := range prog.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			c.funcs[d.Name] = d
		case *ast.VarDecl:
			for _, name := range d.Names {
				if d.Type != nil {
					c.types[name] = typeString(d.Type)
				}
			}
		case *ast.TypeDecl:
			// Simple type alias: type UserId = Integer (not class/interface/etc.)
			if ident, ok := d.Type.(*ast.Identifier); ok {
				c.aliases[d.Name] = ident.Value
			}
		}
	}
}

// validateAliases detects recursive/circular type aliases.
func (c *checker) validateAliases(prog *ast.Program, sourceFile string) {
	for _, decl := range prog.Declarations {
		td, ok := decl.(*ast.TypeDecl)
		if !ok {
			continue
		}
		if _, isAlias := c.aliases[td.Name]; !isAlias {
			continue
		}
		// Walk the alias chain; detect cycle
		seen := make(map[string]bool)
		current := td.Name
		for {
			seen[current] = true
			next, ok := c.aliases[current]
			if !ok {
				break // resolved to a non-alias type
			}
			if seen[next] {
				c.diags = append(c.diags, TypeDiagnostic{
					File:    sourceFile,
					Line:    td.Token.Line,
					Column:  td.Token.Column,
					Code:    ErrTypeAliasLoop,
					Message: fmt.Sprintf("type alias '%s' is recursive (cycle detected)", td.Name),
					Hint:    "type aliases cannot reference themselves directly or indirectly",
				})
				break
			}
			current = next
		}
	}
}

// resolveAlias follows the alias chain to find the underlying type name.
// Returns the input unchanged if it is not an alias.
func (c *checker) resolveAlias(typeName string) string {
	seen := make(map[string]bool)
	for {
		seen[typeName] = true
		underlying, ok := c.aliases[typeName]
		if !ok {
			return typeName
		}
		if seen[underlying] {
			return typeName // cycle guard
		}
		typeName = underlying
	}
}

// ── pass 2: check ─────────────────────────────────────────────────────────────

func (c *checker) checkProgram(prog *ast.Program) {
	// Check top-level function bodies
	for _, decl := range prog.Declarations {
		if fd, ok := decl.(*ast.FunctionDecl); ok {
			c.checkFunction(fd)
		}
	}
	// Check main body statements
	if len(prog.Statements) > 0 {
		scope := c.globalScope()
		c.checkStatements(prog.Statements, scope)
	}
}

func (c *checker) checkFunction(fd *ast.FunctionDecl) {
	scope := c.globalScope()
	// Add parameters to scope
	for _, p := range fd.Parameters {
		if p.Type != nil {
			scope[p.Name] = typeString(p.Type)
		} else {
			scope[p.Name] = "?"
		}
	}
	// Add local declarations
	for _, node := range fd.LocalDecls {
		if vd, ok := node.(*ast.VarDecl); ok {
			for _, name := range vd.Names {
				if vd.Type != nil {
					scope[name] = typeString(vd.Type)
				} else {
					scope[name] = "?"
				}
			}
		}
	}
	// Add implicit 'result' variable
	if fd.ReturnType != nil {
		scope["result"] = typeString(fd.ReturnType)
	}
	if fd.Body != nil {
		c.checkStatements(fd.Body.Statements, scope)
	}
}

func (c *checker) globalScope() map[string]string {
	scope := make(map[string]string, len(c.types)+len(c.funcs))
	for k, v := range c.types {
		scope[k] = v
	}
	// Built-in identifiers
	for _, bi := range []string{
		"WriteLn", "Write", "ReadLn", "Read",
		"Length", "SetLength", "Append", "Copy",
		"Ord", "Chr", "Succ", "Pred",
		"Inc", "Dec", "Abs", "Sqr", "Sqrt",
		"IntToStr", "StrToInt", "FloatToStr", "StrToFloat",
		"Trim", "UpperCase", "LowerCase", "Pos",
		"true", "false", "nil", "Self", "self",
		// Pascal implicit return variable — always in scope inside functions
		"result",
	} {
		scope[bi] = "builtin"
	}
	return scope
}

func (c *checker) checkStatements(stmts []ast.Statement, scope map[string]string) {
	for _, stmt := range stmts {
		c.checkStatement(stmt, scope)
	}
}

func (c *checker) checkStatement(stmt ast.Statement, scope map[string]string) {
	if stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.VarDecl:
		for _, name := range s.Names {
			if s.Type != nil {
				scope[name] = typeString(s.Type)
			} else {
				scope[name] = "?"
			}
		}
		// Check obvious type mismatch for initializer
		if s.Value != nil && s.Type != nil {
			c.checkAssignCompat(s.Token, typeString(s.Type), s.Value)
		}

	case *ast.AssignmentStatement:
		// Check undeclared LHS (only for plain identifiers, not qualified names)
		if ident, ok := s.Name.(*ast.Identifier); ok {
			if !c.isDeclared(ident.Value, scope) {
				hint := ""
				if near := NearestName(ident.Value, scopeKeys(scope), 2); near != "" {
					hint = fmt.Sprintf("did you mean '%s'?", near)
				}
				d := TypeDiagnostic{
					File:    c.file,
					Line:    s.Token.Line,
					Column:  s.Token.Column,
					Code:    ErrUndeclared,
					Message: fmt.Sprintf("undeclared variable or function '%s'", ident.Value),
					Hint:    hint,
				}
				c.diags = append(c.diags, d)
			} else if declType, known := scope[ident.Value]; known && declType != "?" && declType != "builtin" {
				c.checkAssignCompat(s.Token, declType, s.Value)
			}
		}
		c.checkExpression(s.Value, scope)

	case *ast.ExpressionStatement:
		c.checkExpression(s.Expression, scope)

	case *ast.ReturnStatement:
		c.checkExpression(s.Value, scope)

	case *ast.IfStatement:
		c.checkExpression(s.Condition, scope)
		if s.Consequence != nil {
			c.checkStatements(s.Consequence.Statements, copyScope(scope))
		}
		if s.Alternative != nil {
			c.checkStatements(s.Alternative.Statements, copyScope(scope))
		}

	case *ast.WhileStatement:
		c.checkExpression(s.Condition, scope)
		if s.Body != nil {
			c.checkStatements(s.Body.Statements, copyScope(scope))
		}

	case *ast.ForStatement:
		scope[s.Variable] = "Integer"
		c.checkExpression(s.From, scope)
		c.checkExpression(s.To, scope)
		if s.Body != nil {
			c.checkStatements(s.Body.Statements, copyScope(scope))
		}

	case *ast.ForEachStatement:
		scope[s.Variable] = "?"
		if s.Body != nil {
			c.checkStatements(s.Body.Statements, copyScope(scope))
		}

	case *ast.TryStatement:
		if s.Body != nil {
			c.checkStatements(s.Body.Statements, copyScope(scope))
		}
		for _, on := range s.OnClauses {
			inner := copyScope(scope)
			if on.Variable != "" {
				inner[on.Variable] = "?"
			}
			if on.Body != nil {
				c.checkStatements(on.Body.Statements, inner)
			}
		}
		if s.ExceptBlock != nil {
			c.checkStatements(s.ExceptBlock.Statements, copyScope(scope))
		}
		if s.FinallyBlock != nil {
			c.checkStatements(s.FinallyBlock.Statements, copyScope(scope))
		}

	case *ast.BlockStatement:
		c.checkStatements(s.Statements, copyScope(scope))
	}
}

func (c *checker) checkExpression(expr ast.Expression, scope map[string]string) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.CallExpression:
		c.checkCall(e, scope)
	case *ast.InfixExpression:
		c.checkExpression(e.Left, scope)
		c.checkExpression(e.Right, scope)
	case *ast.PrefixExpression:
		c.checkExpression(e.Right, scope)
	case *ast.IndexExpression:
		c.checkExpression(e.Left, scope)
		c.checkExpression(e.Index, scope)
	case *ast.TupleLiteral:
		for _, el := range e.Elements {
			c.checkExpression(el, scope)
		}
	}
}

func (c *checker) checkCall(call *ast.CallExpression, scope map[string]string) {
	// Resolve function name: direct identifier or module.func
	name := ""
	switch fn := call.Function.(type) {
	case *ast.Identifier:
		name = fn.Value
	case *ast.MemberExpression:
		// module.func — skip arity check (external)
		c.checkExpression(fn.Object, scope)
		for _, arg := range call.Arguments {
			c.checkExpression(arg, scope)
		}
		return
	}

	if name == "" {
		return
	}

	// Check arguments
	for _, arg := range call.Arguments {
		c.checkExpression(arg, scope)
	}

	// Arity check against known local functions only
	fd, known := c.funcs[name]
	if !known {
		return // builtin or cross-unit — skip
	}

	// Count required parameters (Pascal passes multiple names per type group)
	required := len(fd.Parameters)
	got := len(call.Arguments)
	if got != required {
		c.diag(call.Token, ErrWrongArity,
			"wrong number of arguments to '%s': expected %d, got %d",
			name, required, got)
	}
}

// checkAssignCompat reports an error when assigning a literal of the obviously
// wrong kind to a typed variable. Only flags the cases we can prove wrong
// without full type inference.
func (c *checker) checkAssignCompat(tok token.Token, declaredType string, value ast.Expression) {
	if value == nil {
		return
	}
	// Resolve alias to underlying type before comparing
	resolved := c.resolveAlias(declaredType)
	norm := strings.ToLower(resolved)

	switch v := value.(type) {
	case *ast.StringLiteral:
		if norm == "integer" || norm == "int64" || norm == "real" || norm == "double" || norm == "boolean" {
			hint := typeConversionHint(declaredType, "string")
			d := TypeDiagnostic{
				File:    c.file,
				Line:    tok.Line,
				Column:  tok.Column,
				Code:    ErrTypeMismatch,
				Message: fmt.Sprintf("cannot assign String literal to variable of type '%s'", declaredType),
				Hint:    hint,
			}
			c.diags = append(c.diags, d)
		}
		_ = v
	case *ast.IntegerLiteral:
		if norm == "string" {
			hint := typeConversionHint(declaredType, "integer")
			d := TypeDiagnostic{
				File:    c.file,
				Line:    tok.Line,
				Column:  tok.Column,
				Code:    ErrTypeMismatch,
				Message: "cannot assign Integer literal to variable of type 'String'",
				Hint:    hint,
			}
			c.diags = append(c.diags, d)
		}
	case *ast.BooleanLiteral:
		if norm == "integer" || norm == "int64" || norm == "real" || norm == "double" || norm == "string" {
			c.diag(tok, ErrTypeMismatch, "cannot assign Boolean literal to variable of type '%s'", declaredType)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (c *checker) isDeclared(name string, scope map[string]string) bool {
	if _, ok := scope[name]; ok {
		return true
	}
	if _, ok := c.funcs[name]; ok {
		return true
	}
	return false
}

func scopeKeys(scope map[string]string) []string {
	keys := make([]string, 0, len(scope))
	for k := range scope {
		if k != "builtin" {
			keys = append(keys, k)
		}
	}
	return keys
}

// typeString converts a type expression to a normalized string.
func typeString(t ast.Expression) string {
	if t == nil {
		return "?"
	}
	switch v := t.(type) {
	case *ast.Identifier:
		return v.Value
	case *ast.ArrayType:
		return "array"
	case *ast.MapType:
		return "map"
	default:
		return fmt.Sprintf("%T", t)
	}
}

func copyScope(s map[string]string) map[string]string {
	out := make(map[string]string, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}
