package lsp

import (
	"kylix/ast"
	"kylix/token"
)

// SymbolKind represents the type of a symbol
type SymbolKind int

const (
	SymbolVariable SymbolKind = iota
	SymbolConstant
	SymbolType
	SymbolFunction
	SymbolProcedure
	SymbolClass
	SymbolInterface
	SymbolMethod
	SymbolField
	SymbolProperty
	SymbolParameter
)

// String returns a human-readable string representation of SymbolKind
func (k SymbolKind) String() string {
	switch k {
	case SymbolVariable:
		return "Variable"
	case SymbolConstant:
		return "Constant"
	case SymbolType:
		return "Type"
	case SymbolFunction:
		return "Function"
	case SymbolProcedure:
		return "Procedure"
	case SymbolClass:
		return "Class"
	case SymbolInterface:
		return "Interface"
	case SymbolMethod:
		return "Method"
	case SymbolField:
		return "Field"
	case SymbolProperty:
		return "Property"
	case SymbolParameter:
		return "Parameter"
	default:
		return "Unknown"
	}
}

// Symbol represents a named entity in the code
type Symbol struct {
	Name     string
	Kind     SymbolKind
	Type     string // Type signature or description
	Location token.Token
	Scope    *Scope
	Children []*Symbol
}

// Scope represents a lexical scope containing symbols
type Scope struct {
	Parent   *Scope
	Symbols  map[string]*Symbol
	Children []*Scope
}

// NewScope creates a new scope
func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent:  parent,
		Symbols: make(map[string]*Symbol),
	}
}

// AddSymbol adds a symbol to this scope
func (s *Scope) AddSymbol(sym *Symbol) {
	s.Symbols[sym.Name] = sym
	sym.Scope = s
}

// FindSymbol finds a symbol in this scope or parent scopes
func (s *Scope) FindSymbol(name string) *Symbol {
	if sym, ok := s.Symbols[name]; ok {
		return sym
	}
	if s.Parent != nil {
		return s.Parent.FindSymbol(name)
	}
	return nil
}

// SymbolTable manages all symbols in a document
type SymbolTable struct {
	Root       *Scope
	AllSymbols []*Symbol
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		Root:       NewScope(nil),
		AllSymbols: make([]*Symbol, 0),
	}
}

// FindSymbol finds a symbol by name in the root scope
func (t *SymbolTable) FindSymbol(name string) *Symbol {
	return t.Root.FindSymbol(name)
}

// CollectSymbols walks the AST and collects all symbols
func CollectSymbols(program *ast.Program) *SymbolTable {
	table := NewSymbolTable()
	collector := &symbolCollector{
		table:       table,
		currentScope: table.Root,
	}

	// Collect global declarations
	for _, decl := range program.Declarations {
		collector.collectDeclaration(decl)
	}

	// Collect main program block
	if len(program.Statements) > 0 {
		mainScope := NewScope(table.Root)
		table.Root.Children = append(table.Root.Children, mainScope)
		collector.currentScope = mainScope
		for _, stmt := range program.Statements {
			collector.collectStatement(stmt)
		}
	}

	return table
}

type symbolCollector struct {
	table        *SymbolTable
	currentScope *Scope
}

func (c *symbolCollector) addSymbol(sym *Symbol) {
	c.currentScope.AddSymbol(sym)
	c.table.AllSymbols = append(c.table.AllSymbols, sym)
}

