// parser_expr.go — Expression parsing: literals, operators, calls, lambdas, types.
package parser

import (
	"fmt"
	"kylix/ast"
	"kylix/lexer"
	"kylix/token"
	"strconv"
	"strings"
)

func (p *Parser) parseExpressionOrAssignment() ast.Statement {
	firstToken := p.curToken

	// Multi-variable assignment: x, y := expr  or  x, y := f()
	// Detected when curToken is IDENT and peekToken is COMMA.
	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COMMA) {
		if stmt := p.tryParseMultiAssign(); stmt != nil {
			return stmt
		}
	}

	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}

	// parseExpression leaves curToken at the last token of the expression;
	// check peekToken for assignment operators.
	if p.peekTokenIs(token.ASSIGN) || p.peekTokenIs(token.ASSIGN_OP) {
		p.nextToken()
		assignToken := p.curToken
		p.nextToken()
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
		p.nextToken()
		return &ast.AssignmentStatement{Token: assignToken, Name: expr, Value: value}
	}

	p.nextToken()
	return &ast.ExpressionStatement{Token: firstToken, Expression: expr}
}

// tryParseMultiAssign handles: x, y, z := expr
// Returns nil if the lookahead does not confirm the pattern.
func (p *Parser) tryParseMultiAssign() *ast.AssignmentStatement {
	// Collect identifiers separated by commas.
	var names []string
	names = append(names, p.curToken.Literal)
	firstTok := p.curToken

	// Save position in case we need to back out — use a simple check:
	// after the first ident+comma, we must see ident followed by := or another comma.
	p.nextToken() // skip first ident
	for p.curTokenIs(token.COMMA) {
		p.nextToken() // skip comma
		if !p.curTokenIs(token.IDENT) {
			// Not a multi-assign pattern; can't easily backtrack, so report error.
			return nil
		}
		names = append(names, p.curToken.Literal)
		p.nextToken() // skip ident
	}

	if !p.curTokenIs(token.ASSIGN) && !p.curTokenIs(token.ASSIGN_OP) {
		return nil
	}
	assignTok := p.curToken
	p.nextToken() // skip :=

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}
	p.nextToken() // advance past last token of value

	// Build a TupleLiteral of LHS identifiers as the Name of the assignment.
	lhsTuple := &ast.TupleLiteral{Token: firstTok}
	for _, n := range names {
		lhsTuple.Elements = append(lhsTuple.Elements, &ast.Identifier{Token: firstTok, Value: n})
	}
	return &ast.AssignmentStatement{Token: assignTok, Name: lhsTuple, Value: value}
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
		p.errors = append(p.errors, fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
		return nil
	}
	return &ast.IntegerLiteral{Token: p.curToken, Value: value}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("could not parse %q as float", p.curToken.Literal))
		return nil
	}
	return &ast.FloatLiteral{Token: p.curToken, Value: value}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseStringInterpolation() ast.Expression {
	parts := p.parseInterpolatedParts(p.curToken.Literal)
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
			if currentText.Len() > 0 {
				parts = append(parts, &ast.StringLiteral{Value: currentText.String()})
				currentText.Reset()
			}
			// Find matching closing brace (handles nesting).
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
			exprStr := raw[i+2 : j-1]
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
	if currentText.Len() > 0 {
		parts = append(parts, &ast.StringLiteral{Value: currentText.String()})
	}
	return parts
}

// parseExpressionString creates a temporary sub-parser to parse a single expression
// from a string (used for string interpolation).
func (p *Parser) parseExpressionString(input string) ast.Expression {
	if input == "" {
		return &ast.StringLiteral{Value: ""}
	}
	l := lexer.New(input)
	sub := New(l)
	expr := sub.parseExpression(LOWEST)
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
	expr := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	openParen := p.curToken
	p.nextToken()

	// Empty params: () -> body
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

	// (ident: type, ...) -> lambda params
	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COLON) {
		return p.tryParseLambdaParams(openParen)
	}

	exp := p.parseExpression(LOWEST)

	// Tuple literal: (expr1, expr2, ...)
	if p.peekTokenIs(token.COMMA) {
		p.nextToken()
		tuple := &ast.TupleLiteral{Token: openParen}
		tuple.Elements = append(tuple.Elements, exp)
		for p.curTokenIs(token.COMMA) {
			p.nextToken()
			tuple.Elements = append(tuple.Elements, p.parseExpression(LOWEST))
			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
			}
		}
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

