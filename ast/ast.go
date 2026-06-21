package ast

import "kylix/token"

// Node represents a node in the AST
type Node interface {
	TokenLiteral() string
}

// Statement represents a statement node
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node
type Program struct {
	Name         string
	NameToken    token.Token // NEW: position of program name
	UnitName     string      // module name for unit files
	IsUnit       bool        // true if this is a unit file (no main function)
	Uses         []string
	Declarations []Node
	Statements   []Statement
}

func (p *Program) TokenLiteral() string { return p.Name }

// Variable Declaration
type VarDecl struct {
	Token    token.Token // NEW: the 'var' keyword
	Names    []string
	Type     Expression
	Value    Expression
	Inferred bool // true if using :=
}

func (v *VarDecl) statementNode()       {}
func (v *VarDecl) TokenLiteral() string { return v.Token.Literal }

// Constant Declaration
type ConstDecl struct {
	Token token.Token // NEW: the identifier token
	Name  string
	Type  Expression
	Value Expression
}

func (c *ConstDecl) statementNode()       {}
func (c *ConstDecl) TokenLiteral() string { return c.Token.Literal }

// Type Declaration
type TypeDecl struct {
	Token token.Token // NEW: the identifier token
	Name  string
	Type  Expression
}

func (t *TypeDecl) statementNode()       {}
func (t *TypeDecl) TokenLiteral() string { return t.Token.Literal }

// Function/Procedure Declaration
type FunctionDecl struct {
	Token       token.Token // NEW: the 'function' or 'procedure' keyword
	Name        string
	TypeParams  []*TypeParameter // generic type parameters (Go 1.18+)
	Parameters  []*Parameter
	ReturnType  Expression   // single return type (traditional)
	ReturnTypes []Expression // multiple return types (modern: (Type1, Type2))
	Body        *BlockStatement
	LocalDecls  []Node // local var/const declarations before begin block
	IsAsync     bool
	IsExport    bool
	IsExternal  bool // body implemented in Go stdlib; no Kylix body emitted
}

func (f *FunctionDecl) statementNode()       {}
func (f *FunctionDecl) TokenLiteral() string { return f.Token.Literal }

// TypeParameter represents a generic type parameter
type TypeParameter struct {
	Token      token.Token // the identifier token
	Name       string
	Constraint Expression // optional constraint (e.g. "constraint" or type name)
}

type Parameter struct {
	Token token.Token // NEW: the parameter name token
	Name  string
	Type  Expression
}

// Block Statement
type BlockStatement struct {
	Token      token.Token // NEW: the 'begin' keyword
	Statements []Statement
}

func (b *BlockStatement) statementNode()       {}
func (b *BlockStatement) TokenLiteral() string { return b.Token.Literal }

// Assignment Statement
type AssignmentStatement struct {
	Token token.Token // NEW: the ':=' token
	Name  Expression
	Value Expression
}

func (a *AssignmentStatement) statementNode()       {}
func (a *AssignmentStatement) TokenLiteral() string { return a.Token.Literal }

// Return Statement
type ReturnStatement struct {
	Token token.Token // NEW: the 'return' keyword
	Value Expression
}

func (r *ReturnStatement) statementNode()       {}
func (r *ReturnStatement) TokenLiteral() string { return r.Token.Literal }

// Expression Statement
type ExpressionStatement struct {
	Token      token.Token // NEW: the first token of the expression
	Expression Expression
}

func (e *ExpressionStatement) statementNode()       {}
func (e *ExpressionStatement) TokenLiteral() string { return e.Token.Literal }

