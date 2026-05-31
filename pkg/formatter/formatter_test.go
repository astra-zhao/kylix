package formatter

import (
	"kylix/lexer"
	"kylix/parser"
	"strings"
	"testing"
)

// Helper function to format Kylix source code
func formatSource(t *testing.T, source string) string {
	t.Helper()
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Logf("Parser errors (may be expected): %v", p.Errors())
	}

	f := New()
	return f.Format(program)
}

// Test bare raise statement (no exception)
func TestBareRaiseStatement(t *testing.T) {
	source := `
program Test;
begin
  raise;
end.
`
	result := formatSource(t, source)

	if !strings.Contains(result, "raise;") {
		t.Errorf("Expected 'raise;' in output, got:\n%s", result)
	}

	// Make sure there's no panic or nil dereference
	t.Logf("Bare raise formatted correctly:\n%s", result)
}

// Test raise with exception
func TestRaiseWithException(t *testing.T) {
	source := `
program Test;
begin
  raise Error;
end.
`
	result := formatSource(t, source)

	if !strings.Contains(result, "raise Error;") {
		t.Errorf("Expected 'raise Error;' in output, got:\n%s", result)
	}
}

// Test nested declarations in blocks
// Note: Parser has limitations with nested type declarations
func TestNestedDeclarations(t *testing.T) {
	// Simplified test - only test what parser can handle
	source := `
program Test;
const
  MAX = 100;

function Add(a: Integer; b: Integer): Integer;
begin
  result := a + b;
end;

var
  x: Integer;
begin
  x := 10;
end.
`
	result := formatSource(t, source)

	// Check that declarations are present
	checks := []string{
		"const",
		"MAX = 100",
		"function Add",
		"var",
		"x: Integer",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("Expected '%s' in output, got:\n%s", check, result)
		}
	}

	t.Logf("Declarations formatted:\n%s", result)
}

// Test expression precedence - should not add unnecessary parentheses
func TestExpressionPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		contains string
		notContains []string
	}{
		{
			input:    "result := a + b;",
			contains: "a + b",
			notContains: []string{"(a + b)"},
		},
		{
			input:    "result := a * b + c;",
			contains: "a * b + c",
			notContains: []string{"(a * b)", "(a * b + c)"},
		},
		{
			input:    "result := (a + b) * c;",
			contains: "(a + b) * c",
		},
		{
			input:    "if x > 0 then",
			contains: "x > 0",
			notContains: []string{"(x > 0)"},
		},
		{
			input:    "result := not x;",
			contains: "not x",
			notContains: []string{"(not x)"},
		},
	}

	for _, tt := range tests {
		source := "program Test;\nbegin\n  " + tt.input + "\nend."
		result := formatSource(t, source)

		if !strings.Contains(result, tt.contains) {
			t.Errorf("For input '%s': expected to contain '%s', got:\n%s",
				tt.input, tt.contains, result)
		}

		for _, notContain := range tt.notContains {
			if strings.Contains(result, notContain) {
				t.Errorf("For input '%s': should not contain '%s', got:\n%s",
					tt.input, notContain, result)
			}
		}
	}
}

// Test all statement types
func TestAllStatementTypes(t *testing.T) {
	source := `
program Test;
var
  x: Integer;
  y: Integer;
  i: Integer;
begin
  x := 10;

  if x > 5 then
  begin
    y := 1;
  end
  else
  begin
    y := 2;
  end;

  while x > 0 do
  begin
    x := x - 1;
  end;

  for i := 0 to 10 do
  begin
    y := y + i;
  end;
end.
`
	result := formatSource(t, source)

	// Check key statements are present
	checks := []string{
		"x := 10",
		"if x > 5",
		"while x > 0",
		"for i := 0 to 10",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("Expected '%s' in output, got:\n%s", check, result)
		}
	}

	t.Logf("Statement types formatted:\n%s", result)
}

// Test complex expressions
func TestComplexExpressions(t *testing.T) {
	source := `
program Test;
var
  x: Integer;
begin
  WriteLn('Hello');
  x := 10 + 20;
  if x > 5 then
  begin
    x := 1;
  end;
end.
`
	result := formatSource(t, source)

	checks := []string{
		"WriteLn('Hello')",
		"10 + 20",
		"x > 5",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("Expected '%s' in output, got:\n%s", check, result)
		}
	}

	t.Logf("Complex expressions formatted:\n%s", result)
}

// Test idempotency - formatting twice should give same result
func TestIdempotency(t *testing.T) {
	source := `
program Test;
var
  x: Integer;
begin
  x := 10;
  if x > 5 then
  begin
    x := 20;
  end;
end.
`
	first := formatSource(t, source)
	second := formatSource(t, first)

	if first != second {
		t.Errorf("Formatter is not idempotent!\nFirst:\n%s\nSecond:\n%s", first, second)
	}
}
