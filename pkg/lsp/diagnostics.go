package lsp

import (
	"fmt"
	"reflect"
	"strings"

	"kylix/ast"
	"kylix/token"
)

// diagnostics.go — v0.6.8 semantic diagnostics at warning severity.
//
// The LSP server previously published only parse/semantic errors (severity 1).
// This file adds a warning-level check for identifiers that are used but never
// declared (undefined variable/function/type references), so the IDE can draw
// a yellow squiggle under them while typing. Declarations and Kylix built-ins
// (WriteLn, Integer, Self, ...) are never reported. Pascal identifiers are
// case-insensitive, so all lookups are lowercased.

// kylixBuiltinIdentifiers are names the language provides without a `uses`
// clause or a user declaration. Lowercased keys; conservative on purpose — an
// over-eager warning is worse than a missed one.
var kylixBuiltinIdentifiers = map[string]bool{
	// primitive types
	"integer": true, "string": true, "boolean": true, "real": true,
	"double": true, "single": true, "variant": true, "char": true,
	"byte": true, "shortint": true, "longint": true, "word": true,
	"cardinal": true, "int64": true, "pointer": true, "object": true,
	"tobject": true,
	// KylixBoot framework types (declared in stdlib/klx/web.klx, which the
	// module loader does not recurse into from boot.klx).
	"trequest": true, "tresponse": true, "tserver": true, "thttpclient": true,
	"thttpresponse": true, "tcache": true,
	// procedures / functions
	"write": true, "writeln": true, "read": true, "readln": true,
	"inc": true, "dec": true, "length": true, "high": true, "low": true,
	"ord": true, "chr": true, "abs": true, "sqr": true, "sqrt": true,
	"trunc": true, "round": true, "random": true, "randomize": true,
	"halt": true, "exit": true, "break": true, "continue": true,
	"strtoint": true, "strtofloat": true, "inttostr": true,
	"floattostr": true, "booltostr": true, "strtobool": true,
	"upcase": true, "lowercase": true, "trim": true, "copy": true,
	"pos": true, "setlength": true, "append": true, "new": true,
	"dispose": true, "pred": true, "succ": true,
	// pseudo variables
	"self": true, "result": true,
	// literals
	"true": true, "false": true, "nil": true,
	// exception base class (implicitly available in try/except)
	"exception": true,
}

// collectIdentifierTokens walks v recursively and records every
// *ast.Identifier's name → first token position into out. Only identifiers in
// reference position are visited; member names (`.Field`) and string contents
// are not *ast.Identifier nodes.
func collectIdentifierTokens(v reflect.Value, out map[string]token.Token) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		if id, ok := v.Interface().(*ast.Identifier); ok {
			if _, exists := out[id.Value]; !exists {
				out[id.Value] = id.Token
			}
			return
		}
		collectIdentifierTokens(v.Elem(), out)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			collectIdentifierTokens(v.Field(i), out)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			collectIdentifierTokens(v.Index(i), out)
		}
	case reflect.Interface:
		if v.IsNil() {
			return
		}
		collectIdentifierTokens(v.Elem(), out)
	}
}

// undefinedIdentifierWarnings returns warning diagnostics for every identifier
// used in the program that is neither declared in the symbol table nor a Kylix
// built-in. Declaration sites (function names, var names, params, fields) are
// themselves identifiers but are recorded in syms by CollectSymbols, so they
// are skipped.
// collectImplicitDecls records names that are declared implicitly — loop
// variables, enum members, generic type params, and lambda parameters — which
// the LSP SymbolTable does not collect. Added to the declared set so the
// undefined-identifier check never flags them.
func collectImplicitDecls(v reflect.Value, out map[string]bool) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		switch n := v.Interface().(type) {
		case *ast.ForStatement:
			if n.Variable != "" {
				out[strings.ToLower(n.Variable)] = true
			}
		case *ast.ForEachStatement:
			if n.Variable != "" {
				out[strings.ToLower(n.Variable)] = true
			}
		case *ast.EnumType:
			for _, nm := range n.Names {
				out[strings.ToLower(nm)] = true
			}
		case *ast.GenericType:
			for _, tp := range n.TypeParams {
				if id, ok := tp.(*ast.Identifier); ok {
					out[strings.ToLower(id.Value)] = true
				}
			}
		case *ast.LambdaExpression:
			for _, p := range n.Parameters {
				out[strings.ToLower(p.Name)] = true
			}
		}
		collectImplicitDecls(v.Elem(), out)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			collectImplicitDecls(v.Field(i), out)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			collectImplicitDecls(v.Index(i), out)
		}
	case reflect.Interface:
		if v.IsNil() {
			return
		}
		collectImplicitDecls(v.Elem(), out)
	}
}

func undefinedIdentifierWarnings(prog *ast.Program, syms *SymbolTable) []Diagnostic {
	if prog == nil || syms == nil {
		return nil
	}
	idents := make(map[string]token.Token)
	collectIdentifierTokens(reflect.ValueOf(prog), idents)

	// Case-insensitive set of every declared symbol (AllSymbols includes
	// function params / locals in child scopes, which Root.FindSymbol would
	// miss) plus implicit declarations.
	declared := make(map[string]bool, len(syms.AllSymbols))
	for _, s := range syms.AllSymbols {
		declared[strings.ToLower(s.Name)] = true
	}
	collectImplicitDecls(reflect.ValueOf(prog), declared)

	var diags []Diagnostic
	for name, tk := range idents {
		lname := strings.ToLower(name)
		if kylixBuiltinIdentifiers[lname] || declared[lname] {
			continue
		}
		line := tk.Line - 1
		col := tk.Column - 1
		if line < 0 {
			line = 0
		}
		if col < 0 {
			col = 0
		}
		diags = append(diags, Diagnostic{
			Range: Range{
				Start: Position{Line: line, Character: col},
				End:   Position{Line: line, Character: col + 1},
			},
			Severity: 2, // Warning
			Message:  fmt.Sprintf("Undefined identifier '%s'", name),
			Source:   "kylix",
		})
	}
	return diags
}