func (c *symbolCollector) collectDeclaration(node ast.Node) {
	switch d := node.(type) {
	case *ast.VarDecl:
		for _, name := range d.Names {
			sym := &Symbol{
				Name:     name,
				Kind:     SymbolVariable,
				Type:     c.formatType(d.Type),
				Location: d.Token,
			}
			c.addSymbol(sym)
		}

	case *ast.ConstDecl:
		sym := &Symbol{
			Name:     d.Name,
			Kind:     SymbolConstant,
			Type:     c.formatType(d.Type),
			Location: d.Token,
		}
		c.addSymbol(sym)

	case *ast.TypeDecl:
		sym := &Symbol{
			Name:     d.Name,
			Kind:     SymbolType,
			Type:     c.formatType(d.Type),
			Location: d.Token,
		}
		c.addSymbol(sym)

	case *ast.FunctionDecl:
		kind := SymbolFunction
		if d.ReturnType == nil {
			kind = SymbolProcedure
		}

		sym := &Symbol{
			Name:     d.Name,
			Kind:     SymbolKind(kind),
			Type:     c.formatFunctionSignature(d),
			Location: d.Token,
		}

		// Create function scope
		funcScope := NewScope(c.currentScope)
		c.currentScope.Children = append(c.currentScope.Children, funcScope)

		// Collect parameters
		for _, param := range d.Parameters {
			paramSym := &Symbol{
				Name:     param.Name,
				Kind:     SymbolParameter,
				Type:     c.formatType(param.Type),
				Location: param.Token,
			}
			funcScope.AddSymbol(paramSym)
			c.table.AllSymbols = append(c.table.AllSymbols, paramSym)
		}

		// Collect from function body
		if d.Body != nil {
			oldScope := c.currentScope
			c.currentScope = funcScope
			for _, stmt := range d.Body.Statements {
				c.collectStatement(stmt)
			}
			c.currentScope = oldScope
		}

		c.addSymbol(sym)

	case *ast.ClassDecl:
		sym := &Symbol{
			Name:     d.Name,
			Kind:     SymbolClass,
			Type:     c.formatClassDetail(d),
			Location: d.Token,
		}

		// Create class scope
		classScope := NewScope(c.currentScope)
		c.currentScope.Children = append(c.currentScope.Children, classScope)

		// Collect fields
		for _, field := range d.Fields {
			for _, name := range field.Names {
				fieldSym := &Symbol{
					Name:     name,
					Kind:     SymbolField,
					Type:     c.formatType(field.Type),
					Location: field.Token,
				}
				classScope.AddSymbol(fieldSym)
				c.table.AllSymbols = append(c.table.AllSymbols, fieldSym)
				sym.Children = append(sym.Children, fieldSym)
			}
		}

		// Collect properties
		for _, prop := range d.Properties {
			propSym := &Symbol{
				Name:     prop.Name,
				Kind:     SymbolProperty,
				Type:     c.formatType(prop.Type),
				Location: prop.Token,
			}
			classScope.AddSymbol(propSym)
			c.table.AllSymbols = append(c.table.AllSymbols, propSym)
			sym.Children = append(sym.Children, propSym)
		}

		// Collect methods
		for _, method := range d.Methods {
			methodSym := c.collectMethod(method, classScope)
			sym.Children = append(sym.Children, methodSym)
		}

		c.addSymbol(sym)

	case *ast.InterfaceDecl:
		sym := &Symbol{
			Name:     d.Name,
			Kind:     SymbolInterface,
			Type:     c.formatInterfaceDetail(d),
			Location: d.Token,
		}

		// Create interface scope
		ifaceScope := NewScope(c.currentScope)
		c.currentScope.Children = append(c.currentScope.Children, ifaceScope)

		// Collect methods
		for _, method := range d.Methods {
			methodSym := c.collectMethod(method, ifaceScope)
			sym.Children = append(sym.Children, methodSym)
		}

		c.addSymbol(sym)
	}
}

func (c *symbolCollector) collectMethod(method *ast.FunctionDecl, classScope *Scope) *Symbol {
	kind := SymbolMethod
	if method.ReturnType == nil {
		kind = SymbolProcedure
	}

	sym := &Symbol{
		Name:     method.Name,
		Kind:     SymbolKind(kind),
		Type:     c.formatFunctionSignature(method),
		Location: method.Token,
	}

	// Create method scope
	methodScope := NewScope(classScope)
	classScope.Children = append(classScope.Children, methodScope)

	// Collect parameters
	for _, param := range method.Parameters {
		paramSym := &Symbol{
			Name:     param.Name,
			Kind:     SymbolParameter,
			Type:     c.formatType(param.Type),
			Location: param.Token,
		}
		methodScope.AddSymbol(paramSym)
		c.table.AllSymbols = append(c.table.AllSymbols, paramSym)
	}

	// Collect from method body
	if method.Body != nil {
		oldScope := c.currentScope
		c.currentScope = methodScope
		for _, stmt := range method.Body.Statements {
			c.collectStatement(stmt)
		}
		c.currentScope = oldScope
	}

	classScope.AddSymbol(sym)
	c.table.AllSymbols = append(c.table.AllSymbols, sym)

	return sym
}

