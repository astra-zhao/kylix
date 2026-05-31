package parser

import (
	"fmt"
	"kylix/ast"
	"kylix/lexer"
	"kylix/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
	MEMBER      // object.member
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:   EQUALS,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.LT_EQ:    LESSGREATER,
	token.GT:       LESSGREATER,
	token.GT_EQ:    LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.OR:       SUM,
	token.XOR:      SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.DIV:      PRODUCT,
	token.MOD:      PRODUCT,
	token.AND:      PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.DOT:      MEMBER,
	token.IS:       EQUALS,
	token.AS:       EQUALS,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.RESULT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.STRING_INTERPOLATION, p.parseStringInterpolation)
	p.registerPrefix(token.CHAR, p.parseStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.NIL, p.parseNilLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.AWAIT, p.parseAwaitExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.DIV, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GT_EQ, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.XOR, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)
	p.registerInfix(token.IS, p.parseIsExpression)
	p.registerInfix(token.AS, p.parseAsExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead (line %d, column %d)",
		t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found (line %d, column %d)",
		t, p.curToken.Line, p.curToken.Column)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	// Parse program name if present
	if p.curTokenIs(token.PROGRAM) {
		p.nextToken()
		if p.curTokenIs(token.IDENT) {
			program.Name = p.curToken.Literal
			program.NameToken = p.curToken // Set position
			p.nextToken()
		}
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	// Parse uses clause
	if p.curTokenIs(token.USES) {
		p.nextToken()
		for p.curTokenIs(token.IDENT) {
			program.Uses = append(program.Uses, p.curToken.Literal)
			p.nextToken()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			} else {
				break
			}
		}
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	// Parse declarations and statements
	iterations := 0
	for !p.curTokenIs(token.EOF) && !p.curTokenIs(token.END) && !p.curTokenIs(token.DOT) {
		iterations++
		if iterations > 10000 {
			p.errors = append(p.errors, fmt.Sprintf("Parser exceeded maximum iterations at token: %s (%s) line %d",
				p.curToken.Literal, p.curToken.Type, p.curToken.Line))
			break
		}

		if p.curTokenIs(token.VAR) {
			// var section: consume 'var' then parse all declarations in this section
			varToken := p.curToken // Capture 'var' token
			p.nextToken()
			for p.curTokenIs(token.IDENT) {
				decl := p.parseSingleVarDecl(varToken)
				if decl != nil {
					program.Declarations = append(program.Declarations, decl)
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.CONST) {
			// const section: consume 'const' then parse all declarations
			p.nextToken()
			for p.curTokenIs(token.IDENT) {
				decl := p.parseSingleConstDecl()
				if decl != nil {
					program.Declarations = append(program.Declarations, decl)
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.TYPE) {
			// type section: consume 'type' then parse all declarations
			p.nextToken()
			for p.curTokenIs(token.IDENT) {
				decl := p.parseSingleTypeDecl()
				if decl != nil {
					program.Declarations = append(program.Declarations, decl)
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) {
			decl := p.parseFunctionDecl()
			if decl != nil {
				program.Declarations = append(program.Declarations, decl)
			}
		} else if p.curTokenIs(token.CLASS) {
			decl := p.parseClassDecl()
			if decl != nil {
				program.Declarations = append(program.Declarations, decl)
			}
		} else if p.curTokenIs(token.INTERFACE) {
			decl := p.parseInterfaceDecl()
			if decl != nil {
				program.Declarations = append(program.Declarations, decl)
			}
		} else if p.curTokenIs(token.BEGIN) {
			block := p.parseBlockStatement()
			if block != nil {
				program.Statements = append(program.Statements, block.Statements...)
			}
		} else {
			stmt := p.parseStatement()
			if stmt != nil {
				program.Statements = append(program.Statements, stmt)
			} else {
				// If we couldn't parse a statement, skip the token to avoid infinite loop
				p.nextToken()
			}
		}

		// Skip semicolons
		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	// Handle END. at the end of program
	if p.curTokenIs(token.END) {
		p.nextToken()
		if p.curTokenIs(token.DOT) {
			p.nextToken()
		}
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.VAR:
		varToken := p.curToken
		p.nextToken() // skip 'var'
		decl := p.parseSingleVarDecl(varToken)
		// Skip trailing semicolons
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
		return p.parseMatchStatement()
	case token.TRY:
		return p.parseTryStatement()
	case token.RAISE:
		return p.parseRaiseStatement()
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
	default:
		return p.parseExpressionOrAssignment()
	}
}

// parseSingleVarDecl parses a single variable declaration (the 'var' keyword was already consumed by the caller).
// Examples:
//   age: Integer = 25;
//   count := 42;
//   a, b: String;
func (p *Parser) parseSingleVarDecl(varToken token.Token) *ast.VarDecl {
	decl := &ast.VarDecl{
		Token: varToken, // Store the 'var' keyword position
	}

	// Parse variable names (comma-separated)
	for p.curTokenIs(token.IDENT) {
		decl.Names = append(decl.Names, p.curToken.Literal)
		p.nextToken()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	// Check for type inference :=
	if p.curTokenIs(token.ASSIGN_OP) {
		decl.Inferred = true
		p.nextToken()
		decl.Value = p.parseExpression(LOWEST)
		// parseExpression leaves curToken on the last token of the expression.
		// Advance past it so curToken is at ';' (or next separator) for the caller.
		p.nextToken()
		return decl
	}

	// Parse type annotation
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		decl.Type = p.parseTypeExpression()
		// parseTypeExpression advances past the type identifier, so curToken is
		// at the next token after the type name (e.g. ';' or '=').
	}

	// Parse optional initial value
	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()
		decl.Value = p.parseExpression(LOWEST)
		// Advance past the expression so curToken is at ';' for the caller.
		p.nextToken()
	}

	return decl
}

// parseSingleConstDecl parses a single constant declaration (the 'const' keyword was already consumed).
// Examples:
//   MAX_SIZE = 100;
//   APP_NAME: String = 'Kylix';
func (p *Parser) parseSingleConstDecl() *ast.ConstDecl {
	decl := &ast.ConstDecl{}

	if p.curTokenIs(token.IDENT) {
		decl.Token = p.curToken // Store identifier position
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	// Optional type annotation
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		decl.Type = p.parseTypeExpression()
	}

	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()
		decl.Value = p.parseExpression(LOWEST)
		// Advance past the expression so curToken is at ';' for the caller.
		p.nextToken()
	}

	return decl
}

// parseSingleTypeDecl parses a single type declaration (the 'type' keyword was already consumed).
// Examples:
//   TPoint = record ... end;
//   TIntList = array of Integer;
func (p *Parser) parseSingleTypeDecl() *ast.TypeDecl {
	decl := &ast.TypeDecl{}

	if p.curTokenIs(token.IDENT) {
		decl.Token = p.curToken // Store identifier position
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()
		decl.Type = p.parseTypeExpression()
	}

	return decl
}

func (p *Parser) parseFunctionDecl() *ast.FunctionDecl {
	decl := &ast.FunctionDecl{}

	// Check for async
	if p.curTokenIs(token.ASYNC) {
		decl.IsAsync = true
		p.nextToken()
	}

	decl.Token = p.curToken // Store 'function' or 'procedure' keyword position
	isProcedure := p.curTokenIs(token.PROCEDURE)
	p.nextToken() // skip function/procedure

	if p.curTokenIs(token.IDENT) {
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	// Parse parameters
	if p.curTokenIs(token.LPAREN) {
		decl.Parameters = p.parseParameterList()
	}

	// Parse return type (for functions)
	if !isProcedure && p.curTokenIs(token.COLON) {
		p.nextToken()
		decl.ReturnType = p.parseTypeExpression()
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// Parse body
	if p.curTokenIs(token.BEGIN) {
		decl.Body = p.parseBlockStatement()
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parseParameterList() []*ast.Parameter {
	params := []*ast.Parameter{}
	p.nextToken() // skip '('

	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		param := &ast.Parameter{}

		// Check for var/out parameter
		if p.curTokenIs(token.VAR) {
			p.nextToken()
		}

		if p.curTokenIs(token.IDENT) {
			param.Token = p.curToken // Store parameter name position
			param.Name = p.curToken.Literal
			p.nextToken()
		}

		if p.curTokenIs(token.COLON) {
			p.nextToken()
			param.Type = p.parseTypeExpression()
		}

		params = append(params, param)

		// Handle comma-separated parameters (e.g., a, b: integer)
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			continue
		}

		// Handle semicolon-separated parameter groups
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	p.nextToken() // skip ')'
	return params
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken} // Store 'begin' position
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
			// If we couldn't parse a statement and it's not a semicolon, skip the token
			p.nextToken()
		}

		// Skip semicolons
		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.END) {
		p.nextToken() // skip 'end'
	}
	return block
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken} // Store 'if' position
	p.nextToken() // skip 'if'

	stmt.Condition = p.parseExpression(LOWEST)

	// Advance past the expression to get to 'then'
	p.nextToken()

	if p.curTokenIs(token.THEN) {
		p.nextToken()
	}

	if p.curTokenIs(token.BEGIN) {
		stmt.Consequence = p.parseBlockStatement()
	} else {
		// Single statement
		s := p.parseStatement()
		stmt.Consequence = &ast.BlockStatement{Statements: []ast.Statement{s}}
	}

	// Skip semicolons between consequence and else
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
	stmt := &ast.WhileStatement{Token: p.curToken} // Store 'while' position
	p.nextToken() // skip 'while'

	stmt.Condition = p.parseExpression(LOWEST)

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
	forToken := p.curToken // Capture 'for' before advancing
	p.nextToken() // skip 'for'

	variable := ""
	if p.curTokenIs(token.IDENT) {
		variable = p.curToken.Literal
		p.nextToken()
	}

	// Check for foreach (for x in collection)
	if p.curTokenIs(token.IN) {
		p.nextToken()
		iterable := p.parseExpression(LOWEST)
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

		return &ast.ForEachStatement{
			Token:    forToken,
			Variable: variable,
			Iterable: iterable,
			Body:     body,
		}
	}

	// Regular for loop
	stmt := &ast.ForStatement{Token: forToken, Variable: variable}

	if p.curTokenIs(token.ASSIGN_OP) {
		p.nextToken()
	}

	stmt.From = p.parseExpression(LOWEST)

	if p.curTokenIs(token.TO) {
		p.nextToken()
		stmt.DownTo = false
	} else if p.curTokenIs(token.DOWNTO) {
		p.nextToken()
		stmt.DownTo = true
	}

	stmt.To = p.parseExpression(LOWEST)

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
	}

	return stmt
}

func (p *Parser) parseCaseStatement() *ast.CaseStatement {
	stmt := &ast.CaseStatement{Token: p.curToken}
	p.nextToken() // skip 'case'

	stmt.Expression = p.parseExpression(LOWEST)

	if p.curTokenIs(token.OF) {
		p.nextToken()
	}

	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		branch := &ast.CaseBranch{}

		// Parse case values
		for {
			val := p.parseExpression(LOWEST)
			branch.Values = append(branch.Values, val)
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
			branch.Body = &ast.BlockStatement{Statements: []ast.Statement{s}}
		}

		stmt.Branches = append(stmt.Branches, branch)

		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	p.nextToken() // skip 'end'
	return stmt
}

func (p *Parser) parseMatchStatement() *ast.MatchStatement {
	stmt := &ast.MatchStatement{Token: p.curToken}
	p.nextToken() // skip 'match'

	stmt.Expression = p.parseExpression(LOWEST)

	// Expect { or begin
	if p.curTokenIs(token.LBRACE) || p.curTokenIs(token.BEGIN) {
		isBrace := p.curTokenIs(token.LBRACE)
		p.nextToken()

		endToken := token.TokenType(token.RBRACE)
		if !isBrace {
			endToken = token.END
		}

		for !p.curTokenIs(endToken) && !p.curTokenIs(token.EOF) {
			branch := &ast.MatchBranch{}
			branch.Pattern = p.parseExpression(LOWEST)

			// Optional when guard
			if p.curTokenIs(token.WHEN) {
				p.nextToken()
				branch.When = p.parseExpression(LOWEST)
			}

			if p.curTokenIs(token.FAT_ARROW) || p.curTokenIs(token.COLON) {
				p.nextToken()
			}

			if p.curTokenIs(token.BEGIN) {
				branch.Body = p.parseBlockStatement()
			} else {
				s := p.parseStatement()
				branch.Body = &ast.BlockStatement{Statements: []ast.Statement{s}}
			}

			stmt.Branches = append(stmt.Branches, branch)

			if p.curTokenIs(token.COMMA) || p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}

		p.nextToken() // skip end token
	}

	return stmt
}

func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{Token: p.curToken}
	p.nextToken() // skip 'try'

	if p.curTokenIs(token.BEGIN) {
		stmt.Body = p.parseBlockStatement()
	}

	if p.curTokenIs(token.EXCEPT) {
		p.nextToken()
		if p.curTokenIs(token.BEGIN) {
			stmt.ExceptBlock = p.parseBlockStatement()
		}
	}

	if p.curTokenIs(token.FINALLY) {
		p.nextToken()
		if p.curTokenIs(token.BEGIN) {
			stmt.FinallyBlock = p.parseBlockStatement()
		}
	}

	if p.curTokenIs(token.END) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseRaiseStatement() *ast.RaiseStatement {
	stmt := &ast.RaiseStatement{Token: p.curToken}
	p.nextToken() // skip 'raise'

	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Exception = p.parseExpression(LOWEST)
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

func (p *Parser) parseClassDecl() *ast.ClassDecl {
	decl := &ast.ClassDecl{Visibility: token.PUBLIC}
	p.nextToken() // skip 'class'

	if p.curTokenIs(token.IDENT) {
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	// Parse inheritance
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		if p.curTokenIs(token.IDENT) {
			decl.Parent = p.curToken.Literal
			p.nextToken()
		}
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	// Parse implements
	if p.curTokenIs(token.IMPLEMENTS) {
		p.nextToken()
		for p.curTokenIs(token.IDENT) {
			decl.Interfaces = append(decl.Interfaces, p.curToken.Literal)
			p.nextToken()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			} else {
				break
			}
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// Parse class body
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		// Visibility modifiers
		if p.curTokenIs(token.PUBLIC) || p.curTokenIs(token.PRIVATE) || p.curTokenIs(token.PROTECTED) {
			decl.Visibility = p.curToken.Type
			p.nextToken()
			continue
		}

		if p.curTokenIs(token.VAR) {
			varToken := p.curToken
			p.nextToken() // skip 'var'
			for p.curTokenIs(token.IDENT) {
				field := p.parseSingleVarDecl(varToken)
				if field != nil {
					decl.Fields = append(decl.Fields, field)
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) {
			method := p.parseFunctionDecl()
			if method != nil {
				decl.Methods = append(decl.Methods, method)
			}
		} else if p.curTokenIs(token.PROPERTY) {
			prop := p.parsePropertyDecl()
			if prop != nil {
				decl.Properties = append(decl.Properties, prop)
			}
		} else {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.END) {
		p.nextToken()
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parseInterfaceDecl() *ast.InterfaceDecl {
	decl := &ast.InterfaceDecl{}
	p.nextToken() // skip 'interface'

	if p.curTokenIs(token.IDENT) {
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		for p.curTokenIs(token.IDENT) {
			decl.Parents = append(decl.Parents, p.curToken.Literal)
			p.nextToken()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			} else {
				break
			}
		}
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) {
			method := p.parseFunctionDecl()
			if method != nil {
				decl.Methods = append(decl.Methods, method)
			}
		} else {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.END) {
		p.nextToken()
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parsePropertyDecl() *ast.PropertyDecl {
	decl := &ast.PropertyDecl{}
	p.nextToken() // skip 'property'

	if p.curTokenIs(token.IDENT) {
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	if p.curTokenIs(token.COLON) {
		p.nextToken()
		decl.Type = p.parseTypeExpression()
	}

	// Parse read/write specifiers
	for p.curTokenIs(token.READ) || p.curTokenIs(token.WRITE) || p.curTokenIs(token.DEFAULT) {
		if p.curTokenIs(token.READ) {
			p.nextToken()
			if p.curTokenIs(token.IDENT) {
				decl.Getter = p.curToken.Literal
				p.nextToken()
			}
		} else if p.curTokenIs(token.WRITE) {
			p.nextToken()
			if p.curTokenIs(token.IDENT) {
				decl.Setter = p.curToken.Literal
				p.nextToken()
			}
		} else if p.curTokenIs(token.DEFAULT) {
			p.nextToken()
			decl.Default = p.parseExpression(LOWEST)
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parseExpressionOrAssignment() ast.Statement {
	firstToken := p.curToken
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}

	// After parsing expression, we need to check the next token
	// parseExpression leaves us on the last token of the expression
	// So we need to look at peekToken for the operator
	if p.peekTokenIs(token.ASSIGN) || p.peekTokenIs(token.ASSIGN_OP) {
		p.nextToken() // move to := or =
		assignToken := p.curToken
		p.nextToken() // skip := or =
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
		// Advance past the value expression
		p.nextToken()
		return &ast.AssignmentStatement{
			Token: assignToken,
			Name:  expr,
			Value: value,
		}
	}

	// Advance past the expression for expression statements
	p.nextToken()

	return &ast.ExpressionStatement{Token: firstToken, Expression: expr}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	return &ast.IntegerLiteral{Token: p.curToken, Value: value}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	return &ast.FloatLiteral{Token: p.curToken, Value: value}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseStringInterpolation() ast.Expression {
	// For now, treat it as a regular string literal
	// TODO: Parse the interpolation expressions inside {}
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	openParen := p.curToken
	p.nextToken()

	// Check for lambda: (params) -> body
	if p.curTokenIs(token.RPAREN) && p.peekTokenIs(token.ARROW) {
		p.nextToken() // skip )
		p.nextToken() // skip ->
		body := p.parseExpression(LOWEST)
		return &ast.LambdaExpression{
			Token:      openParen,
			Parameters: []*ast.Parameter{},
			Body:       body,
		}
	}

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	p.nextToken()
	for !p.curTokenIs(token.RBRACKET) && !p.curTokenIs(token.EOF) {
		elem := p.parseExpression(LOWEST)
		array.Elements = append(array.Elements, elem)
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return array
}

func (p *Parser) parseAwaitExpression() ast.Expression {
	awaitToken := p.curToken
	p.nextToken() // skip 'await'
	expr := p.parseExpression(PREFIX)
	return &ast.AwaitExpression{Token: awaitToken, Expression: expr}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	dotToken := p.curToken
	p.nextToken() // skip .
	if p.curTokenIs(token.IDENT) {
		return &ast.MemberExpression{Token: dotToken, Object: left, Member: p.curToken.Literal}
	}
	return left
}

func (p *Parser) parseIsExpression(left ast.Expression) ast.Expression {
	isToken := p.curToken
	p.nextToken() // skip 'is'
	right := p.parseTypeExpression()
	return &ast.IsExpression{Token: isToken, Expression: left, TargetType: right}
}

func (p *Parser) parseAsExpression(left ast.Expression) ast.Expression {
	asToken := p.curToken
	p.nextToken() // skip 'as'
	right := p.parseTypeExpression()
	return &ast.TypeCastExpression{Token: asToken, Expression: left, TargetType: right}
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseTypeExpression() ast.Expression {
	if p.curTokenIs(token.IDENT) {
		name := p.curToken.Literal
		p.nextToken()

		// Check for generic type
		if p.curTokenIs(token.LT) {
			p.nextToken()
			generic := &ast.GenericType{Base: name}
			for !p.curTokenIs(token.GT) && !p.curTokenIs(token.EOF) {
				generic.TypeParams = append(generic.TypeParams, p.parseTypeExpression())
				if p.curTokenIs(token.COMMA) {
					p.nextToken()
				}
			}
			if p.curTokenIs(token.GT) {
				p.nextToken()
			}
			return generic
		}

		return &ast.Identifier{Token: p.curToken, Value: name}
	}

	if p.curTokenIs(token.ARRAY) {
		p.nextToken()
		arrayType := &ast.ArrayType{Dynamic: true}

		// Parse size if present
		if p.curTokenIs(token.LBRACKET) {
			p.nextToken()
			arrayType.Size = p.parseExpression(LOWEST)
			arrayType.Dynamic = false
			if p.curTokenIs(token.DOTDOT) {
				p.nextToken()
				p.parseExpression(LOWEST) // skip upper bound for now
			}
			if p.curTokenIs(token.RBRACKET) {
				p.nextToken()
			}
		}

		if p.curTokenIs(token.OF) {
			p.nextToken()
			arrayType.ElementType = p.parseTypeExpression()
		}

		return arrayType
	}

	if p.curTokenIs(token.RECORD) {
		p.nextToken()
		record := &ast.RecordType{}
		for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
			if p.curTokenIs(token.VAR) {
				varToken := p.curToken
				p.nextToken() // skip 'var'
				for p.curTokenIs(token.IDENT) {
					field := p.parseSingleVarDecl(varToken)
					if field != nil {
						record.Fields = append(record.Fields, field)
					}
					for p.curTokenIs(token.SEMICOLON) {
						p.nextToken()
					}
				}
			} else {
				p.nextToken()
			}
		}
		if p.curTokenIs(token.END) {
			p.nextToken()
		}
		return record
	}

	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}