// If Statement
type IfStatement struct {
	Token       token.Token // NEW: the 'if' keyword
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (i *IfStatement) statementNode()       {}
func (i *IfStatement) TokenLiteral() string { return i.Token.Literal }

// While Statement
type WhileStatement struct {
	Token     token.Token // NEW: the 'while' keyword
	Condition Expression
	Body      *BlockStatement
}

func (w *WhileStatement) statementNode()       {}
func (w *WhileStatement) TokenLiteral() string { return w.Token.Literal }

// For Statement
type ForStatement struct {
	Token    token.Token // NEW: the 'for' keyword
	Variable string
	From     Expression
	To       Expression
	DownTo   bool
	Body     *BlockStatement
}

func (f *ForStatement) statementNode()       {}
func (f *ForStatement) TokenLiteral() string { return f.Token.Literal }

// ForEach Statement (modern feature)
type ForEachStatement struct {
	Token    token.Token // NEW: the 'for' keyword
	Variable string
	Iterable Expression
	Body     *BlockStatement
}

func (f *ForEachStatement) statementNode()       {}
func (f *ForEachStatement) TokenLiteral() string { return f.Token.Literal }

// Repeat Statement
type RepeatStatement struct {
	Token     token.Token // NEW: the 'repeat' keyword
	Body      *BlockStatement
	Condition Expression
}

func (r *RepeatStatement) statementNode()       {}
func (r *RepeatStatement) TokenLiteral() string { return r.Token.Literal }

// Case Statement
type CaseStatement struct {
	Token      token.Token // NEW: the 'case' keyword
	Expression Expression
	Branches   []*CaseBranch
	ElseBranch *BlockStatement
}

func (c *CaseStatement) statementNode()       {}
func (c *CaseStatement) TokenLiteral() string { return c.Token.Literal }

type CaseBranch struct {
	Values []Expression
	Body   *BlockStatement
}

// Match Statement (modern pattern matching)
type MatchStatement struct {
	Token      token.Token // NEW: the 'match' keyword
	Expression Expression
	Branches   []*MatchBranch
}

func (m *MatchStatement) statementNode()       {}
func (m *MatchStatement) TokenLiteral() string { return m.Token.Literal }

type MatchBranch struct {
	Pattern            Expression
	AdditionalPatterns []Expression // for multi-pattern: 2, 3 =>
	When               Expression   // optional guard
	Body               *BlockStatement
}

// On Clause (exception filter: on E: ExceptionType do)
type OnClause struct {
	Token    token.Token // the 'on' keyword
	Variable string      // E
	Type     Expression  // ExceptionType
	Body     *BlockStatement
}

func (o *OnClause) statementNode()       {}
func (o *OnClause) TokenLiteral() string { return o.Token.Literal }

// Try Statement
type TryStatement struct {
	Token        token.Token // NEW: the 'try' keyword
	Body         *BlockStatement
	OnClauses    []*OnClause // on E: Type do clauses (Modern Pascal)
	ExceptBlock  *BlockStatement
	FinallyBlock *BlockStatement
}

func (t *TryStatement) statementNode()       {}
func (t *TryStatement) TokenLiteral() string { return t.Token.Literal }

// Raise Statement
type RaiseStatement struct {
	Token     token.Token // NEW: the 'raise' keyword
	Exception Expression
}

func (r *RaiseStatement) statementNode()       {}
func (r *RaiseStatement) TokenLiteral() string { return r.Token.Literal }

// Break Statement
type BreakStatement struct {
	Token token.Token // NEW: the 'break' keyword
}

func (b *BreakStatement) statementNode()       {}
func (b *BreakStatement) TokenLiteral() string { return b.Token.Literal }

// Continue Statement
type ContinueStatement struct {
	Token token.Token // NEW: the 'continue' keyword
}

func (c *ContinueStatement) statementNode()       {}
func (c *ContinueStatement) TokenLiteral() string { return c.Token.Literal }

// Inherited Statement
type InheritedStatement struct {
	Token token.Token // the 'inherited' keyword
	Expr  Expression  // Optional: inherited MethodName(args)
}

func (is *InheritedStatement) statementNode()       {}
func (is *InheritedStatement) TokenLiteral() string { return is.Token.Literal }

// Class Declaration
type ClassDecl struct {
	Token      token.Token // NEW: the 'class' keyword
	Name       string
	TypeParams []*TypeParameter // generic type parameters
	Parent     string
	Interfaces []string
	Fields     []*VarDecl
	Methods    []*FunctionDecl
	Properties []*PropertyDecl
	Visibility token.TokenType
}

func (c *ClassDecl) statementNode()       {}
func (c *ClassDecl) expressionNode()      {}
func (c *ClassDecl) TokenLiteral() string { return c.Token.Literal }

// Interface Declaration
type InterfaceDecl struct {
	Token   token.Token // NEW: the 'interface' keyword
	Name    string
	Parents []string
	Methods []*FunctionDecl
}

func (i *InterfaceDecl) statementNode()       {}
func (i *InterfaceDecl) expressionNode()      {}
func (i *InterfaceDecl) TokenLiteral() string { return i.Token.Literal }

// Property Declaration
type PropertyDecl struct {
	Token   token.Token // NEW: the 'property' keyword
	Name    string
	Type    Expression
	Getter  string
	Setter  string
	Default Expression
}

func (p *PropertyDecl) statementNode()       {}
func (p *PropertyDecl) TokenLiteral() string { return p.Token.Literal }

// Record Type
type RecordType struct {
	Fields []*VarDecl
}

func (r *RecordType) expressionNode()      {}
func (r *RecordType) TokenLiteral() string { return "record" }

// Array Type
type ArrayType struct {
	ElementType Expression
	Size        Expression
	Dynamic     bool
}

func (a *ArrayType) expressionNode()      {}
func (a *ArrayType) TokenLiteral() string { return "array" }

// Map Type (modern feature) — e.g., map[String]Integer
type MapType struct {
	KeyType   Expression
	ValueType Expression
}

func (m *MapType) expressionNode()      {}
func (m *MapType) TokenLiteral() string { return "map" }

// Variant Type (discriminated union) — e.g.,
//
//	type TExpr = variant
//	  IntLiteral: Integer;
//	  StrLiteral: String;
//	end;
type VariantType struct {
	Cases []*VariantCase
}

func (v *VariantType) expressionNode()      {}
func (v *VariantType) TokenLiteral() string { return "variant" }

type VariantCase struct {
	Name string
	Type Expression
}

// Enum Type — e.g., type TTokenType = (tkEOF, tkIdent, ...);
type EnumType struct {
	Names []string
}

func (e *EnumType) expressionNode()      {}
func (e *EnumType) TokenLiteral() string { return "enum" }

// Generic Type (modern feature)
type GenericType struct {
	Base       string
	TypeParams []Expression
}

func (g *GenericType) expressionNode()      {}
func (g *GenericType) TokenLiteral() string { return g.Base }

// Expressions

// Identifier
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// Integer Literal
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (i *IntegerLiteral) expressionNode()      {}
func (i *IntegerLiteral) TokenLiteral() string { return i.Token.Literal }

// Float Literal
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (f *FloatLiteral) expressionNode()      {}
func (f *FloatLiteral) TokenLiteral() string { return f.Token.Literal }

// String Literal
type StringLiteral struct {
	Token token.Token
	Value string
}

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }

