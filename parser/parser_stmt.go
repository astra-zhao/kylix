// parser_stmt.go — Statement parsing: if, while, for, repeat, case, match, try, raise, return, etc.
package parser

import (
	"fmt"
	"kylix/ast"
	"kylix/token"
)

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.VAR:
		varToken := p.curToken
		p.nextToken()
		decl := p.parseSingleVarDecl(varToken)
		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		return decl
	case token.IF:
		return p.parseIfStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.REPEAT:
		return p.parseRepeatStatement()
	case token.CASE:
		return p.parseCaseStatement()
	case token.MATCH:
		// 'match' can be used as a variable name: match := value; or match: Type;
		if p.peekTokenIs(token.ASSIGN_OP) || p.peekTokenIs(token.COLON) {
			return p.parseExpressionOrAssignment()
		}
		return p.parseMatchStatement()
	case token.TRY:
		return p.parseTryStatement()
	case token.RAISE:
		return p.parseRaiseStatement()
	case token.INHERITED:
		return p.parseInheritedStatement()
	case token.BREAK:
		tok := p.curToken
		p.nextToken()
		return &ast.BreakStatement{Token: tok}
	case token.CONTINUE:
		tok := p.curToken
		p.nextToken()
		return &ast.ContinueStatement{Token: tok}
	case token.RETURN:
		return p.parseReturnStatement()
	case token.EXIT:
		tok := p.curToken
		p.nextToken()
		return &ast.ExpressionStatement{
			Token:      tok,
			Expression: &ast.Identifier{Token: tok, Value: "Exit"},
		}
	default:
		return p.parseExpressionOrAssignment()
	}
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	p.nextToken() // skip 'begin'

	iterations := 0
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		iterations++
		if iterations > 1000 {
			p.errors = append(p.errors, fmt.Sprintf("Block parsing exceeded maximum iterations at token: %s (%s)",
				p.curToken.Literal, p.curToken.Type))
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		} else if !p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.END) {
		p.nextToken()
	}
	return block
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}
	p.nextToken() // skip 'if'

	stmt.Condition = p.parseExpression(LOWEST)
	p.nextToken() // advance past condition

	if p.curTokenIs(token.THEN) {
		p.nextToken()
	}

	if p.curTokenIs(token.BEGIN) {
		stmt.Consequence = p.parseBlockStatement()
	} else {
		s := p.parseStatement()
		stmt.Consequence = &ast.BlockStatement{Statements: []ast.Statement{s}}
	}

	for p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	if p.curTokenIs(token.ELSE) {
		p.nextToken()
		if p.curTokenIs(token.BEGIN) {
			stmt.Alternative = p.parseBlockStatement()
		} else {
			s := p.parseStatement()
			stmt.Alternative = &ast.BlockStatement{Statements: []ast.Statement{s}}
		}
	}
	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}
	p.nextToken() // skip 'while'

	stmt.Condition = p.parseExpression(LOWEST)
	p.nextToken() // advance past condition

	if p.curTokenIs(token.DO) {
		p.nextToken()
	}
	if p.curTokenIs(token.BEGIN) {
		stmt.Body = p.parseBlockStatement()
	} else {
		s := p.parseStatement()
		stmt.Body = &ast.BlockStatement{Statements: []ast.Statement{s}}
	}
	return stmt
}

func (p *Parser) parseForStatement() ast.Statement {
	forToken := p.curToken
	p.nextToken() // skip 'for'

	variable := ""
	if p.curTokenIs(token.IDENT) {
		variable = p.curToken.Literal
		p.nextToken()
	}

	// for x in collection do — foreach variant
	if p.curTokenIs(token.IN) {
		p.nextToken()
		iterable := p.parseExpression(LOWEST)
		p.nextToken()
		if p.curTokenIs(token.DO) {
			p.nextToken()
		}
		var body *ast.BlockStatement
		if p.curTokenIs(token.BEGIN) {
			body = p.parseBlockStatement()
		} else {
			s := p.parseStatement()
			body = &ast.BlockStatement{Statements: []ast.Statement{s}}
		}
		return &ast.ForEachStatement{Token: forToken, Variable: variable, Iterable: iterable, Body: body}
	}

	// for i := start to end do — counted loop
	stmt := &ast.ForStatement{Token: forToken, Variable: variable}
	if p.curTokenIs(token.ASSIGN_OP) {
		p.nextToken()
	}
	stmt.From = p.parseExpression(LOWEST)
	p.nextToken()

	if p.curTokenIs(token.TO) {
		p.nextToken()
		stmt.DownTo = false
	} else if p.curTokenIs(token.DOWNTO) {
		p.nextToken()
		stmt.DownTo = true
	}

	stmt.To = p.parseExpression(LOWEST)
	p.nextToken()

	if p.curTokenIs(token.DO) {
		p.nextToken()
	}
	if p.curTokenIs(token.BEGIN) {
		stmt.Body = p.parseBlockStatement()
	} else {
		s := p.parseStatement()
		stmt.Body = &ast.BlockStatement{Statements: []ast.Statement{s}}
	}
	return stmt
}

