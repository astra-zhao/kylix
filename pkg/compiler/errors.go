// errors.go — Kylix error code definitions and Diagnostic constructors.
//
// Error code ranges:
//   KLX001-099  Lexer / syntax errors (from parser)
//   KLX100-199  Type errors
//   KLX200-299  Semantic errors (undeclared identifiers, arity, etc.)
//   KLX300-399  Interface / contract errors
//   KLX400-499  Compiler internal errors
package compiler

import "fmt"

// ── Error code constants ──────────────────────────────────────────────────────

const (
	// Syntax errors
	ErrUnexpectedToken   = "KLX001"
	ErrMissingToken      = "KLX002"
	ErrUnterminatedStr   = "KLX003"
	ErrParseGeneric      = "KLX004"
	ErrCircularDep       = "KLX005"

	// Type errors
	ErrTypeMismatch      = "KLX101"
	ErrCannotInferType   = "KLX102"
	ErrInvalidCast       = "KLX103"
	ErrGenericConstraint = "KLX104"
	ErrTypeAliasLoop     = "KLX105"

	// Semantic errors
	ErrUndeclared        = "KLX201"
	ErrWrongArity        = "KLX202"
	ErrDuplicateDecl     = "KLX203"
	ErrUninitializedVar  = "KLX204"
	ErrBreakOutsideLoop  = "KLX205"
	ErrReturnTypeMissing = "KLX206"

	// Interface / contract errors
	ErrMissingMethod     = "KLX301"
	ErrMethodSignature   = "KLX302"
	ErrUnknownInterface  = "KLX303"

	// Internal errors
	ErrInternal          = "KLX401"
	ErrCannotRead        = "KLX402"
	ErrCannotWrite       = "KLX403"
)

// ── Diagnostic constructors ───────────────────────────────────────────────────

// NewError creates an error Diagnostic with a code.
func NewError(file string, line, col int, code, msg string) Diagnostic {
	return Diagnostic{
		File:    file,
		Line:    line,
		Column:  col,
		Level:   "error",
		Code:    code,
		Message: msg,
	}
}

// NewErrorHint creates an error Diagnostic with a code and fix hint.
func NewErrorHint(file string, line, col int, code, msg, hint string) Diagnostic {
	return Diagnostic{
		File:    file,
		Line:    line,
		Column:  col,
		Level:   "error",
		Code:    code,
		Message: msg,
		Hint:    hint,
	}
}

// NewWarning creates a warning Diagnostic.
func NewWarning(file string, line, col int, code, msg string) Diagnostic {
	return Diagnostic{
		File:  file,
		Line:  line,
		Column: col,
		Level: "warning",
		Code:  code,
		Message: msg,
	}
}

// ── Diagnostic formatting ─────────────────────────────────────────────────────

// Format returns a rustc-style single-line representation.
// Example: "error[KLX201]: undeclared variable 'x' (main.klx:10:5)"
func (d Diagnostic) Format() string {
	codeStr := ""
	if d.Code != "" {
		codeStr = fmt.Sprintf("[%s]", d.Code)
	}
	loc := ""
	if d.File != "" {
		loc = fmt.Sprintf(" (%s:%d:%d)", d.File, d.Line, d.Column)
	}
	return fmt.Sprintf("%s%s: %s%s", d.Level, codeStr, d.Message, loc)
}

// FormatFull returns a multi-line, human-readable representation.
// Example:
//   error[KLX201]: undeclared variable 'x'
//     --> main.klx:10:5
//      |
//   10 |   WriteLn(x);
//      |           ^ not found in this scope
//      = help: declare it with 'var x: Type;'
func (d Diagnostic) FormatFull() string {
	codeStr := ""
	if d.Code != "" {
		codeStr = fmt.Sprintf("[%s]", d.Code)
	}
	out := fmt.Sprintf("%s%s: %s\n", d.Level, codeStr, d.Message)
	if d.File != "" {
		out += fmt.Sprintf("  --> %s:%d:%d\n", d.File, d.Line, d.Column)
	}
	if d.Source != "" {
		lineStr := fmt.Sprintf("%d", d.Line)
		out += fmt.Sprintf("   |\n")
		out += fmt.Sprintf("%s | %s\n", lineStr, d.Source)
		if d.Column > 0 {
			pad := spaces(len(lineStr) + 3 + d.Column - 1)
			out += fmt.Sprintf("%s^\n", pad)
		}
		out += fmt.Sprintf("   |\n")
	}
	if d.Hint != "" {
		out += fmt.Sprintf("   = help: %s\n", d.Hint)
	}
	return out
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}