// String Interpolation (modern feature)
type StringInterpolation struct {
	Parts []Expression
}

func (s *StringInterpolation) expressionNode()      {}
func (s *StringInterpolation) TokenLiteral() string { return "interpolation" }

// Boolean Literal
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }

// Nil Literal
type NilLiteral struct {
	Token token.Token
}

func (n *NilLiteral) expressionNode()      {}
func (n *NilLiteral) TokenLiteral() string { return n.Token.Literal }

// Array Literal
type ArrayLiteral struct {
	Token    token.Token // NEW: the '[' token
	Elements []Expression
}

func (a *ArrayLiteral) expressionNode()      {}
func (a *ArrayLiteral) TokenLiteral() string { return a.Token.Literal }

// Tuple Literal (modern feature) — e.g., (expr1, expr2)
type TupleLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (t *TupleLiteral) expressionNode()      {}
func (t *TupleLiteral) TokenLiteral() string { return "tuple" }

// Lambda Expression (modern feature)
type LambdaExpression struct {
	Token      token.Token // NEW: the '(' or first param token
	Parameters []*Parameter
	Body       Node // can be BlockStatement or Expression
}

func (l *LambdaExpression) expressionNode()      {}
func (l *LambdaExpression) TokenLiteral() string { return l.Token.Literal }

// Prefix Expression
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (p *PrefixExpression) expressionNode()      {}
func (p *PrefixExpression) TokenLiteral() string { return p.Token.Literal }

// Infix Expression
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (i *InfixExpression) expressionNode()      {}
func (i *InfixExpression) TokenLiteral() string { return i.Token.Literal }

// Call Expression
type CallExpression struct {
	Token     token.Token // NEW: the '(' token
	Function  Expression
	Arguments []Expression
}

func (c *CallExpression) expressionNode()      {}
func (c *CallExpression) TokenLiteral() string { return c.Token.Literal }

// Member Expression
type MemberExpression struct {
	Token  token.Token // NEW: the '.' token
	Object Expression
	Member string
}

func (m *MemberExpression) expressionNode()      {}
func (m *MemberExpression) TokenLiteral() string { return m.Token.Literal }

// Index Expression
type IndexExpression struct {
	Token token.Token // NEW: the '[' token
	Left  Expression
	Index Expression
}

func (i *IndexExpression) expressionNode()      {}
func (i *IndexExpression) TokenLiteral() string { return i.Token.Literal }

// Slice Expression — e.g., s[a:b]
type SliceExpression struct {
	Token token.Token
	Left  Expression
	Low   Expression
	High  Expression
}

func (s *SliceExpression) expressionNode()      {}
func (s *SliceExpression) TokenLiteral() string { return s.Token.Literal }

// Await Expression (modern feature)
type AwaitExpression struct {
	Token      token.Token // NEW: the 'await' keyword
	Expression Expression
}

func (a *AwaitExpression) expressionNode()      {}
func (a *AwaitExpression) TokenLiteral() string { return a.Token.Literal }

// Type Cast Expression
type TypeCastExpression struct {
	Token      token.Token // NEW: the 'as' keyword
	Expression Expression
	TargetType Expression
}

func (t *TypeCastExpression) expressionNode()      {}
func (t *TypeCastExpression) TokenLiteral() string { return t.Token.Literal }

// Is Expression (type check)
type IsExpression struct {
	Token      token.Token // NEW: the 'is' keyword
	Expression Expression
	TargetType Expression
}

func (i *IsExpression) expressionNode()      {}
func (i *IsExpression) TokenLiteral() string { return i.Token.Literal }