// tryParseLambdaParams parses (x: Type, y: Type) -> expr lambda expressions.
func (p *Parser) tryParseLambdaParams(openParen token.Token) ast.Expression {
	params := make([]*ast.Parameter, 0)

	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		param := &ast.Parameter{}
		if !p.curTokenIs(token.IDENT) {
			return nil
		}
		param.Token = p.curToken
		param.Name = p.curToken.Literal
		p.nextToken()
		if p.curTokenIs(token.COLON) {
			p.nextToken()
			param.Type = p.parseTypeExpression()
		}
		params = append(params, param)
		if p.curTokenIs(token.SEMICOLON) || p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	if !p.curTokenIs(token.RPAREN) {
		return nil
	}
	p.nextToken() // skip ')'

	if !p.curTokenIs(token.ARROW) {
		return nil
	}
	p.nextToken() // skip ->

	var body ast.Node
	if p.curTokenIs(token.BEGIN) {
		body = p.parseBlockStatement()
	} else {
		body = p.parseExpression(LOWEST)
	}

	return &ast.LambdaExpression{Token: openParen, Parameters: params, Body: body}
}

// parseAnonymousFunction parses inline procedure/function expressions.
// Syntax: procedure(params); begin ... end
//
//	function(params): RetType; begin ... end
func (p *Parser) parseAnonymousFunction() ast.Expression {
	kind := p.curToken
	p.nextToken() // skip 'procedure'/'function'

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
				if p.curTokenIs(token.SEMICOLON) || p.curTokenIs(token.COMMA) {
					p.nextToken()
					continue
				}
			} else {
				p.nextToken()
			}
		}
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	// Optional return type for functions.
	if kind.Type == token.FUNCTION && p.curTokenIs(token.COLON) {
		p.nextToken()
		p.parseTypeExpression() // consume but discard in anonymous context
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// Local declarations (var/const/type/nested functions) before begin.
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
			p.skipNestedDeclaration()
		} else {
			break
		}
	}

	var body ast.Node
	if p.curTokenIs(token.BEGIN) {
		// Manual block parsing: don't use parseBlockStatement here because it
		// consumes 'end', which would eat the ')' of the surrounding call when
		// the anonymous function is an argument (e.g., DoWork(procedure() begin ... end)).
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
		// Leave 'end' for the outer parseExpressionList to consume via expectPeek.
		body = block
	}

	return &ast.LambdaExpression{Token: kind, Parameters: params, Body: body}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	p.nextToken() // skip '['

	for !p.curTokenIs(token.RBRACKET) && !p.curTokenIs(token.EOF) {
		elem := p.parseExpression(LOWEST)
		if elem != nil {
			array.Elements = append(array.Elements, elem)
		}
		p.nextToken()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	return array
}

