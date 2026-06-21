// parser_decl.go — Declaration parsing: var, const, type, function, class, interface, property.
package parser

import (
	"kylix/ast"
	"kylix/token"
)

// isSoftKeyword returns true for keywords that are also valid identifiers
// in certain positions (field names, variable names, etc.).
func (p *Parser) isSoftKeyword() bool {
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

func (p *Parser) isIdentOrSoftKeyword() bool {
	return p.curTokenIs(token.IDENT) || p.isSoftKeyword()
}

// skipNestedDeclaration skips a nested function/procedure body.
// Used inside anonymous functions to avoid re-parsing local declarations.
func (p *Parser) skipNestedDeclaration() {
	p.nextToken() // skip 'function' or 'procedure'
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
	for !p.curTokenIs(token.BEGIN) && !p.curTokenIs(token.EOF) {
		p.nextToken()
	}
	if p.curTokenIs(token.BEGIN) {
		_ = p.parseBlockStatement()
	}
}

// parseSingleVarDecl parses one variable declaration after the 'var' keyword has been consumed.
// Supports: name: Type, name := expr, (a, b) := expr, a, b: Type = expr
func (p *Parser) parseSingleVarDecl(varToken token.Token) *ast.VarDecl {
	decl := &ast.VarDecl{Token: varToken}

	// Destructuring: var (a, b) := expr
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
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
			p.nextToken()
		}
		if p.curTokenIs(token.ASSIGN_OP) {
			decl.Inferred = true
			p.nextToken()
			decl.Value = p.parseExpression(LOWEST)
			p.nextToken()
		}
		return decl
	}

	// Comma-separated names: a, b, c
	for p.isIdentOrSoftKeyword() {
		decl.Names = append(decl.Names, p.curToken.Literal)
		p.nextToken()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	// Type inference: :=
	if p.curTokenIs(token.ASSIGN_OP) {
		decl.Inferred = true
		p.nextToken()
		decl.Value = p.parseExpression(LOWEST)
		p.nextToken()
		return decl
	}

	// Explicit type annotation: : Type
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		decl.Type = p.parseTypeExpression()
	}

	// Optional initializer: = expr
	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()
		decl.Value = p.parseExpression(LOWEST)
		p.nextToken()
	}

	return decl
}

// parseSingleConstDecl parses one constant declaration after 'const' has been consumed.
// Examples: MAX = 100;  APP_NAME: String = 'Kylix';
func (p *Parser) parseSingleConstDecl() *ast.ConstDecl {
	decl := &ast.ConstDecl{}
	if p.curTokenIs(token.IDENT) {
		decl.Token = p.curToken
		decl.Name = p.curToken.Literal
		p.nextToken()
	}
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		decl.Type = p.parseTypeExpression()
	}
	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()
		decl.Value = p.parseExpression(LOWEST)
		p.nextToken()
	}
	return decl
}

// parseSingleTypeDecl parses one type declaration after 'type' has been consumed.
// Examples: TPoint = record ... end;  TList<T> = class ... end;
func (p *Parser) parseSingleTypeDecl() *ast.TypeDecl {
	decl := &ast.TypeDecl{}
	if p.curTokenIs(token.IDENT) {
		decl.Token = p.curToken
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	var typeParams []*ast.TypeParameter
	if p.curTokenIs(token.LT) {
		typeParams = p.parseTypeParameterList()
	}

	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()
		if p.curTokenIs(token.CLASS) {
			classDecl := p.parseClassDecl()
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

	if p.curTokenIs(token.ASYNC) {
		decl.IsAsync = true
		p.nextToken()
	}

	decl.Token = p.curToken
	isProcedure := p.curTokenIs(token.PROCEDURE)
	isConstructor := p.curTokenIs(token.CONSTRUCTOR)
	isDestructor := p.curTokenIs(token.DESTRUCTOR)
	hasFuncKeyword := !isConstructor && !isDestructor
	p.nextToken()

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

	if p.curTokenIs(token.LT) {
		decl.TypeParams = p.parseTypeParameterList()
	}

	if p.curTokenIs(token.LPAREN) {
		decl.Parameters = p.parseParameterList()
	}

	// Return type: only for functions (not procedures, constructors, or destructors).
	if !isProcedure && hasFuncKeyword && p.curTokenIs(token.COLON) {
		p.nextToken()
		if p.curTokenIs(token.LPAREN) {
			// Tuple return type: (Type1, Type2)
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

	// virtual / override / abstract / static / external modifiers
	if p.curTokenIs(token.VIRTUAL) || p.curTokenIs(token.OVERRIDE) ||
		p.curTokenIs(token.ABSTRACT) || p.curTokenIs(token.STATIC) ||
		p.curTokenIs(token.DYNAMIC) {
		p.nextToken()
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.EXTERNAL) {
		decl.IsExternal = true
		p.nextToken()
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	// Local var/const sections before begin
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
		if p.curTokenIs(token.VAR) {
			p.nextToken() // skip 'var' modifier
		}
		if p.curTokenIs(token.IDENT) {
			param.Token = p.curToken
			param.Name = p.curToken.Literal
			p.nextToken()
		}
		if p.curTokenIs(token.COLON) {
			p.nextToken()
			param.Type = p.parseTypeExpression()
		}
		params = append(params, param)

		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			continue
		}
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		} else if !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
			p.nextToken() // safety advance to avoid infinite loop
		}
	}

	p.nextToken() // skip ')'
	return params
}

// parseTypeParameterList parses <T, U: Constraint> generic type parameter lists.
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

func (p *Parser) parseClassDecl() *ast.ClassDecl {
	decl := &ast.ClassDecl{Visibility: token.PUBLIC}
	p.nextToken() // skip 'class'

	// Class name (skip if immediately followed by ':' — that's a field declaration)
	if p.curTokenIs(token.IDENT) && !p.peekTokenIs(token.COLON) {
		decl.Name = p.curToken.Literal
		p.nextToken()
	}

	if p.curTokenIs(token.LT) {
		decl.TypeParams = p.parseTypeParameterList()
	}

	// Inheritance: class(TParent) or class inherits TParent
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
		p.nextToken()
		if p.curTokenIs(token.IDENT) {
			decl.Parent = p.curToken.Literal
			p.nextToken()
		}
	}

	// Implements clause
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

	// Class body
	for !p.curTokenIs(token.END) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.PUBLIC) || p.curTokenIs(token.PRIVATE) || p.curTokenIs(token.PROTECTED) {
			decl.Visibility = p.curToken.Type
			p.nextToken()
			continue
		}
		if p.curTokenIs(token.VAR) {
			varToken := p.curToken
			p.nextToken()
			for p.isIdentOrSoftKeyword() {
				field := p.parseSingleVarDecl(varToken)
				if field != nil {
					decl.Fields = append(decl.Fields, field)
				}
				for p.curTokenIs(token.SEMICOLON) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(token.FUNCTION) || p.curTokenIs(token.PROCEDURE) ||
			p.curTokenIs(token.CONSTRUCTOR) || p.curTokenIs(token.DESTRUCTOR) {
			method := p.parseFunctionDecl()
			if method != nil {
				decl.Methods = append(decl.Methods, method)
			}
		} else if p.curTokenIs(token.PROPERTY) {
			prop := p.parsePropertyDecl()
			if prop != nil {
				decl.Properties = append(decl.Properties, prop)
			}
		} else if p.isIdentOrSoftKeyword() && p.peekTokenIs(token.COLON) {
			// Bare field declaration without 'var': name: Type;
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

	// Parent interfaces: interface(IBase, IExtra)
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

// parsePropertyDecl parses: property Name: Type read Getter write Setter;
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
