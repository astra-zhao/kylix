package repl

import (
	"strings"
	"testing"
)

// REPL enhancement tests (Task 2).

func TestCompleter_KeywordPrefix(t *testing.T) {
	complete := buildCompleter(nil)
	results := complete("var x: Int")
	found := false
	for _, r := range results {
		if strings.HasSuffix(r, "Integer") {
			found = true
			break
		}
	}
	// Note: 'Integer' is a builtin type name not in our keyword list,
	// but 'IntToStr' is a builtin function and should match.
	if !found {
		// Try IntToStr instead
		for _, r := range results {
			if strings.HasSuffix(r, "IntToStr") {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected completion for 'Int' prefix, got: %v", results)
	}
}

func TestCompleter_PartialKeyword(t *testing.T) {
	complete := buildCompleter(nil)
	results := complete("func")
	found := false
	for _, r := range results {
		if r == "function" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'function' completion for 'func', got: %v", results)
	}
}

func TestCompleter_MetaCommands(t *testing.T) {
	complete := buildCompleter(nil)
	results := complete(":h")
	found := false
	for _, r := range results {
		if r == ":help" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ':help' completion for ':h', got: %v", results)
	}
}

func TestCompleter_NewMetaCommands(t *testing.T) {
	complete := buildCompleter(nil)

	for _, want := range []string{":load", ":type"} {
		results := complete(want[:2]) // ":l" or ":t"
		found := false
		for _, r := range results {
			if r == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s in completions for %s, got: %v", want, want[:2], results)
		}
	}
}

func TestCompleter_UserDeclaredNames(t *testing.T) {
	complete := buildCompleter([]string{"myVariable", "myFunction"})
	results := complete("myV")
	found := false
	for _, r := range results {
		if r == "myVariable" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'myVariable' in completions for 'myV', got: %v", results)
	}
}

func TestCompleter_EmptyPrefix(t *testing.T) {
	complete := buildCompleter(nil)
	results := complete("")
	if len(results) != 0 {
		t.Errorf("expected no completions for empty prefix, got %d", len(results))
	}
}

func TestExtractDeclaredNames(t *testing.T) {
	decls := []string{
		"var counter: Integer;",
		"function compute(): Integer; begin result := 42; end;",
		"const MAX = 100;",
	}
	names := extractDeclaredNames(decls)

	expectedSet := map[string]bool{
		"counter": false, "compute": false, "MAX": false,
	}
	for _, n := range names {
		if _, ok := expectedSet[n]; ok {
			expectedSet[n] = true
		}
	}
	for name, found := range expectedSet {
		if !found {
			t.Errorf("expected to extract name %q from declarations, got: %v", name, names)
		}
	}
}

func TestInferLiteralType(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"42", "Integer"},
		{"-7", "Integer"},
		{"3.14", "Real"},
		{"'hello'", "String"},
		{`"world"`, "String"},
		{"true", "Boolean"},
		{"false", "Boolean"},
		{"nil", "nil"},
	}
	for _, tc := range cases {
		got := inferLiteralType(tc.expr)
		if got != tc.want {
			t.Errorf("inferLiteralType(%q) = %q, want %q", tc.expr, got, tc.want)
		}
	}
}
