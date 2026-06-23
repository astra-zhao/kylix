// parser.go — Parser core: struct, constructor, token helpers, and ParseProgram.
package parser

import (
	"fmt"
	"kylix/ast"
	"kylix/lexer"
	"kylix/token"
)

// Operator precedence levels (lowest to highest).
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
	p.registerInfix(token.LT, p.parseLTExpression)
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

	// Read two tokens so curToken and peekToken are both set.
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

// ParseProgram is the entry point. It parses a full Kylix source file
// (unit declaration, uses clause, declarations, and top-level statements).
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

	// Parse program name: program MyApp;
	if p.curTokenIs(token.PROGRAM) {
		p.nextToken()
		if p.curTokenIs(token.IDENT) {
			program.Name = p.curToken.Literal
			program.NameToken = p.curToken
			p.nextToken()
		}
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	// Parse uses clause: uses Web, Orm;
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

	iterations := 0
	for !p.curTokenIs(token.EOF) && !p.curTokenIs(token.END) && !p.curTokenIs(token.DOT) {
		iterations++
		if iterations > 10000 {
			p.errors = append(p.errors, fmt.Sprintf("Parser exceeded maximum iterations at token: %s (%s) line %d",
				p.curToken.Literal, p.curToken.Type, p.curToken.Line))
			break
		}

		if p.curTokenIs(token.VAR) {
			varToken := p.curToken
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
		} else if p.curTokenIs(token.LBRACKET) {
				// Could be an attribute or an array indexing expression.
				// Peek ahead: if followed by IDENT + RBRACKET, it's an attribute.
				attrs := p.parseAttributeList()
				if len(attrs) > 0 {
					// Attribute precedes a declaration. Route to the correct handler.
					if p.curTokenIs(token.TYPE) {
						p.nextToken()
						for p.curTokenIs(token.IDENT) {
							decl := p.parseSingleTypeDecl()
							if decl != nil {
								decl.Attributes = attrs
								program.Declarations = append(program.Declarations, decl)
							}
							for p.curTokenIs(token.SEMICOLON) {
								p.nextToken()
							}
						}
					} else if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) || p.curTokenIs(token.ASYNC) {
						decl := p.parseFunctionDecl()
						if decl != nil {
							decl.Attributes = attrs
							program.Declarations = append(program.Declarations, decl)
						}
					} else if p.curTokenIs(token.CLASS) {
						decl := p.parseClassDecl()
						if decl != nil {
							decl.Attributes = attrs
							program.Declarations = append(program.Declarations, decl)
						}
					} else if p.curTokenIs(token.VAR) {
						varToken := p.curToken
						p.nextToken()
						for p.isIdentOrSoftKeyword() || p.curTokenIs(token.LPAREN) {
							decl := p.parseSingleVarDecl(varToken)
							if decl != nil {
								decl.Attributes = attrs
								program.Declarations = append(program.Declarations, decl)
							}
							for p.curTokenIs(token.SEMICOLON) {
								p.nextToken()
							}
						}
					} else {
						// Fallback: return attributes to the statement expr
						for _, a := range attrs {
							_ = a
						}
					}
				} else {
					stmt := p.parseStatement()
					if stmt != nil {
						program.Statements = append(program.Statements, stmt)
					} else {
						p.nextToken()
					}
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
				p.nextToken()
			}
		}

		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	// Handle trailing END.
	if p.curTokenIs(token.END) {
		p.nextToken()
		if p.curTokenIs(token.DOT) {
			p.nextToken()
		}
	}

	return program
}
