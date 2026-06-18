package token

import (
	"testing"
)

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		// Keywords
		{"program", PROGRAM},
		{"unit", UNIT},
		{"begin", BEGIN},
		{"end", END},
		{"function", FUNCTION},
		{"procedure", PROCEDURE},
		{"if", IF},
		{"then", THEN},
		{"else", ELSE},
		{"while", WHILE},
		{"for", FOR},
		{"class", CLASS},
		{"interface", INTERFACE},
		{"var", VAR},
		{"const", CONST},
		{"type", TYPE},
		{"true", TRUE},
		{"false", FALSE},
		{"nil", NIL},
		{"and", AND},
		{"or", OR},
		{"not", NOT},
		{"return", RETURN},
		{"exit", EXIT},
		{"match", MATCH},
		{"try", TRY},
		{"except", EXCEPT},
		{"finally", FINALLY},
		{"raise", RAISE},
		// Identifiers (not keywords)
		{"hello", IDENT},
		{"foo", IDENT},
		{"x", IDENT},
		{"myVar", IDENT},
		{"ClassName", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := LookupIdent(tt.input)
			if result != tt.expected {
				t.Errorf("LookupIdent(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLookupIdentCaseSensitive(t *testing.T) {
	// LookupIdent is case-sensitive in Go reference implementation.
	// Case-insensitivity is handled by the caller (e.g., lexer lowercases first).
	input := "Program"
	result := LookupIdent(input)
	if result != IDENT {
		t.Errorf("LookupIdent(%q) = %q, want %q (case-sensitive lookup)", input, result, IDENT)
	}

	// Lowercase version should match
	input = "program"
	result = LookupIdent(input)
	if result != PROGRAM {
		t.Errorf("LookupIdent(%q) = %q, want %q", input, result, PROGRAM)
	}
}

func TestKeywordsConsistency(t *testing.T) {
	// Every keyword in the map should have a valid TokenType
	for kw, tt := range keywords {
		if tt == "" {
			t.Errorf("keyword %q has empty TokenType", kw)
		}
		// Should be able to look it up
		result := LookupIdent(kw)
		if result != tt {
			t.Errorf("LookupIdent(%q) = %q, want %q (map has %q)", kw, result, tt, tt)
		}
	}
}

func TestTokenTypes(t *testing.T) {
	// Verify critical token type values
	if ILLEGAL != "ILLEGAL" {
		t.Error("ILLEGAL token type mismatch")
	}
	if EOF != "EOF" {
		t.Error("EOF token type mismatch")
	}
	if IDENT != "IDENT" {
		t.Error("IDENT token type mismatch")
	}
}

func TestOperators(t *testing.T) {
	// Verify operator token values
	ops := map[TokenType]string{
		ASSIGN:    "=",
		PLUS:      "+",
		MINUS:     "-",
		EQ:        "==",
		NOT_EQ:    "!=",
		LT:        "<",
		GT:        ">",
		LT_EQ:     "<=",
		GT_EQ:     ">=",
		ASSIGN_OP: ":=",
	}
	for tok, expected := range ops {
		if string(tok) != expected {
			t.Errorf("token %q has value %q, expected %q", tok, string(tok), expected)
		}
	}
}
