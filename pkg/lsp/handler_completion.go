// handler_completion.go — LSP completion and hover handlers.
package lsp

import (
	"encoding/json"
	"strings"
)

// CompletionParams represents the parameters for a completion request.
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"` // 1=text, 2=method, 3=function, 6=variable, 14=keyword
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

func (s *Server) handleCompletion(msg *Message) *Message {
	var params CompletionParams
	json.Unmarshal(msg.Params, &params)

	items := []CompletionItem{}

	// Keyword completions
	keywords := []string{
		"program", "unit", "uses", "var", "const", "type", "begin", "end",
		"function", "procedure", "if", "then", "else", "while", "do", "for",
		"to", "downto", "repeat", "until", "case", "of", "with", "try",
		"except", "finally", "raise", "class", "interface", "record", "array",
		"public", "private", "protected", "property", "read", "write",
		"virtual", "override", "abstract", "async", "await", "match", "when",
		"import", "export", "module", "return", "break", "continue", "exit",
		"and", "or", "not", "xor", "in", "is", "as", "new", "nil", "true", "false",
	}
	for _, kw := range keywords {
		items = append(items, CompletionItem{Label: kw, Kind: 14})
	}

	// Built-in function completions
	builtins := []CompletionItem{
		{Label: "WriteLn", Kind: 3, Detail: "procedure", Documentation: "Write to stdout with newline", InsertText: "WriteLn("},
		{Label: "Write", Kind: 3, Detail: "procedure", Documentation: "Write to stdout", InsertText: "Write("},
		{Label: "ReadLn", Kind: 3, Detail: "procedure", Documentation: "Read from stdin", InsertText: "ReadLn("},
		{Label: "Length", Kind: 3, Detail: "function", Documentation: "Get length of array or string", InsertText: "Length("},
		{Label: "IntToStr", Kind: 3, Detail: "function", Documentation: "Convert integer to string", InsertText: "IntToStr("},
		{Label: "StrToInt", Kind: 3, Detail: "function", Documentation: "Convert string to integer", InsertText: "StrToInt("},
		{Label: "Copy", Kind: 3, Detail: "function", Documentation: "Copy substring", InsertText: "Copy("},
		{Label: "Concat", Kind: 3, Detail: "function", Documentation: "Concatenate strings", InsertText: "Concat("},
		{Label: "UpperCase", Kind: 3, Detail: "function", Documentation: "Convert to uppercase", InsertText: "UpperCase("},
		{Label: "LowerCase", Kind: 3, Detail: "function", Documentation: "Convert to lowercase", InsertText: "LowerCase("},
		{Label: "Sqrt", Kind: 3, Detail: "function", Documentation: "Square root", InsertText: "Sqrt("},
		{Label: "Abs", Kind: 3, Detail: "function", Documentation: "Absolute value", InsertText: "Abs("},
		{Label: "Round", Kind: 3, Detail: "function", Documentation: "Round to nearest integer", InsertText: "Round("},
	}
	items = append(items, builtins...)

	// Built-in type completions
	types := []string{"Integer", "Real", "Boolean", "String", "Char", "Byte"}
	for _, t := range types {
		items = append(items, CompletionItem{Label: t, Kind: 6, Detail: "type"})
	}

	// Symbol completions from the open document
	doc := s.docs.Get(params.TextDocument.URI)
	if doc != nil && doc.Symbols != nil {
		for _, sym := range doc.Symbols.AllSymbols {
			items = append(items, CompletionItem{
				Label:  sym.Name,
				Kind:   symbolKindToCompletionKind(sym.Kind),
				Detail: sym.Type,
			})
		}
	}

	return &Message{ID: msg.ID, Result: items}
}

func symbolKindToCompletionKind(kind SymbolKind) int {
	switch kind {
	case SymbolVariable:
		return 6 // Variable
	case SymbolConstant:
		return 21 // Constant
	case SymbolType:
		return 22 // Struct
	case SymbolFunction, SymbolProcedure:
		return 3 // Function
	case SymbolClass:
		return 7 // Class
	case SymbolInterface:
		return 8 // Interface
	case SymbolMethod:
		return 2 // Method
	case SymbolField:
		return 5 // Field
	case SymbolProperty:
		return 10 // Property
	case SymbolParameter:
		return 6 // Variable
	default:
		return 1 // Text
	}
}

// ── Hover ────────────────────────────────────────────────────────────────────

// HoverParams represents the parameters for a hover request.
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Hover is the LSP hover response.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent holds markdown or plain-text hover content.
type MarkupContent struct {
	Kind  string `json:"kind"` // "plaintext" or "markdown"
	Value string `json:"value"`
}

