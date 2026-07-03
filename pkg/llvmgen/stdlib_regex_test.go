package llvmgen_test

import (
	"strings"
	"testing"
)

// TestRegexIsEmail verifies IR generation for IsEmail(str) -> i1
func TestRegexIsEmail(t *testing.T) {
	ir := generateIR(t, `
		program RegexTest;
		uses regex;
		begin
		  var valid := IsEmail('test@example.com');
		end.
	`)
	// Check function definition with correct signature
	if !strings.Contains(ir, "define i1 @__kylix_regex_IsEmail(ptr %str)") {
		t.Fatal("missing IsEmail function definition")
	}
	// Check regcomp call with REG_EXTENDED | REG_NOSUB (9)
	if !strings.Contains(ir, "call i32 @regcomp(ptr") {
		t.Fatal("missing regcomp call")
	}
	if !strings.Contains(ir, ", i32 9)") {
		t.Fatal("regcomp should use cflags=9 (REG_EXTENDED|REG_NOSUB)")
	}
	// Check regexec call
	if !strings.Contains(ir, "call i32 @regexec(ptr") {
		t.Fatal("missing regexec call")
	}
	// Check regfree call
	if !strings.Contains(ir, "call void @regfree(ptr") {
		t.Fatal("missing regfree call")
	}
	// Check main calls IsEmail
	if !strings.Contains(ir, "call i1 @__kylix_regex_IsEmail(ptr") {
		t.Fatal("main should call IsEmail")
	}
}

// TestRegexIsURL verifies IR generation for IsURL(str) -> i1
func TestRegexIsURL(t *testing.T) {
	ir := generateIR(t, `
		program RegexTest;
		uses regex;
		begin
		  var valid := IsURL('https://example.com');
		end.
	`)
	if !strings.Contains(ir, "define i1 @__kylix_regex_IsURL(ptr %str)") {
		t.Fatal("missing IsURL function definition")
	}
	if !strings.Contains(ir, "call i32 @regcomp(ptr") {
		t.Fatal("missing regcomp call")
	}
	if !strings.Contains(ir, "call i32 @regexec(ptr") {
		t.Fatal("missing regexec call")
	}
	if !strings.Contains(ir, "call void @regfree(ptr") {
		t.Fatal("missing regfree call")
	}
}

// TestRegexIsNumeric verifies IR generation for IsNumeric(str) -> i1
func TestRegexIsNumeric(t *testing.T) {
	ir := generateIR(t, `
		program RegexTest;
		uses regex;
		begin
		  var valid := IsNumeric('12345');
		end.
	`)
	if !strings.Contains(ir, "define i1 @__kylix_regex_IsNumeric(ptr %str)") {
		t.Fatal("missing IsNumeric function definition")
	}
	if !strings.Contains(ir, "call i32 @regcomp(ptr") {
		t.Fatal("missing regcomp call")
	}
	if !strings.Contains(ir, "call i32 @regexec(ptr") {
		t.Fatal("missing regexec call")
	}
}

// TestRegexIsAlpha verifies IR generation for IsAlpha(str) -> i1
func TestRegexIsAlpha(t *testing.T) {
	ir := generateIR(t, `
		program RegexTest;
		uses regex;
		begin
		  var valid := IsAlpha('hello');
		end.
	`)
	if !strings.Contains(ir, "define i1 @__kylix_regex_IsAlpha(ptr %str)") {
		t.Fatal("missing IsAlpha function definition")
	}
	if !strings.Contains(ir, "call i32 @regcomp(ptr") {
		t.Fatal("missing regcomp call")
	}
}

// TestRegexIsAlphaNumeric verifies IR generation for IsAlphaNumeric(str) -> i1
func TestRegexIsAlphaNumeric(t *testing.T) {
	ir := generateIR(t, `
		program RegexTest;
		uses regex;
		begin
		  var valid := IsAlphaNumeric('abc123');
		end.
	`)
	if !strings.Contains(ir, "define i1 @__kylix_regex_IsAlphaNumeric(ptr %str)") {
		t.Fatal("missing IsAlphaNumeric function definition")
	}
	if !strings.Contains(ir, "call i32 @regcomp(ptr") {
		t.Fatal("missing regcomp call")
	}
}

// TestRegexIsIP verifies IR generation for IsIP(str) -> i1
func TestRegexIsIP(t *testing.T) {
	ir := generateIR(t, `
		program RegexTest;
		uses regex;
		begin
		  var valid := IsIP('192.168.1.1');
		end.
	`)
	if !strings.Contains(ir, "define i1 @__kylix_regex_IsIP(ptr %str)") {
		t.Fatal("missing IsIP function definition")
	}
	if !strings.Contains(ir, "call i32 @regcomp(ptr") {
		t.Fatal("missing regcomp call")
	}
}

// TestRegexBareNameCall verifies bare-name calls (IsEmail(...) without regex. prefix)
func TestRegexBareNameCall(t *testing.T) {
	ir := generateIR(t, `
		program RegexTest;
		uses regex;
		begin
		  if IsEmail('test@example.com') then
		    WriteLn('Valid');
		end.
	`)
	// Should resolve bare IsEmail to regex.IsEmail
	if !strings.Contains(ir, "call i1 @__kylix_regex_IsEmail(ptr") {
		t.Fatal("bare-name IsEmail should resolve to @__kylix_regex_IsEmail")
	}
}

// TestRegexQualifiedCall verifies qualified calls (regex.IsEmail(...))
func TestRegexQualifiedCall(t *testing.T) {
	ir := generateIR(t, `
		program RegexTest;
		uses regex;
		begin
		  if regex.IsEmail('test@example.com') then
		    WriteLn('Valid');
		end.
	`)
	// Should dispatch qualified call
	if !strings.Contains(ir, "call i1 @__kylix_regex_IsEmail(ptr") {
		t.Fatal("qualified regex.IsEmail should resolve to @__kylix_regex_IsEmail")
	}
}