func (p *Parser) parseRepeatStatement() *ast.RepeatStatement {
	stmt := &ast.RepeatStatement{Token: p.curToken}
	p.nextToken() // skip 'repeat'

	stmt.Body = &ast.BlockStatement{}
	for !p.curTokenIs(token.UNTIL) && !p.curTokenIs(token.EOF) {
		s := p.parseStatement()
		if s != nil {
			stmt.Body.Statements = append(stmt.Body.Statements, s)
		}
		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.UNTIL) {
		p.nextToken()
		stmt.Condition = p.parseExpression(LOWEST)
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseCaseStatement() *ast.CaseStatement {
	stmt := &ast.CaseStatement{Token: p.curToken}
	p.nextToken() // skip 'case'

	stmt.Expression = p.parseExpression(LOWEST)
	p.nextToken()

	if p.curTokenIs(token.OF) {
		p.nextToken()
	}

	iterations := 0
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		iterations++
		if iterations > 10000 {
			p.errors = append(p.errors, "case statement parsing exceeded maximum iterations")
			break
		}

		branch := &ast.CaseBranch{}
		for {
			val := p.parseExpression(LOWEST)
			if val == nil {
				p.nextToken()
				break
			}
			branch.Values = append(branch.Values, val)
			p.nextToken()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			} else {
				break
			}
		}

		if p.curTokenIs(token.COLON) {
			p.nextToken()
		}
		if p.curTokenIs(token.BEGIN) {
			branch.Body = p.parseBlockStatement()
		} else {
			s := p.parseStatement()
			if s != nil {
				branch.Body = &ast.BlockStatement{Statements: []ast.Statement{s}}
			} else {
				p.nextToken()
			}
		}
		stmt.Branches = append(stmt.Branches, branch)
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.END) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseMatchStatement() *ast.MatchStatement {
	stmt := &ast.MatchStatement{Token: p.curToken}
	p.nextToken() // skip 'match'

	stmt.Expression = p.parseExpression(LOWEST)
	p.nextToken()

	if p.curTokenIs(token.LBRACE) || p.curTokenIs(token.BEGIN) {
		isBrace := p.curTokenIs(token.LBRACE)
		p.nextToken()
		endToken := token.TokenType(token.RBRACE)
		if !isBrace {
			endToken = token.END
		}

		for !p.curTokenIs(endToken) && !p.curTokenIs(token.EOF) {
			branch := &ast.MatchBranch{}

			if p.curTokenIs(token.WHEN) {
				// Guard-only branch: when condition => body
				p.nextToken()
				branch.When = p.parseExpression(LOWEST)
				p.nextToken()
			} else {
				branch.Pattern = p.parseExpression(LOWEST)
				p.nextToken()
				// Multiple patterns: 2, 3 =>
				for p.curTokenIs(token.COMMA) && !p.peekTokenIs(token.FAT_ARROW) {
					p.nextToken()
					branch.AdditionalPatterns = append(branch.AdditionalPatterns, p.parseExpression(LOWEST))
					p.nextToken()
				}
				if p.curTokenIs(token.WHEN) {
					p.nextToken()
					branch.When = p.parseExpression(LOWEST)
					p.nextToken()
				}
			}

			if p.curTokenIs(token.FAT_ARROW) || p.curTokenIs(token.COLON) {
				p.nextToken()
			}
			if p.curTokenIs(token.BEGIN) {
				branch.Body = p.parseBlockStatement()
			} else {
				s := p.parseStatement()
				if s != nil {
					branch.Body = &ast.BlockStatement{Statements: []ast.Statement{s}}
				}
			}
			stmt.Branches = append(stmt.Branches, branch)
			if p.curTokenIs(token.COMMA) || p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
		p.nextToken() // skip closing token
	}
	return stmt
}

func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{Token: p.curToken}
	p.nextToken() // skip 'try'

	if p.curTokenIs(token.BEGIN) {
		stmt.Body = p.parseBlockStatement()
	} else {
		// try body without begin...end
		stmt.Body = &ast.BlockStatement{}
		for !p.curTokenIs(token.EXCEPT) && !p.curTokenIs(token.FINALLY) &&
			!p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
			s := p.parseStatement()
			if s != nil {
				stmt.Body.Statements = append(stmt.Body.Statements, s)
			} else if !p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
			for p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
	}

	if p.curTokenIs(token.EXCEPT) {
		p.nextToken()
		if p.curTokenIs(token.BEGIN) {
			stmt.ExceptBlock = p.parseBlockStatement()
		} else {
			p.parseExceptClauses(stmt)
		}
	}

	if p.curTokenIs(token.FINALLY) {
		p.nextToken()
		if p.curTokenIs(token.BEGIN) {
			stmt.FinallyBlock = p.parseBlockStatement()
		} else {
			stmt.FinallyBlock = &ast.BlockStatement{}
			for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
				s := p.parseStatement()
				if s != nil {
					stmt.FinallyBlock.Statements = append(stmt.FinallyBlock.Statements, s)
				} else if !p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		}
	}

	if p.curTokenIs(token.END) {
		p.nextToken()
	}
	return stmt
}

// parseExceptClauses handles ON clauses and an optional else block inside except.
func (p *Parser) parseExceptClauses(stmt *ast.TryStatement) {
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.FINALLY) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.ON) {
			on := p.parseOnClause()
			if on != nil {
				stmt.OnClauses = append(stmt.OnClauses, on)
			}
		} else if p.curTokenIs(token.ELSE) {
			p.nextToken()
			if stmt.ExceptBlock == nil {
				stmt.ExceptBlock = &ast.BlockStatement{}
			}
			for !p.curTokenIs(token.END) && !p.curTokenIs(token.FINALLY) && !p.curTokenIs(token.EOF) {
				s := p.parseStatement()
				if s != nil {
					stmt.ExceptBlock.Statements = append(stmt.ExceptBlock.Statements, s)
				} else if !p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else {
			// Regular statements act as the except body
			if stmt.ExceptBlock == nil {
				stmt.ExceptBlock = &ast.BlockStatement{}
			}
			s := p.parseStatement()
			if s != nil {
				stmt.ExceptBlock.Statements = append(stmt.ExceptBlock.Statements, s)
			} else if !p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}
}

// parseOnClause parses: on E: ExceptionType do body
func (p *Parser) parseOnClause() *ast.OnClause {
	clause := &ast.OnClause{Token: p.curToken}
	p.nextToken() // skip 'on'

	if p.curTokenIs(token.IDENT) {
		clause.Variable = p.curToken.Literal
		p.nextToken()
	}
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		clause.Type = p.parseTypeExpression()
	}
	if p.curTokenIs(token.DO) {
		p.nextToken()
	}
	if p.curTokenIs(token.BEGIN) {
		clause.Body = p.parseBlockStatement()
	} else {
		s := p.parseStatement()
		if s != nil {
			clause.Body = &ast.BlockStatement{Statements: []ast.Statement{s}}
		}
	}
	return clause
}

func (p *Parser) parseRaiseStatement() *ast.RaiseStatement {
	stmt := &ast.RaiseStatement{Token: p.curToken}
	p.nextToken() // skip 'raise'

	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Exception = p.parseExpression(LOWEST)
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken() // skip 'return'

	if !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.END) {
		stmt.Value = p.parseExpression(LOWEST)
	}
	return stmt
}

// parseInheritedStatement handles: inherited; or inherited MethodName(args);
func (p *Parser) parseInheritedStatement() *ast.InheritedStatement {
	stmt := &ast.InheritedStatement{Token: p.curToken}
	p.nextToken() // skip 'inherited'

	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Expr = p.parseExpression(LOWEST)
	}
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