func (s *Server) handleHover(msg *Message) *Message {
	var params HoverParams
	json.Unmarshal(msg.Params, &params)

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil {
		return &Message{ID: msg.ID, Result: nil}
	}

	word := doc.GetWordAt(params.Position.Line, params.Position.Character)
	if word == "" {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Symbol from document takes priority over built-in docs.
	if doc.Symbols != nil {
		sym := doc.Symbols.FindSymbol(word)
		if sym != nil {
			return &Message{
				ID: msg.ID,
				Result: Hover{
					Contents: MarkupContent{Kind: "markdown", Value: formatSymbolHover(sym)},
				},
			}
		}
	}

	docText := lookupDocumentation(word)
	if docText == "" {
		return &Message{ID: msg.ID, Result: nil}
	}
	return &Message{
		ID:     msg.ID,
		Result: Hover{Contents: MarkupContent{Kind: "markdown", Value: docText}},
	}
}

// formatSymbolHover builds a markdown hover string for a symbol.
func formatSymbolHover(sym *Symbol) string {
	var result strings.Builder

	switch sym.Kind {
	case SymbolVariable:
		result.WriteString("**Variable** `" + sym.Name + "`: " + sym.Type)
	case SymbolConstant:
		result.WriteString("**Constant** `" + sym.Name + "`: " + sym.Type)
	case SymbolFunction:
		result.WriteString("**Function** `" + sym.Name + "`")
		if sym.Type != "" {
			result.WriteString(": " + sym.Type)
		}
	case SymbolProcedure:
		result.WriteString("**Procedure** `" + sym.Name + "`")
	case SymbolClass:
		result.WriteString("**Class** `" + sym.Name + "`")
	case SymbolInterface:
		result.WriteString("**Interface** `" + sym.Name + "`")
	case SymbolMethod:
		result.WriteString("**Method** `" + sym.Name + "`")
		if sym.Type != "" {
			result.WriteString(": " + sym.Type)
		}
	case SymbolField:
		result.WriteString("**Field** `" + sym.Name + "`: " + sym.Type)
	case SymbolProperty:
		result.WriteString("**Property** `" + sym.Name + "`: " + sym.Type)
	case SymbolParameter:
		result.WriteString("**Parameter** `" + sym.Name + "`: " + sym.Type)
	case SymbolType:
		result.WriteString("**Type** `" + sym.Name + "`")
	default:
		result.WriteString("`" + sym.Name + "`")
	}

	return result.String()
}

func getWordAt(text string, line, col int) string {
	lines := strings.Split(text, "\n")
	if line >= len(lines) {
		return ""
	}
	lineText := lines[line]
	if col >= len(lineText) {
		return ""
	}
	start := col
	for start > 0 && isIdentChar(lineText[start-1]) {
		start--
	}
	end := col
	for end < len(lineText) && isIdentChar(lineText[end]) {
		end++
	}
	if start == end {
		return ""
	}
	return lineText[start:end]
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// lookupDocumentation returns built-in markdown documentation for a keyword or built-in name.
func lookupDocumentation(word string) string {
	docs := map[string]string{
		"WriteLn":   "**WriteLn**(...)\n\nWrite values to stdout followed by a newline.",
		"Write":     "**Write**(...)\n\nWrite values to stdout.",
		"ReadLn":    "**ReadLn**(...)\n\nRead a line from stdin.",
		"Length":    "**Length**(s: String|Array): Integer\n\nReturns the length of a string or array.",
		"IntToStr":  "**IntToStr**(i: Integer): String\n\nConverts an integer to its string representation.",
		"StrToInt":  "**StrToInt**(s: String): Integer\n\nParses a string as an integer.",
		"Copy":      "**Copy**(s: String; start, count: Integer): String\n\nReturns a substring.",
		"Concat":    "**Concat**(s1, s2: String): String\n\nConcatenates two strings.",
		"UpperCase": "**UpperCase**(s: String): String\n\nConverts string to uppercase.",
		"LowerCase": "**LowerCase**(s: String): String\n\nConverts string to lowercase.",
		"Sqrt":      "**Sqrt**(x: Real): Real\n\nReturns the square root of x.",
		"Abs":       "**Abs**(x: Integer|Real): Integer|Real\n\nReturns the absolute value.",
		"Round":     "**Round**(x: Real): Integer\n\nRounds to the nearest integer.",
		"Integer":   "**Integer**\n\n64-bit signed integer type.",
		"Real":      "**Real**\n\n64-bit floating-point type.",
		"Boolean":   "**Boolean**\n\nTrue or false value.",
		"String":    "**String**\n\nSequence of characters.",
		"Char":      "**Char**\n\nSingle character (byte).",
		"function":  "**function** name(params): ReturnType;\n\nDeclares a function with a return value.",
		"procedure": "**procedure** name(params);\n\nDeclares a procedure (no return value).",
		"class":     "**class** name [inherits Parent] [implements Interfaces]\n\nDeclares a class.",
		"interface": "**interface** name [extends Parents]\n\nDeclares an interface.",
		"begin":     "**begin** ... **end**\n\nDefines a block of statements.",
		"if":        "**if** condition **then** ... [**else** ...]\n\nConditional statement.",
		"while":     "**while** condition **do** ...\n\nLoop while condition is true.",
		"for":       "**for** i := start **to** end **do** ...\n\nCounted loop.",
		"match":     "**match** value { pattern => result, ... }\n\nPattern matching expression.",
		"try":       "**try** ... **except** ... **finally** ... **end**\n\nException handling.",
		"async":     "**async function** name(): ReturnType;\n\nDeclares an asynchronous function.",
		"await":     "**await** expression\n\nAwaits the result of an async operation.",
		"nil":       "**nil**\n\nThe null/empty value.",
		"true":      "**true**\n\nBoolean true literal.",
		"false":     "**false**\n\nBoolean false literal.",
	}
	return docs[word]
}