func (p *Parser) parseAwaitExpression() ast.Expression {
	awaitToken := p.curToken
	p.nextToken()
	expr := p.parseExpression(PREFIX)
	return &ast.AwaitExpression{Token: awaitToken, Expression: expr}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// parseTypeAsExpression allows map/variant type keywords to appear in expression position.
// e.g.: var m := map[String]Integer;
func (p *Parser) parseTypeAsExpression() ast.Expression {
	return p.parseTypeExpression()
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken()
	first := p.parseExpression(LOWEST)

	// Slice expression: [a:b]
	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		p.nextToken()
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

// parseMemberExpression parses obj.member. Keywords can appear as member names
// (e.g., obj.default, obj.type) so we accept soft keywords too.
func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	dotToken := p.curToken
	p.nextToken()
	if p.curTokenIs(token.IDENT) || p.isSoftKeyword() {
		return &ast.MemberExpression{Token: dotToken, Object: left, Member: p.curToken.Literal}
	}
	return left
}

func (p *Parser) parseIsExpression(left ast.Expression) ast.Expression {
	isToken := p.curToken
	p.nextToken()
	right := p.parseTypeExpression()
	return &ast.IsExpression{Token: isToken, Expression: left, TargetType: right}
}

func (p *Parser) parseAsExpression(left ast.Expression) ast.Expression {
	asToken := p.curToken
	p.nextToken()
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

// parseTypeExpression parses a type reference: simple name, generic, array, map,
// record, variant, function type, or enum.
func (p *Parser) parseTypeExpression() ast.Expression {
	// Enum: (Ident1, Ident2, ...)
	if p.curTokenIs(token.LPAREN) {
		if enumType := p.tryParseEnumType(); enumType != nil {
			return enumType
		}
	}

	if p.curTokenIs(token.IDENT) {
		name := p.curToken.Literal
		p.nextToken()
		// Generic: Name<T, U>
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
		p.nextToken()
		variant := &ast.VariantType{}
		for p.isIdentOrSoftKeyword() {
			caseNode := &ast.VariantCase{}
			caseNode.Name = p.curToken.Literal
			p.nextToken()
			if p.curTokenIs(token.COLON) {
				p.nextToken()
				caseNode.Type = p.parseTypeExpression()
			}
			variant.Cases = append(variant.Cases, caseNode)
			for p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
		if p.curTokenIs(token.END) {
			p.nextToken()
		}
		return variant
	}

	if p.curTokenIs(token.ARRAY) {
		p.nextToken()
		arrayType := &ast.ArrayType{Dynamic: true}
		if p.curTokenIs(token.LBRACKET) {
			p.nextToken()
			lowerBound := p.parseExpression(LOWEST)
			p.nextToken()
			arrayType.Dynamic = false
			if p.curTokenIs(token.DOTDOT) {
				p.nextToken()
				upperBound := p.parseExpression(LOWEST)
				p.nextToken()
				// Size = upperBound - lowerBound + 1 (Pascal range semantics)
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
		depth := 1
		for depth > 0 && !p.curTokenIs(token.EOF) {
			if p.curTokenIs(token.RECORD) {
				depth++
			} else if p.curTokenIs(token.END) {
				depth--
				if depth == 0 {
					p.nextToken()
					continue
				}
			}
			if depth > 0 {
				if p.curTokenIs(token.VAR) {
					varToken := p.curToken
					p.nextToken()
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

	// Function/procedure type: function(ParamTypes): ReturnType
	if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) {
		funcToken := p.curToken
		p.nextToken()
		funcType := &ast.Identifier{Token: funcToken, Value: funcToken.Literal}
		if p.curTokenIs(token.LPAREN) {
			p.nextToken()
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
				p.nextToken()
			}
		}
		if p.curTokenIs(token.COLON) {
			p.nextToken()
			p.parseTypeExpression()
		}
		return funcType
	}

	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// tryParseEnumType attempts to parse an enum type: (Ident1, Ident2, ...).
// Returns nil and restores state if the token stream doesn't match.
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

// parseLTExpression handles the < token, which is ambiguous:
//   - TBox<Integer>  → generic type instantiation
//   - a < b          → comparison expression
//
// Heuristic: only treat as generic if:
//  1. left is an Identifier whose name starts with an uppercase letter (Pascal type convention)
//  2. peekToken is an Identifier starting with uppercase, or a known type keyword
func (p *Parser) parseLTExpression(left ast.Expression) ast.Expression {
	ident, ok := left.(*ast.Identifier)
	if ok && isTypeIdent(ident.Value) && p.looksLikeGenericArgs() {
		return p.parseGenericInstantiation(ident)
	}
	return p.parseInfixExpression(left)
}

// isTypeIdent returns true if name follows the Pascal type naming convention
// (starts with an uppercase letter, e.g. TBox, IAnimal, Integer).
func isTypeIdent(name string) bool {
	if name == "" {
		return false
	}
	c := name[0]
	return c >= 'A' && c <= 'Z'
}

// looksLikeGenericArgs returns true when the token after < looks like the start
// of a type argument list (uppercase identifier or a type keyword).
func (p *Parser) looksLikeGenericArgs() bool {
	switch p.peekToken.Type {
	case token.ARRAY, token.MAP, token.RECORD:
		return true
	case token.IDENT:
		return isTypeIdent(p.peekToken.Literal)
	}
	return false
}

// parseGenericInstantiation parses  Foo<T1, T2, ...> as a GenericType node.
// curToken is < on entry; returns with curToken on > (not consumed).
func (p *Parser) parseGenericInstantiation(base *ast.Identifier) ast.Expression {
	p.nextToken() // skip <

	gen := &ast.GenericType{Base: base.Value}
	for !p.curTokenIs(token.GT) && !p.curTokenIs(token.EOF) {
		gen.TypeParams = append(gen.TypeParams, p.parseTypeExpression())
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	// Leave curToken on >; Pratt loop will see peekToken (e.g. DOT) and continue.
	return gen
}
