package lexer

import (
	"kylix/token"
	"testing"
)

func TestNextToken_Basic(t *testing.T) {
	input := `program hello;
begin
  WriteLn(42);
end.`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PROGRAM, "program"},
		{token.IDENT, "hello"},
		{token.SEMICOLON, ";"},
		{token.BEGIN, "begin"},
		{token.IDENT, "WriteLn"},
		{token.LPAREN, "("},
		{token.INT, "42"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.END, "end"},
		{token.DOT, "."},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Keywords(t *testing.T) {
	input := "if then else while do for to downto repeat until case of"

	tests := []token.TokenType{
		token.IF, token.THEN, token.ELSE,
		token.WHILE, token.DO, token.FOR, token.TO, token.DOWNTO,
		token.REPEAT, token.UNTIL, token.CASE, token.OF,
	}

	l := New(input)
	for i, expected := range tests {
		tok := l.NextToken()
		if tok.Type != expected {
			t.Fatalf("tests[%d] - expected %q, got %q (literal=%q)",
				i, expected, tok.Type, tok.Literal)
		}
	}
}

func TestNextToken_Operators(t *testing.T) {
	input := "= + - * / < > <= >= <> := == -> =>"
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.LT_EQ, "<="},
		{token.GT_EQ, ">="},
		{token.NOT_EQ, "<>"},
		{token.ASSIGN_OP, ":="},
		{token.EQ, "=="},
		{token.ARROW, "->"},
		{token.FAT_ARROW, "=>"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
	}
}

func TestNextToken_Strings(t *testing.T) {
	input := `"hello world" 'single quoted'`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.STRING || tok.Literal != "hello world" {
		t.Errorf("expected STRING 'hello world', got %q %q", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != token.STRING || tok.Literal != "single quoted" {
		t.Errorf("expected STRING 'single quoted', got %q %q", tok.Type, tok.Literal)
	}
}

func TestNextToken_StringInterpolation(t *testing.T) {
	input := `$"hello ${name}"`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.STRING_INTERPOLATION {
		t.Errorf("expected STRING_INTERPOLATION, got %q (literal=%q)", tok.Type, tok.Literal)
	}
}

func TestNextToken_Numbers(t *testing.T) {
	input := "42 3.14"
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.INT || tok.Literal != "42" {
		t.Errorf("expected INT 42, got %q %q", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != token.FLOAT || tok.Literal != "3.14" {
		t.Errorf("expected FLOAT 3.14, got %q %q", tok.Type, tok.Literal)
	}
}

func TestNextToken_Comments(t *testing.T) {
	input := `// line comment
program (* block comment *) hello`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.PROGRAM {
		t.Fatalf("expected PROGRAM, got %q (literal=%q)", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != token.IDENT || tok.Literal != "hello" {
		t.Fatalf("expected IDENT 'hello', got %q %q", tok.Type, tok.Literal)
	}
}

func TestNextToken_LineColumn(t *testing.T) {
	input := "a\nb\nc"
	l := New(input)

	// Line 1
	tok := l.NextToken()
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("expected line=1 col=1, got line=%d col=%d", tok.Line, tok.Column)
	}

	// Line 2
	tok = l.NextToken()
	if tok.Line != 2 || tok.Column != 1 {
		t.Errorf("expected line=2 col=1, got line=%d col=%d", tok.Line, tok.Column)
	}

	// Line 3
	tok = l.NextToken()
	if tok.Line != 3 || tok.Column != 1 {
		t.Errorf("expected line=3 col=1, got line=%d col=%d", tok.Line, tok.Column)
	}
}
