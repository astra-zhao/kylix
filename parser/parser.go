package parser

import (
	"fmt"
	"kylix/ast"
	"kylix/lexer"
	"kylix/token"
	"strconv"
	"strings"
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
	p.registerPrefix(token.SELF, p.parseSelfExpression)
	p.registerPrefix(token.PROCEDURE, p.parseAnonymousFunction)
	p.registerPrefix(token.FUNCTION, p.parseAnonymousFunction)
	p.registerPrefix(token.MATCH, p.parseIdentifier) // 'match' can be used as identifier
	p.registerPrefix(token.MAP, p.parseTypeAsExpression)
	p.registerPrefix(token.VARIANT, p.parseTypeAsExpression)

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

	// Parse unit declaration: unit X;
	if p.curTokenIs(token.UNIT) {
		p.nextToken()
		if p.curTokenIs(token.IDENT) {
			program.UnitName = p.curToken.Literal
			program.IsUnit = true
			program.Name = p.curToken.Literal
			program.NameToken = p.curToken
			p.nextToken()
		}
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

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
			for p.isIdentOrSoftKeyword() || p.curTokenIs(token.LPAREN) {
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
		} else if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) || p.curTokenIs(token.ASYNC) {
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
		// 'match' can be used as a variable name: match := value; or match: Type;
		// Only treat as match statement if followed by an expression and { or begin
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

// parseSingleVarDecl parses a single variable declaration (the 'var' keyword was already consumed by the caller).
// Examples:
//   age: Integer = 25;
//   count := 42;
//   a, b: String;
// isSoftKeyword returns true if the current token is a keyword that can also be used as an identifier
// in certain contexts (variable names, field names, etc.)
func (p *Parser) isSoftKeyword() bool {
	// All keywords can be used as identifiers in member positions (e.g., obj.Default, obj.To)
	// Most Pascal keywords are also valid field/method/property names.
	switch p.curToken.Type {
	case token.MATCH, token.RESULT, token.DEFAULT, token.DOWNTO,
		token.DYNAMIC, token.TO, token.DO, token.OF, token.IN,
		token.READ, token.WRITE, token.ABSTRACT, token.EXTERNAL,
		token.FORWARD, token.VIRTUAL, token.OVERRIDE, token.STATIC,
		token.STORED, token.PACKED, token.FILE, token.NEW, token.DELETE,
		token.EXPORT, token.IMPORT, token.MODULE, token.IS,
		token.EXCEPT, token.ON, token.WHEN:
		return true
	}
	return false
}

// isIdentOrSoftKeyword returns true if the current token is an identifier or a soft keyword
func (p *Parser) isIdentOrSoftKeyword() bool {
	return p.curTokenIs(token.IDENT) || p.isSoftKeyword()
}

// skipNestedDeclaration skips a nested function/procedure declaration
// Used in anonymous functions to skip local function declarations
func (p *Parser) skipNestedDeclaration() {
	p.nextToken() // skip 'function' or 'procedure'
	// Skip name and parameters
	if p.curTokenIs(token.IDENT) {
		p.nextToken()
	}
	if p.curTokenIs(token.LPAREN) {
		depth := 1
		p.nextToken()
		for depth > 0 && !p.curTokenIs(token.EOF) {
			if p.curTokenIs(token.LPAREN) {
				depth++
			} else if p.curTokenIs(token.RPAREN) {
				depth--
			}
			if depth > 0 {
				p.nextToken()
			}
		}
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}
	// Skip return type, semicolons
	for !p.curTokenIs(token.BEGIN) && !p.curTokenIs(token.EOF) {
		p.nextToken()
	}
	if p.curTokenIs(token.BEGIN) {
		_ = p.parseBlockStatement()
	}
}

func (p *Parser) parseSingleVarDecl(varToken token.Token) *ast.VarDecl {
	decl := &ast.VarDecl{
		Token: varToken, // Store the 'var' keyword position
	}

	// Handle destructuring: var (a, b) := expr
	if p.curTokenIs(token.LPAREN) {
		p.nextToken() // skip (
		for p.isIdentOrSoftKeyword() {
			decl.Names = append(decl.Names, p.curToken.Literal)
			p.nextToken()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			} else {
				break
			}
		}
		if p.curTokenIs(token.RPAREN) {
			p.nextToken() // skip )
		}
		if p.curTokenIs(token.ASSIGN_OP) {
			decl.Inferred = true
			p.nextToken()
			decl.Value = p.parseExpression(LOWEST)
			p.nextToken()
		}
		return decl
	}

	// Parse variable names (comma-separated)
	for p.isIdentOrSoftKeyword() {
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
//   TList<T> = class ... end;
func (p *Parser) parseSingleTypeDecl() *ast.TypeDecl {
	decl := &ast.TypeDecl{}

	if p.curTokenIs(token.IDENT) {
		decl.Token = p.curToken // Store identifier position
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	// Parse optional generic type parameters: TList<T> or TMap<K, V>
	var typeParams []*ast.TypeParameter
	if p.curTokenIs(token.LT) {
		typeParams = p.parseTypeParameterList()
	}

	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()

		// Handle class, interface, and record declarations
		if p.curTokenIs(token.CLASS) {
			classDecl := p.parseClassDecl()
			// Name is already set on the type declaration
			classDecl.Name = decl.Name
			classDecl.TypeParams = typeParams
			decl.Type = classDecl
		} else if p.curTokenIs(token.INTERFACE) {
			iface := p.parseInterfaceDecl()
			iface.Name = decl.Name
			decl.Type = iface
		} else {
			decl.Type = p.parseTypeExpression()
		}
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
	isConstructor := p.curTokenIs(token.CONSTRUCTOR)
	isDestructor := p.curTokenIs(token.DESTRUCTOR)
	hasFuncKeyword := !isConstructor && !isDestructor
	p.nextToken() // skip function/procedure/constructor/destructor

	if p.curTokenIs(token.IDENT) || p.isSoftKeyword() {
		decl.Name = p.curToken.Literal
		p.nextToken()

		// Handle method definitions: ClassName.MethodName
		if p.curTokenIs(token.DOT) {
			p.nextToken()
			if p.curTokenIs(token.IDENT) {
				decl.Name = decl.Name + "." + p.curToken.Literal
				p.nextToken()
			}
		}
	}

	// Parse optional generic type parameters: Foo<T> or Swap<T, U>
	if p.curTokenIs(token.LT) {
		decl.TypeParams = p.parseTypeParameterList()
	}

	// Parse parameters
	if p.curTokenIs(token.LPAREN) {
		decl.Parameters = p.parseParameterList()
	}

	// Parse return type (for functions only, not procedures/constructors/destructors)
	if !isProcedure && hasFuncKeyword && p.curTokenIs(token.COLON) {
		p.nextToken()
		// Check for tuple return type: (Type1, Type2, ...)
		if p.curTokenIs(token.LPAREN) {
			p.nextToken()
			var types []ast.Expression
			if !p.curTokenIs(token.RPAREN) {
				types = append(types, p.parseTypeExpression())
				for p.curTokenIs(token.COMMA) {
					p.nextToken()
					types = append(types, p.parseTypeExpression())
				}
			}
			if p.curTokenIs(token.RPAREN) {
				p.nextToken()
			}
			decl.ReturnTypes = types
		} else {
			decl.ReturnType = p.parseTypeExpression()
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// Parse virtual/override/abstract modifiers
	if p.curTokenIs(token.VIRTUAL) || p.curTokenIs(token.OVERRIDE) ||
		p.curTokenIs(token.ABSTRACT) || p.curTokenIs(token.STATIC) ||
		p.curTokenIs(token.DYNAMIC) {
		p.nextToken()
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	// Parse local declarations (var, const) before begin block
	for {
		if p.curTokenIs(token.VAR) {
			varToken := p.curToken
			p.nextToken()
			for p.isIdentOrSoftKeyword() {
				vd := p.parseSingleVarDecl(varToken)
				if vd != nil {
					decl.LocalDecls = append(decl.LocalDecls, vd)
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.CONST) {
			p.nextToken()
			for p.isIdentOrSoftKeyword() {
				cd := p.parseSingleConstDecl()
				if cd != nil {
					decl.LocalDecls = append(decl.LocalDecls, cd)
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else {
			break
		}
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

	iterations := 0
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		iterations++
		if iterations > 1000 {
			p.errors = append(p.errors, "parameter parsing exceeded maximum iterations")
			break
		}

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
		} else if !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
			// Safety: if we're not at a known delimiter or end, advance to avoid infinite loop
			p.nextToken()
		}
	}

	p.nextToken() // skip ')'
	return params
}

// parseTypeParameterList parses <T, U: SomeConstraint, V>
// Called after LT has been consumed; advances to token after GT
func (p *Parser) parseTypeParameterList() []*ast.TypeParameter {
	params := []*ast.TypeParameter{}
	p.nextToken() // skip '<'

	for !p.curTokenIs(token.GT) && !p.curTokenIs(token.EOF) {
		tp := &ast.TypeParameter{}
		if p.curTokenIs(token.IDENT) {
			tp.Token = p.curToken
			tp.Name = p.curToken.Literal
			p.nextToken()
		}

		// Optional constraint: T: SomeType
		if p.curTokenIs(token.COLON) {
			p.nextToken()
			tp.Constraint = p.parseTypeExpression()
		}

		params = append(params, tp)

		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.GT) {
		p.nextToken()
	}

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
	p.nextToken() // advance past condition expression

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
		p.nextToken() // advance past iterable expression
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
	p.nextToken() // advance past From expression

	if p.curTokenIs(token.TO) {
		p.nextToken()
		stmt.DownTo = false
	} else if p.curTokenIs(token.DOWNTO) {
		p.nextToken()
		stmt.DownTo = true
	}

	stmt.To = p.parseExpression(LOWEST)
	p.nextToken() // advance past To expression

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
		// Advance past condition expression
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseCaseStatement() *ast.CaseStatement {
	stmt := &ast.CaseStatement{Token: p.curToken}
	p.nextToken() // skip 'case'

	stmt.Expression = p.parseExpression(LOWEST)
	p.nextToken() // advance past expression

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

		// Parse case values
		for {
			val := p.parseExpression(LOWEST)
			if val == nil {
				p.nextToken() // skip problematic token to avoid infinite loop
				break
			}
			branch.Values = append(branch.Values, val)
			p.nextToken() // advance past the value
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
				p.nextToken() // advance to avoid infinite loop
			}
		}

		stmt.Branches = append(stmt.Branches, branch)

		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.END) {
		p.nextToken() // skip 'end'
	}
	return stmt
}

func (p *Parser) parseMatchStatement() *ast.MatchStatement {
	stmt := &ast.MatchStatement{Token: p.curToken}
	p.nextToken() // skip 'match'

	stmt.Expression = p.parseExpression(LOWEST)
	p.nextToken() // advance past the match expression

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

			// Check if this is a guard-only branch: when condition =>
			if p.curTokenIs(token.WHEN) {
				p.nextToken() // skip 'when'
				branch.When = p.parseExpression(LOWEST)
				p.nextToken() // advance past the guard
			} else {
				// Parse first pattern
				branch.Pattern = p.parseExpression(LOWEST)
				p.nextToken() // advance past the pattern

				// Parse additional patterns: 2, 3 =>
				for p.curTokenIs(token.COMMA) && !p.peekTokenIs(token.FAT_ARROW) {
					p.nextToken() // skip comma
					additionalPattern := p.parseExpression(LOWEST)
					branch.AdditionalPatterns = append(branch.AdditionalPatterns, additionalPattern)
					p.nextToken() // advance past the pattern
				}

				// Optional when guard
				if p.curTokenIs(token.WHEN) {
					p.nextToken()
					branch.When = p.parseExpression(LOWEST)
					p.nextToken() // advance past the guard
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

		p.nextToken() // skip end token
	}

	return stmt
}

func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{Token: p.curToken}
	p.nextToken() // skip 'try'

	if p.curTokenIs(token.BEGIN) {
		stmt.Body = p.parseBlockStatement()
	} else {
		// Try body without begin...end — parse statements until except/finally/end
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

		// Check for begin...end block in except
		if p.curTokenIs(token.BEGIN) {
			stmt.ExceptBlock = p.parseBlockStatement()
		} else {
			// Parse ON clauses and optional else block
			for !p.curTokenIs(token.END) && !p.curTokenIs(token.FINALLY) && !p.curTokenIs(token.EOF) {
				if p.curTokenIs(token.ON) {
					onClause := p.parseOnClause()
					if onClause != nil {
						stmt.OnClauses = append(stmt.OnClauses, onClause)
					}
				} else if p.curTokenIs(token.ELSE) {
					// 'else' clause in except block
					p.nextToken() // skip 'else'
					if stmt.ExceptBlock == nil {
						stmt.ExceptBlock = &ast.BlockStatement{}
					}
					for !p.curTokenIs(token.END) && !p.curTokenIs(token.FINALLY) &&
						!p.curTokenIs(token.EOF) {
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
					// Regular statements = else part of except block
					if stmt.ExceptBlock == nil {
						stmt.ExceptBlock = &ast.BlockStatement{}
					}
					s := p.parseStatement()
					if s != nil {
						stmt.ExceptBlock.Statements = append(stmt.ExceptBlock.Statements, s)
					} else if !p.curTokenIs(token.SEMICOLON) {
						// Skip unknown token to avoid infinite loop
						p.nextToken()
					}
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		}
	}

	if p.curTokenIs(token.FINALLY) {
		p.nextToken()
		if p.curTokenIs(token.BEGIN) {
			stmt.FinallyBlock = p.parseBlockStatement()
		} else {
			// Finally without begin...end — parse statements until end
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

func (p *Parser) parseOnClause() *ast.OnClause {
	clause := &ast.OnClause{Token: p.curToken}
	p.nextToken() // skip 'on'

	// Parse variable name (e.g., "E" in "on E: ExceptionType do")
	if p.curTokenIs(token.IDENT) {
		clause.Variable = p.curToken.Literal
		p.nextToken()
	}

	// Parse : Type
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		clause.Type = p.parseTypeExpression()
	}

	// Parse 'do'
	if p.curTokenIs(token.DO) {
		p.nextToken()
	}

	// Parse body (single statement or block)
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
		// Advance past the expression's last token so ; is at curToken
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

// parseInheritedStatement handles: inherited; or inherited Create(args);
func (p *Parser) parseInheritedStatement() *ast.InheritedStatement {
	stmt := &ast.InheritedStatement{Token: p.curToken}
	p.nextToken() // skip 'inherited'

	// inherits keyword only appears in statement context in Pascal
	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Expr = p.parseExpression(LOWEST)
	}

	// Advance to after the semicolon
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseClassDecl() *ast.ClassDecl {
	decl := &ast.ClassDecl{Visibility: token.PUBLIC}
	p.nextToken() // skip 'class'

	// Only parse name if NOT followed by ':' (which would be a field declaration)
	if p.curTokenIs(token.IDENT) && !p.peekTokenIs(token.COLON) {
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	// Parse optional generic type parameters: class TList<T>
	if p.curTokenIs(token.LT) {
		decl.TypeParams = p.parseTypeParameterList()
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
	} else if p.curTokenIs(token.INHERITS) {
		p.nextToken() // skip 'inherits'
		if p.curTokenIs(token.IDENT) {
			decl.Parent = p.curToken.Literal
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
			for p.isIdentOrSoftKeyword() {
				field := p.parseSingleVarDecl(varToken)
				if field != nil {
					decl.Fields = append(decl.Fields, field)
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) || p.curTokenIs(token.CONSTRUCTOR) || p.curTokenIs(token.DESTRUCTOR) {
			method := p.parseFunctionDecl()
			if method != nil {
				// Mark constructor/destructor
				if p.curTokenIs(token.CONSTRUCTOR) || p.curTokenIs(token.DESTRUCTOR) {
					// Will handle semantic marking in a later pass
				}
				decl.Methods = append(decl.Methods, method)
			}
		} else if p.curTokenIs(token.PROPERTY) {
			prop := p.parsePropertyDecl()
			if prop != nil {
				decl.Properties = append(decl.Properties, prop)
			}
		} else if p.isIdentOrSoftKeyword() && p.peekTokenIs(token.COLON) {
			// Handle bare field declarations without 'var' prefix: name: Type;
			varToken := p.curToken
			field := p.parseSingleVarDecl(varToken)
			if field != nil {
				decl.Fields = append(decl.Fields, field)
			}
			for p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
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
	raw := p.curToken.Literal
	parts := p.parseInterpolatedParts(raw)
	return &ast.StringInterpolation{Parts: parts}
}

// parseInterpolatedParts splits a raw interpolated string like "Hello, ${name}!"
// into alternating text and expression parts.
func (p *Parser) parseInterpolatedParts(raw string) []ast.Expression {
	var parts []ast.Expression
	var currentText strings.Builder

	i := 0
	for i < len(raw) {
		if raw[i] == '$' && i+1 < len(raw) && raw[i+1] == '{' {
			// Flush accumulated text as a string literal
			if currentText.Len() > 0 {
				parts = append(parts, &ast.StringLiteral{Value: currentText.String()})
				currentText.Reset()
			}
			// Find matching closing brace (handles nested braces)
			depth := 1
			j := i + 2
			for j < len(raw) && depth > 0 {
				if raw[j] == '{' {
					depth++
				} else if raw[j] == '}' {
					depth--
				}
				j++
			}
			exprStr := raw[i+2 : j-1] // strip ${ and }
			expr := p.parseExpressionString(strings.TrimSpace(exprStr))
			if expr != nil {
				parts = append(parts, expr)
			}
			i = j
		} else {
			currentText.WriteByte(raw[i])
			i++
		}
	}

	// Flush remaining text
	if currentText.Len() > 0 {
		parts = append(parts, &ast.StringLiteral{Value: currentText.String()})
	}

	return parts
}

// parseExpressionString parses a single expression from a string by creating
// a temporary sub-lexer and sub-parser.
func (p *Parser) parseExpressionString(input string) ast.Expression {
	if input == "" {
		return &ast.StringLiteral{Value: ""}
	}
	l := lexer.New(input)
	sub := New(l)
	expr := sub.parseExpression(LOWEST)
	// Collect any errors from the sub-parser into the main parser
	for _, err := range sub.Errors() {
		p.errors = append(p.errors, "in interpolation '${"+input+"}': "+err)
	}
	return expr
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}

func (p *Parser) parseSelfExpression() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: "self"}
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

	// Check for lambda: () -> body (empty params)
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

	// Check for lambda with parameters: (x: Integer) -> expr
	// Pascal has no syntax like (ident: type) in expression context,
	// so (IDENT : ...) always means lambda parameters
	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COLON) {
		return p.tryParseLambdaParams(openParen)
	}

	// Not a lambda — parse as parenthesized expression or tuple
	exp := p.parseExpression(LOWEST)

	// Check for tuple literal: (expr1, expr2, ...)
	// After parseExpression, peekToken is the next token after the expression.
	// If it's a comma, this is a tuple literal.
	if p.peekTokenIs(token.COMMA) {
		p.nextToken() // advance past last token of expression
		tuple := &ast.TupleLiteral{Token: openParen}
		tuple.Elements = append(tuple.Elements, exp)
		for p.curTokenIs(token.COMMA) {
			p.nextToken() // skip comma
			tuple.Elements = append(tuple.Elements, p.parseExpression(LOWEST))
			// After parseExpression, curToken is last token of expression
			// Advance if more commas follow
			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
			}
		}
		// curToken should now be at RPAREN or we advance
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
		return tuple
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// tryParseLambdaParams attempts to parse a lambda expression with parameters
func (p *Parser) tryParseLambdaParams(openParen token.Token) ast.Expression {
	// Save position for potential rollback
	params := make([]*ast.Parameter, 0)

	// Parse first parameter
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		param := &ast.Parameter{}

		if p.curTokenIs(token.IDENT) {
			param.Token = p.curToken
			param.Name = p.curToken.Literal
			p.nextToken()

			if p.curTokenIs(token.COLON) {
				p.nextToken()
				param.Type = p.parseTypeExpression()
			}

			params = append(params, param)

			// Handle semicolon-separated parameter groups
			if p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
				continue
			}

			if p.curTokenIs(token.COMMA) {
				p.nextToken()
				continue
			}
		} else {
			// Not a valid parameter, give up
			return nil
		}
	}

	// After parsing parameters, we must have RPAREN followed by ARROW
	if !p.curTokenIs(token.RPAREN) {
		return nil
	}
	p.nextToken() // skip ')'

	if !p.curTokenIs(token.ARROW) {
		return nil
	}
	p.nextToken() // skip ->

	// Parse lambda body (expression or block)
	var body ast.Node
	if p.curTokenIs(token.BEGIN) {
		body = p.parseBlockStatement()
	} else {
		body = p.parseExpression(LOWEST)
	}

	return &ast.LambdaExpression{
		Token:      openParen,
		Parameters: params,
		Body:       body,
	}
}

// parseAnonymousFunction parses an anonymous procedure or function expression.
// Syntax: procedure(params); begin ... end
//         function(params): RetType; begin ... end
func (p *Parser) parseAnonymousFunction() ast.Expression {
	kind := p.curToken // 'procedure' or 'function' token
	p.nextToken()      // skip 'procedure'/'function'

	// Parse optional parameter list
	var params []*ast.Parameter
	if p.curTokenIs(token.LPAREN) {
		p.nextToken() // skip '('
		params = make([]*ast.Parameter, 0)
		for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
			param := &ast.Parameter{}

			if p.curTokenIs(token.IDENT) {
				param.Token = p.curToken
				param.Name = p.curToken.Literal
				p.nextToken()

				if p.curTokenIs(token.COLON) {
					p.nextToken()
					param.Type = p.parseTypeExpression()
				}

				params = append(params, param)

				if p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
					continue
				}

				if p.curTokenIs(token.COMMA) {
					p.nextToken()
					continue
				}
			} else {
				p.nextToken()
			}
		}

		if p.curTokenIs(token.RPAREN) {
			p.nextToken() // skip ')'
		}
	}

	// Parse optional return type (for functions only)
	if kind.Type == token.FUNCTION && p.curTokenIs(token.COLON) {
		p.nextToken() // skip ':'
		p.parseTypeExpression()
	}

	// Skip optional semicolon (forward declarations)
	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// Parse local declarations (var, const, type) before begin block
	// Example: procedure() var x: Integer; begin x := 1; end
	for {
		if p.curTokenIs(token.VAR) {
			varToken := p.curToken
			p.nextToken()
			for p.isIdentOrSoftKeyword() {
				p.parseSingleVarDecl(varToken)
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.CONST) {
			p.nextToken()
			for p.isIdentOrSoftKeyword() {
				p.parseSingleConstDecl()
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.TYPE) {
			p.nextToken()
			for p.isIdentOrSoftKeyword() {
				p.parseSingleTypeDecl()
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) {
			// Nested function/procedure - skip for now
			p.skipNestedDeclaration()
		} else {
			break
		}
	}

	// Parse optional body (begin...end block)
	var body ast.Node
	if p.curTokenIs(token.BEGIN) {
		// Don't use parseBlockStatement directly — it consumes 'end' token,
		// which when the anon function is a call argument, eats the outer ')'
		p.nextToken() // skip 'begin'
		block := &ast.BlockStatement{Token: token.Token{Type: token.BEGIN, Literal: "begin"}}
		iter := 0
		for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
			iter++
			if iter > 1000 {
				p.errors = append(p.errors, "anon function body exceeded max iterations")
				break
			}
			s := p.parseStatement()
			if s != nil {
				block.Statements = append(block.Statements, s)
			} else if !p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
			for p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
		if p.curTokenIs(token.END) {
			// Don't skip 'end' — leave it for parseExpressionList's expectPeek
			// When the anon function is a function call argument, the caller
			// needs to see end -> ')' in the standard flow
		}
		body = block
	}

	return &ast.LambdaExpression{
		Token:      kind,
		Parameters: params,
		Body:       body,
	}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	p.nextToken() // Move past [
	for !p.curTokenIs(token.RBRACKET) && !p.curTokenIs(token.EOF) {
		elem := p.parseExpression(LOWEST)
		if elem != nil {
			array.Elements = append(array.Elements, elem)
		}

		// After parseExpression, curToken is at the last token of the expression
		// We need to move past it to see what comes next
		p.nextToken()

		if p.curTokenIs(token.COMMA) {
			p.nextToken() // Move past comma to next element
		}
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

// parseTypeAsExpression parses a type keyword (map, variant) as an expression.
// This handles cases like: var m := map[String]Integer;
func (p *Parser) parseTypeAsExpression() ast.Expression {
	return p.parseTypeExpression()
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken()

	first := p.parseExpression(LOWEST)

	// Check for slice: [a:b]
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // move to ':'
		p.nextToken() // skip ':'
		high := p.parseExpression(LOWEST)
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
		return &ast.SliceExpression{Token: tok, Left: left, Low: first, High: high}
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return &ast.IndexExpression{Token: tok, Left: left, Index: first}
}

// parseMemberExpression parses object.member expressions.
// After '.', members can be identifiers OR keywords used as identifiers
// (e.g., 'default' in property records).
func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	dotToken := p.curToken
	p.nextToken() // skip .
	if p.curTokenIs(token.IDENT) || p.isSoftKeyword() {
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
	// Check for enum type: (ident1, ident2, ...)
	if p.curTokenIs(token.LPAREN) {
		if enumType := p.tryParseEnumType(); enumType != nil {
			return enumType
		}
	}

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

	if p.curTokenIs(token.MAP) {
		p.nextToken()
		mapType := &ast.MapType{}
		if p.curTokenIs(token.LBRACKET) {
			p.nextToken()
			mapType.KeyType = p.parseTypeExpression()
			if p.curTokenIs(token.RBRACKET) {
				p.nextToken()
			}
		}
		mapType.ValueType = p.parseTypeExpression()
		return mapType
	}

	if p.curTokenIs(token.VARIANT) {
		p.nextToken() // skip 'variant'
		variant := &ast.VariantType{}
		for p.isIdentOrSoftKeyword() {
			caseNode := &ast.VariantCase{}
			caseNode.Name = p.curToken.Literal
			p.nextToken() // skip case name
			if p.curTokenIs(token.COLON) {
				p.nextToken() // skip ':'
				caseNode.Type = p.parseTypeExpression()
			}
			variant.Cases = append(variant.Cases, caseNode)
			// Skip semicolons
			for p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
		if p.curTokenIs(token.END) {
			p.nextToken() // skip 'end'
		}
		return variant
	}

	if p.curTokenIs(token.ARRAY) {
		p.nextToken()
		arrayType := &ast.ArrayType{Dynamic: true}

		// Parse size if present
		if p.curTokenIs(token.LBRACKET) {
			p.nextToken()
			lowerBound := p.parseExpression(LOWEST)
			p.nextToken() // advance past the lower bound expression
			arrayType.Dynamic = false
			if p.curTokenIs(token.DOTDOT) {
				p.nextToken()
				upperBound := p.parseExpression(LOWEST) // parse upper bound
				p.nextToken()                            // advance past upper bound
				// Size = upperBound - lowerBound + 1 (Pascal array range semantics)
				arrayType.Size = &ast.InfixExpression{
					Left:     &ast.InfixExpression{Left: upperBound, Operator: "-", Right: lowerBound},
					Operator: "+",
					Right:    &ast.IntegerLiteral{Value: 1},
				}
			} else {
				arrayType.Size = lowerBound
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
		depth := 1 // track nested record depth
		for depth > 0 && !p.curTokenIs(token.EOF) {
			if p.curTokenIs(token.RECORD) {
				depth++
			} else if p.curTokenIs(token.END) {
				depth--
				if depth == 0 {
					p.nextToken() // skip the end that closes this record
					continue
				}
			}
			if depth > 0 {
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
				} else if p.curTokenIs(token.IDENT) {
					// Handle record fields without 'var' keyword
					varToken := p.curToken
					field := p.parseSingleVarDecl(varToken)
					if field != nil {
						record.Fields = append(record.Fields, field)
					}
					for p.curTokenIs(token.SEMICOLON) {
						p.nextToken()
					}
				} else {
					p.nextToken()
				}
			}
		}
		return record
	}

	// Handle function/procedure types: function(ParamTypes): ReturnType
	if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) {
		funcToken := p.curToken
		p.nextToken() // skip 'function'/'procedure'

		funcType := &ast.Identifier{Token: funcToken, Value: funcToken.Literal}

		// Parse optional parameter types
		if p.curTokenIs(token.LPAREN) {
			p.nextToken() // skip '('
			depth := 1
			for depth > 0 && !p.curTokenIs(token.EOF) {
				if p.curTokenIs(token.LPAREN) {
					depth++
				} else if p.curTokenIs(token.RPAREN) {
					depth--
				}
				if depth > 0 {
					p.nextToken()
				}
			}
			if p.curTokenIs(token.RPAREN) {
				p.nextToken() // skip ')'
			}
		}

		// Parse optional return type
		if p.curTokenIs(token.COLON) {
			p.nextToken() // skip ':'
			p.parseTypeExpression() // consume return type
		}

		return funcType
	}

	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// tryParseEnumType attempts to parse an enum type: (ident1, ident2, ...)
func (p *Parser) tryParseEnumType() *ast.EnumType {
	savedCur := p.curToken
	savedPeek := p.peekToken

	p.nextToken() // skip '('

	if !p.curTokenIs(token.IDENT) {
		p.curToken = savedCur
		p.peekToken = savedPeek
		return nil
	}

	if !p.peekTokenIs(token.COMMA) && !p.peekTokenIs(token.RPAREN) {
		p.curToken = savedCur
		p.peekToken = savedPeek
		return nil
	}

	enum := &ast.EnumType{}
	for p.curTokenIs(token.IDENT) {
		enum.Names = append(enum.Names, p.curToken.Literal)
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

	return enum
}