func (c *symbolCollector) collectStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		c.collectDeclaration(s)
	case *ast.ConstDecl:
		c.collectDeclaration(s)
	case *ast.BlockStatement:
		for _, st := range s.Statements {
			c.collectStatement(st)
		}
	case *ast.IfStatement:
		if s.Consequence != nil {
			c.collectStatement(s.Consequence)
		}
		if s.Alternative != nil {
			c.collectStatement(s.Alternative)
		}
	case *ast.WhileStatement:
		if s.Body != nil {
			c.collectStatement(s.Body)
		}
	case *ast.ForStatement:
		if s.Body != nil {
			c.collectStatement(s.Body)
		}
	case *ast.ForEachStatement:
		if s.Body != nil {
			c.collectStatement(s.Body)
		}
	case *ast.RepeatStatement:
		if s.Body != nil {
			c.collectStatement(s.Body)
		}
	case *ast.CaseStatement:
		for _, branch := range s.Branches {
			if branch.Body != nil {
				c.collectStatement(branch.Body)
			}
		}
		if s.ElseBranch != nil {
			c.collectStatement(s.ElseBranch)
		}
	case *ast.MatchStatement:
		for _, branch := range s.Branches {
			if branch.Body != nil {
				c.collectStatement(branch.Body)
			}
		}
	case *ast.TryStatement:
		if s.Body != nil {
			c.collectStatement(s.Body)
		}
		if s.ExceptBlock != nil {
			c.collectStatement(s.ExceptBlock)
		}
		if s.FinallyBlock != nil {
			c.collectStatement(s.FinallyBlock)
		}
	}
}

func (c *symbolCollector) formatType(expr ast.Expression) string {
	if expr == nil {
		return ""
	}
	switch t := expr.(type) {
	case *ast.Identifier:
		return t.Value
	case *ast.ArrayType:
		if t.Dynamic {
			return "array of " + c.formatType(t.ElementType)
		}
		return "array[" + c.formatType(t.Size) + "] of " + c.formatType(t.ElementType)
	case *ast.RecordType:
		return "record"
	case *ast.GenericType:
		result := t.Base + "<"
		for i, param := range t.TypeParams {
			if i > 0 {
				result += ", "
			}
			result += c.formatType(param)
		}
		result += ">"
		return result
	default:
		return ""
	}
}

func (c *symbolCollector) formatFunctionSignature(decl *ast.FunctionDecl) string {
	result := ""
	if decl.ReturnType == nil {
		result = "procedure "
	} else {
		result = "function "
	}

	result += decl.Name + "("

	for i, param := range decl.Parameters {
		if i > 0 {
			result += "; "
		}
		result += param.Name + ": " + c.formatType(param.Type)
	}

	result += ")"

	if decl.ReturnType != nil {
		result += ": " + c.formatType(decl.ReturnType)
	}

	return result
}

func (c *symbolCollector) formatClassDetail(decl *ast.ClassDecl) string {
	result := "class"
	if decl.Parent != "" {
		result += " inherits " + decl.Parent
	}
	if len(decl.Interfaces) > 0 {
		result += " implements"
		for i, iface := range decl.Interfaces {
			if i > 0 {
				result += ","
			}
			result += " " + iface
		}
	}
	return result
}

func (c *symbolCollector) formatInterfaceDetail(decl *ast.InterfaceDecl) string {
	result := "interface"
	if len(decl.Parents) > 0 {
		result += " extends"
		for i, parent := range decl.Parents {
			if i > 0 {
				result += ","
			}
			result += " " + parent
		}
	}
	return result
}
