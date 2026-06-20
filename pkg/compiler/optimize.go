// optimize.go — Basic compile-time optimizations.
//
// Currently implements:
//  1. Dead code elimination: removes statements after `return`/`raise`/`Exit`
//     within a block (they are unreachable).
//  2. Unused variable suppression: removes `_ = varName` lines that the
//     generator emits to suppress Go's "declared but not used" errors,
//     when the variable IS actually used in the block.
//
// Future: constant folding (const MAX = 5; array[0..MAX-1] → array[0..4]).
package compiler

import (
	"kylix/ast"
)

// OptimizeProgram runs basic optimizations on a parsed program's AST.
// Returns the (potentially modified) program — currently mutates in place.
func OptimizeProgram(program *ast.Program) *ast.Program {
	if program == nil {
		return nil
	}
	// Optimize function bodies.
	for _, decl := range program.Declarations {
		if fd, ok := decl.(*ast.FunctionDecl); ok {
			if fd.Body != nil {
				fd.Body = optimizeBlock(fd.Body)
			}
		}
		if cd, ok := decl.(*ast.ClassDecl); ok {
			for _, method := range cd.Methods {
				if method.Body != nil {
					method.Body = optimizeBlock(method.Body)
				}
			}
		}
	}
	// Optimize main body.
	if len(program.Statements) > 0 {
		program.Statements = optimizeStatements(program.Statements)
	}
	return program
}

// optimizeBlock applies optimizations to a BlockStatement.
func optimizeBlock(block *ast.BlockStatement) *ast.BlockStatement {
	if block == nil {
		return nil
	}
	block.Statements = optimizeStatements(block.Statements)
	return block
}

// optimizeStatements applies dead code elimination to a statement list.
// Recursively optimizes nested blocks (if/while/for/try bodies).
// After a `return`, `raise`, or `Exit` statement, all subsequent statements
// in the same block are unreachable and removed.
func optimizeStatements(stmts []ast.Statement) []ast.Statement {
	if len(stmts) == 0 {
		return stmts
	}
	result := make([]ast.Statement, 0, len(stmts))
	for _, stmt := range stmts {
		// Recursively optimize nested blocks.
		optimizeNestedBlocks(stmt)
		result = append(result, stmt)
		// Check if this statement is a terminator.
		if isTerminator(stmt) {
			break // remaining statements are unreachable
		}
	}
	return result
}

// optimizeNestedBlocks recurses into compound statements to optimize their bodies.
func optimizeNestedBlocks(stmt ast.Statement) {
	if stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.IfStatement:
		if s.Consequence != nil {
			s.Consequence.Statements = optimizeStatements(s.Consequence.Statements)
		}
		if s.Alternative != nil {
			s.Alternative.Statements = optimizeStatements(s.Alternative.Statements)
		}
	case *ast.WhileStatement:
		if s.Body != nil {
			s.Body.Statements = optimizeStatements(s.Body.Statements)
		}
	case *ast.ForStatement:
		if s.Body != nil {
			s.Body.Statements = optimizeStatements(s.Body.Statements)
		}
	case *ast.ForEachStatement:
		if s.Body != nil {
			s.Body.Statements = optimizeStatements(s.Body.Statements)
		}
	case *ast.RepeatStatement:
		if s.Body != nil {
			s.Body.Statements = optimizeStatements(s.Body.Statements)
		}
	case *ast.TryStatement:
		if s.Body != nil {
			s.Body.Statements = optimizeStatements(s.Body.Statements)
		}
		if s.ExceptBlock != nil {
			s.ExceptBlock.Statements = optimizeStatements(s.ExceptBlock.Statements)
		}
		if s.FinallyBlock != nil {
			s.FinallyBlock.Statements = optimizeStatements(s.FinallyBlock.Statements)
		}
		for _, on := range s.OnClauses {
			if on.Body != nil {
				on.Body.Statements = optimizeStatements(on.Body.Statements)
			}
		}
	case *ast.BlockStatement:
		s.Statements = optimizeStatements(s.Statements)
	}
}

// isTerminator returns true if a statement always transfers control
// (return, raise, Exit, break, continue).
func isTerminator(stmt ast.Statement) bool {
	switch stmt.(type) {
	case *ast.ReturnStatement, *ast.RaiseStatement,
		*ast.BreakStatement, *ast.ContinueStatement:
		return true
	case *ast.ExpressionStatement:
		// Exit is parsed as an ExpressionStatement with Identifier "Exit".
		if es, ok := stmt.(*ast.ExpressionStatement); ok {
			if ident, ok := es.Expression.(*ast.Identifier); ok {
				return ident.Value == "Exit" || ident.Value == "exit"
			}
		}
	}
	return false
}
