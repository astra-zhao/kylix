package lexer

import (
	"kylix/token"
	"strings"
	"unicode"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	// Loop to handle consecutive comments and whitespace
	for {
		l.skipWhitespace()
		before := l.ch
		l.skipComments()
		if l.ch == before {
			break // No comment was skipped, we're done
		}
	}
	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.EQ, "==")
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.FAT_ARROW, "=>")
		} else {
			tok = l.newToken(token.ASSIGN, string(l.ch))
		}
	case '+':
		tok = l.newToken(token.PLUS, string(l.ch))
	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.ARROW, "->")
		} else {
			tok = l.newToken(token.MINUS, string(l.ch))
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.NOT_EQ, "!=")
		} else {
			tok = l.newToken(token.BANG, string(l.ch))
		}
	case '*':
		tok = l.newToken(token.ASTERISK, string(l.ch))
	case '/':
		tok = l.newToken(token.SLASH, string(l.ch))
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.LT_EQ, "<=")
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(token.NOT_EQ, "<>")
		} else {
			tok = l.newToken(token.LT, string(l.ch))
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.GT_EQ, ">=")
		} else {
			tok = l.newToken(token.GT, string(l.ch))
		}
	case ':':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(token.ASSIGN_OP, ":=")
		} else {
			tok = l.newToken(token.COLON, string(l.ch))
		}
	case ';':
		tok = l.newToken(token.SEMICOLON, string(l.ch))
	case ',':
		tok = l.newToken(token.COMMA, string(l.ch))
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			tok = l.newToken(token.DOTDOT, "..")
		} else {
			tok = l.newToken(token.DOT, string(l.ch))
		}
	case '(':
		tok = l.newToken(token.LPAREN, string(l.ch))
	case ')':
		tok = l.newToken(token.RPAREN, string(l.ch))
	case '{':
		tok = l.newToken(token.LBRACE, string(l.ch))
	case '}':
		tok = l.newToken(token.RBRACE, string(l.ch))
	case '[':
		tok = l.newToken(token.LBRACKET, string(l.ch))
	case ']':
		tok = l.newToken(token.RBRACKET, string(l.ch))
	case '\'':
		tok.Type = token.STRING
		tok.Literal = l.readSingleQuotedString()
		return tok
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		return tok
	case '$':
		if l.peekChar() == '"' {
			l.readChar() // skip $
			tok.Type = token.STRING_INTERPOLATION
			tok.Literal = l.readInterpolatedString()
			return tok
		}
		tok = l.newToken(token.ILLEGAL, string(l.ch))
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Line = l.line
			tok.Column = l.column
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(strings.ToLower(tok.Literal))
			return tok
		} else if isDigit(l.ch) {
			tok.Line = l.line
			tok.Column = l.column
			return l.readNumber()
		} else {
			tok = l.newToken(token.ILLEGAL, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) newToken(tokenType token.TokenType, literal string) token.Token {
	return token.Token{Type: tokenType, Literal: literal, Line: l.line, Column: l.column}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComments() {
	// Handle // line comments
	if l.ch == '/' && l.peekChar() == '/' {
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	}
	// Handle (* block comments *)
	if l.ch == '(' && l.peekChar() == '*' {
		l.readChar()
		l.readChar()
		for !(l.ch == '*' && l.peekChar() == ')') && l.ch != 0 {
			l.readChar()
		}
		if l.ch == '*' {
			l.readChar()
			l.readChar()
		}
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() token.Token {
	position := l.position
	line := l.line
	column := l.column
	tokType := token.TokenType(token.INT)

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		tokType = token.FLOAT
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return token.Token{Type: tokType, Literal: l.input[position:l.position], Line: line, Column: column}
}

func (l *Lexer) readString() string {
	l.readChar() // skip opening "
	position := l.position
	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}
	str := l.input[position:l.position]
	l.readChar() // skip closing "
	return str
}

func (l *Lexer) readInterpolatedString() string {
	l.readChar() // skip opening "
	position := l.position
	braceDepth := 0

	for l.ch != 0 {
		if l.ch == '{' {
			braceDepth++
		} else if l.ch == '}' {
			braceDepth--
		} else if l.ch == '"' && braceDepth == 0 {
			break
		}
		l.readChar()
	}

	str := l.input[position:l.position]
	l.readChar() // skip closing "
	return str
}

func (l *Lexer) readSingleQuotedString() string {
	l.readChar() // skip opening '
	position := l.position
	for l.ch != '\'' && l.ch != 0 {
		l.readChar()
	}
	str := l.input[position:l.position]
	l.readChar() // skip closing '
	return str
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
