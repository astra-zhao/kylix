// parser_attribute.go — KylixBoot annotation parsing.
//
// Syntax:
//
//	[Name]                       — marker attribute (e.g. [Inject])
//	[Name(arg1, arg2)]           — attribute with arguments
//	[Name('/path')]              — with string literal
//	[Name(8080)]                 — with int literal
//
// Multiple attributes can stack:
//
//	[Get('/')]
//	[Authenticated]
//	function Foo(): TResponse;
package parser

import (
	"kylix/ast"
	"kylix/token"
)

// parseAttributeList consumes one or more [Attribute] forms at the current
// position. Returns the parsed attributes and leaves the parser positioned at
// the next non-attribute token (the declaration the attributes precede).
func (p *Parser) parseAttributeList() []*ast.Attribute {
	var attrs []*ast.Attribute
	for p.curTokenIs(token.LBRACKET) {
		attr := p.parseAttribute()
		if attr != nil {
			attrs = append(attrs, attr)
		}
		// skip optional semicolons or newlines between attributes
		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}
	return attrs
}

// parseAttribute parses a single [Name] or [Name(args...)] attribute.
func (p *Parser) parseAttribute() *ast.Attribute {
	if !p.curTokenIs(token.LBRACKET) {
		return nil
	}
	openTok := p.curToken
	p.nextToken() // consume '['

	if !p.isIdentOrSoftKeyword() {
		// Not an attribute — could be an array literal or indexing context.
		// Back up — but we already consumed '['; emit an error.
		p.errors = append(p.errors, "expected attribute name after '['")
		return nil
	}

	attr := &ast.Attribute{
		Token: openTok,
		Name:  p.curToken.Literal,
	}
	p.nextToken() // consume name

	// Optional arguments: ( arg1, arg2, ... )
	if p.curTokenIs(token.LPAREN) {
		p.nextToken() // skip '('
		iterations := 0
		for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
			iterations++
			if iterations > 100 {
				p.errors = append(p.errors, "attribute arg list exceeded max iterations")
				break
			}
			before := p.curToken
			expr := p.parseExpression(LOWEST)
			if expr != nil {
				attr.Args = append(attr.Args, expr)
			}
			// Safety: if parseExpression didn't advance, force a move.
			if p.curToken == before {
				p.nextToken()
			}
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			} else if !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
				// Done with args even without explicit comma — break.
				break
			}
		}
		if p.curTokenIs(token.RPAREN) {
			p.nextToken() // skip ')'
		}
	}

	if !p.curTokenIs(token.RBRACKET) {
		p.errors = append(p.errors, "expected ']' to close attribute")
		return nil
	}
	p.nextToken() // skip ']'

	return attr
}
