package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	// Special tokens
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	COMMENT = "COMMENT"

	// Identifiers + literals
	IDENT                = "IDENT"
	INT                  = "INT"
	FLOAT                = "FLOAT"
	STRING               = "STRING"
	STRING_INTERPOLATION = "STRING_INTERPOLATION"
	CHAR                 = "CHAR"

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	MOD      = "mod"
	DIV      = "div"

	LT     = "<"
	GT     = ">"
	EQ     = "=="
	NOT_EQ = "!="
	LT_EQ  = "<="
	GT_EQ  = ">="

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	DOT       = "."
	DOTDOT    = ".."
	ARROW     = "->"
	FAT_ARROW = "=>"
	ASSIGN_OP = ":="

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	PROGRAM     = "program"
	UNIT        = "unit"
	USES        = "uses"
	VAR         = "var"
	CONST       = "const"
	TYPE        = "type"
	BEGIN       = "begin"
	END         = "end"
	FUNCTION    = "function"
	PROCEDURE   = "procedure"
	IF          = "if"
	THEN        = "then"
	ELSE        = "else"
	WHILE       = "while"
	DO          = "do"
	FOR         = "for"
	TO          = "to"
	DOWNTO      = "downto"
	REPEAT      = "repeat"
	UNTIL       = "until"
	CASE        = "case"
	OF          = "of"
	WITH        = "with"
	TRY         = "try"
	EXCEPT      = "except"
	FINALLY     = "finally"
	RAISE       = "raise"
	CLASS       = "class"
	INTERFACE   = "interface"
	OBJECT      = "object"
	RECORD      = "record"
	ARRAY       = "array"
	SET         = "set"
	PACKED      = "packed"
	FILE        = "file"
	INHERITS    = "inherits"
	IMPLEMENTS  = "implements"
	PUBLIC      = "public"
	PRIVATE     = "private"
	PROTECTED   = "protected"
	PUBLISHED   = "published"
	PROPERTY    = "property"
	READ        = "read"
	WRITE       = "write"
	DEFAULT     = "default"
	STORED      = "stored"
	VIRTUAL     = "virtual"
	OVERRIDE    = "override"
	ABSTRACT    = "abstract"
	STATIC      = "static"
	DYNAMIC     = "dynamic"
	EXTERNAL    = "external"
	FORWARD     = "forward"
	INLINE      = "inline"
	RESULT      = "result"
	SELF        = "self"
	NIL         = "nil"
	TRUE        = "true"
	FALSE       = "false"
	AND         = "and"
	OR          = "or"
	NOT         = "not"
	XOR         = "xor"
	IN          = "in"
	IS          = "is"
	AS          = "as"
	NEW         = "new"
	DELETE      = "delete"
	BREAK       = "break"
	CONTINUE    = "continue"
	EXIT        = "exit"
	ASYNC       = "async"
	AWAIT       = "await"
	CONSTRUCTOR = "constructor"
	DESTRUCTOR  = "destructor"
	INHERITED   = "inherited"
	MATCH       = "match"
	WHEN        = "when"
	ON          = "on"
	IMPORT      = "import"
	EXPORT      = "export"
	MODULE      = "module"
	RETURN      = "return"
	MAP         = "map"
	VARIANT     = "variant"

	// Types
	INTEGER_TYPE = "Integer"
	REAL_TYPE    = "Real"
	BOOLEAN_TYPE = "Boolean"
	STRING_TYPE  = "String"
	CHAR_TYPE    = "Char"
)

var keywords = map[string]TokenType{
	"program":     PROGRAM,
	"unit":        UNIT,
	"uses":        USES,
	"var":         VAR,
	"const":       CONST,
	"type":        TYPE,
	"begin":       BEGIN,
	"end":         END,
	"function":    FUNCTION,
	"procedure":   PROCEDURE,
	"if":          IF,
	"then":        THEN,
	"else":        ELSE,
	"while":       WHILE,
	"do":          DO,
	"for":         FOR,
	"to":          TO,
	"downto":      DOWNTO,
	"repeat":      REPEAT,
	"until":       UNTIL,
	"case":        CASE,
	"of":          OF,
	"with":        WITH,
	"try":         TRY,
	"except":      EXCEPT,
	"finally":     FINALLY,
	"raise":       RAISE,
	"class":       CLASS,
	"interface":   INTERFACE,
	"object":      OBJECT,
	"record":      RECORD,
	"array":       ARRAY,
	"set":         SET,
	"packed":      PACKED,
	"file":        FILE,
	"inherits":    INHERITS,
	"implements":  IMPLEMENTS,
	"public":      PUBLIC,
	"private":     PRIVATE,
	"protected":   PROTECTED,
	"published":   PUBLISHED,
	"property":    PROPERTY,
	"read":        READ,
	"write":       WRITE,
	"default":     DEFAULT,
	"stored":      STORED,
	"virtual":     VIRTUAL,
	"override":    OVERRIDE,
	"abstract":    ABSTRACT,
	"static":      STATIC,
	"dynamic":     DYNAMIC,
	"external":    EXTERNAL,
	"forward":     FORWARD,
	"inline":      INLINE,
	"result":      RESULT,
	"self":        SELF,
	"nil":         NIL,
	"true":        TRUE,
	"false":       FALSE,
	"and":         AND,
	"or":          OR,
	"not":         NOT,
	"xor":         XOR,
	"in":          IN,
	"is":          IS,
	"as":          AS,
	"new":         NEW,
	"delete":      DELETE,
	"break":       BREAK,
	"continue":    CONTINUE,
	"exit":        EXIT,
	"async":       ASYNC,
	"await":       AWAIT,
	"constructor": CONSTRUCTOR,
	"destructor":  DESTRUCTOR,
	"inherited":   INHERITED,
	"match":       MATCH,
	"when":        WHEN,
	"on":          ON,
	"import":      IMPORT,
	"export":      EXPORT,
	"module":      MODULE,
	"return":      RETURN,
	"map":         MAP,
	"variant":     VARIANT,
	"mod":         MOD,
	"div":         DIV,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
