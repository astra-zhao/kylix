package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"kylix/ast"
	"kylix/pkg/formatter"
	"os"
	"strings"
)

// Server implements a minimal LSP server over stdin/stdout
type Server struct {
	in  io.Reader
	out io.Writer

	// document store: URI → Document (with AST, symbols, diagnostics)
	docs *DocumentStore
}

// New creates a new LSP server
func New(in io.Reader, out io.Writer) *Server {
	return &Server{
		in:   in,
		out:  out,
		docs: NewDocumentStore(),
	}
}

// Run starts the LSP server main loop, reading JSON-RPC messages from stdin
func (s *Server) Run() error {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		response := s.handleMessage(msg)
		if response != nil {
			if err := s.writeMessage(response); err != nil {
				return err
			}
		}
	}
}

// Message represents a JSON-RPC 2.0 message
type Message struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *int             `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *ResponseError   `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Server) readMessage() (*Message, error) {
	// Read headers
	var contentLength int
	for {
		var line string
		var c [1]byte
		for {
			n, err := s.in.Read(c[:])
			if err != nil {
				return nil, err
			}
			if n == 0 {
				continue
			}
			if c[0] == '\n' {
				break
			}
			if c[0] != '\r' {
				line += string(c[0])
			}
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		if strings.HasPrefix(line, "Content-Length:") {
			fmt.Sscanf(strings.TrimPrefix(line, "Content-Length:"), "%d", &contentLength)
		}
	}

	if contentLength == 0 {
		return nil, io.EOF
	}

	// Read body
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(s.in, body); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (s *Server) writeMessage(msg *Message) error {
	msg.JSONRPC = "2.0"
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := s.out.Write([]byte(header)); err != nil {
		return err
	}
	_, err = s.out.Write(body)
	return err
}

func (s *Server) handleMessage(msg *Message) *Message {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return nil // notification, no response
	case "shutdown":
		return s.handleShutdown(msg)
	case "exit":
		os.Exit(0)
		return nil
	case "textDocument/didOpen":
		s.handleDidOpen(msg)
		return nil
	case "textDocument/didChange":
		s.handleDidChange(msg)
		return nil
	case "textDocument/didClose":
		s.handleDidClose(msg)
		return nil
	case "textDocument/completion":
		return s.handleCompletion(msg)
	case "textDocument/hover":
		return s.handleHover(msg)
	case "textDocument/definition":
		return s.handleDefinition(msg)
	case "textDocument/documentSymbol":
		return s.handleDocumentSymbol(msg)
	case "textDocument/references":
		return s.handleReferences(msg)
	case "textDocument/rename":
		return s.handleRename(msg)
	case "textDocument/formatting":
		return s.handleFormatting(msg)
	case "textDocument/signatureHelp":
		return s.handleSignatureHelp(msg)
	case "textDocument/codeAction":
		return s.handleCodeAction(msg)
	case "workspace/symbol":
		return s.handleWorkspaceSymbol(msg)
	default:
		if msg.ID != nil {
			return &Message{
				ID: msg.ID,
				Error: &ResponseError{
					Code:    -32601,
					Message: fmt.Sprintf("method not found: %s", msg.Method),
				},
			}
		}
		return nil
	}
}

func (s *Server) handleInitialize(msg *Message) *Message {
	return &Message{
		ID: msg.ID,
		Result: map[string]interface{}{
			"capabilities": map[string]interface{}{
				"textDocumentSync": 1, // Full sync
				"completionProvider": map[string]interface{}{
					"triggerCharacters": []string{".", ":"},
					"resolveProvider":   false,
				},
				"hoverProvider":          true,
				"definitionProvider":     true,
				"documentSymbolProvider": true,
				"referencesProvider":     true,
				"renameProvider":         true,
				"documentFormattingProvider": true,
				"signatureHelpProvider": map[string]interface{}{
					"triggerCharacters": []string{"(", ","},
				},
				"codeActionProvider":  true,
				"workspaceSymbolProvider": true,
			},
			"serverInfo": map[string]interface{}{
				"name":    "kylix-lsp",
				"version": "0.3.0",
			},
		},
	}
}

func (s *Server) handleShutdown(msg *Message) *Message {
	return &Message{ID: msg.ID, Result: nil}
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeParams struct {
	TextDocument   VersionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []TextDocumentContentChange     `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChange struct {
	Text string `json:"text"`
}

type DidCloseParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

func (s *Server) handleDidOpen(msg *Message) {
	var params DidOpenParams
	json.Unmarshal(msg.Params, &params)
	doc := s.docs.Update(params.TextDocument.URI, params.TextDocument.Text)
	s.publishDiagnostics(doc)
}

func (s *Server) handleDidChange(msg *Message) {
	var params DidChangeParams
	json.Unmarshal(msg.Params, &params)
	for _, change := range params.ContentChanges {
		doc := s.docs.Update(params.TextDocument.URI, change.Text)
		s.publishDiagnostics(doc)
	}
}

func (s *Server) handleDidClose(msg *Message) {
	var params DidCloseParams
	json.Unmarshal(msg.Params, &params)
	s.docs.Delete(params.TextDocument.URI)
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1=error, 2=warning, 3=info, 4=hint
	Message  string `json:"message"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

func (s *Server) publishDiagnostics(doc *Document) {
	s.writeMessage(&Message{
		Method: "textDocument/publishDiagnostics",
		Params: json.RawMessage(mustMarshal(PublishDiagnosticsParams{
			URI:         doc.URI,
			Diagnostics: doc.Diagnostics,
		})),
	})
}

func parseLocation(msg string) (int, int) {
	var line, col int
	if idx := strings.Index(msg, "(line "); idx >= 0 {
		fmt.Sscanf(msg[idx:], "(line %d, column %d)", &line, &col)
	}
	return line - 1, col - 1 // LSP uses 0-based indexing
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// Completion handling

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

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
		items = append(items, CompletionItem{
			Label: kw,
			Kind:  14, // Keyword
		})
	}

	// Builtin function completions
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

	// Type completions
	types := []string{"Integer", "Real", "Boolean", "String", "Char", "Byte"}
	for _, t := range types {
		items = append(items, CompletionItem{
			Label:  t,
			Kind:   6, // Variable (type)
			Detail: "type",
		})
	}

	// Symbol completions from document
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

	return &Message{
		ID:     msg.ID,
		Result: items,
	}
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

// Hover handling

type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

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

	// Get the word at the cursor position
	word := doc.GetWordAt(params.Position.Line, params.Position.Character)
	if word == "" {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Look up documentation from symbols
	if doc.Symbols != nil {
		sym := doc.Symbols.FindSymbol(word)
		if sym != nil {
			return &Message{
				ID: msg.ID,
				Result: Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: formatSymbolHover(sym),
					},
				},
			}
		}
	}

	// Look up built-in documentation
	docText := lookupDocumentation(word)
	if docText == "" {
		return &Message{ID: msg.ID, Result: nil}
	}

	return &Message{
		ID: msg.ID,
		Result: Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: docText,
			},
		},
	}
}

func formatSymbolHover(sym *Symbol) string {
	var result strings.Builder

	switch sym.Kind {
	case SymbolVariable:
		result.WriteString("**Variable** `")
		result.WriteString(sym.Name)
		result.WriteString("`: ")
		result.WriteString(sym.Type)
	case SymbolConstant:
		result.WriteString("**Constant** `")
		result.WriteString(sym.Name)
		result.WriteString("`: ")
		result.WriteString(sym.Type)
	case SymbolFunction:
		result.WriteString("**Function** `")
		result.WriteString(sym.Name)
		result.WriteString("`")
		if sym.Type != "" {
			result.WriteString(": ")
			result.WriteString(sym.Type)
		}
	case SymbolProcedure:
		result.WriteString("**Procedure** `")
		result.WriteString(sym.Name)
		result.WriteString("`")
	case SymbolClass:
		result.WriteString("**Class** `")
		result.WriteString(sym.Name)
		result.WriteString("`")
	case SymbolInterface:
		result.WriteString("**Interface** `")
		result.WriteString(sym.Name)
		result.WriteString("`")
	case SymbolMethod:
		result.WriteString("**Method** `")
		result.WriteString(sym.Name)
		result.WriteString("`")
		if sym.Type != "" {
			result.WriteString(": ")
			result.WriteString(sym.Type)
		}
	case SymbolField:
		result.WriteString("**Field** `")
		result.WriteString(sym.Name)
		result.WriteString("`: ")
		result.WriteString(sym.Type)
	case SymbolProperty:
		result.WriteString("**Property** `")
		result.WriteString(sym.Name)
		result.WriteString("`: ")
		result.WriteString(sym.Type)
	case SymbolParameter:
		result.WriteString("**Parameter** `")
		result.WriteString(sym.Name)
		result.WriteString("`: ")
		result.WriteString(sym.Type)
	case SymbolType:
		result.WriteString("**Type** `")
		result.WriteString(sym.Name)
		result.WriteString("`")
	default:
		result.WriteString("`")
		result.WriteString(sym.Name)
		result.WriteString("`")
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

	// Find word boundaries
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

// DefinitionParams represents the parameters for a definition request
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

func (s *Server) handleDefinition(msg *Message) *Message {
	var params DefinitionParams
	json.Unmarshal(msg.Params, &params)

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil || doc.Symbols == nil {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Get the word at the cursor position
	word := doc.GetWordAt(params.Position.Line, params.Position.Character)
	if word == "" {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Find the symbol
	sym := doc.Symbols.FindSymbol(word)
	if sym == nil {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Return the location of the symbol definition
	return &Message{
		ID: msg.ID,
		Result: Location{
			URI: params.TextDocument.URI,
			Range: Range{
				Start: Position{
					Line:      sym.Location.Line - 1,
					Character: sym.Location.Column - 1,
				},
				End: Position{
					Line:      sym.Location.Line - 1,
					Character: sym.Location.Column - 1 + len(sym.Name),
				},
			},
		},
	}
}

// DocumentSymbolParams represents the parameters for a document symbol request
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SymbolInformation represents a symbol in the document
type SymbolInformation struct {
	Name          string   `json:"name"`
	Kind          int      `json:"kind"`
	Location      Location `json:"location"`
	ContainerName string   `json:"containerName,omitempty"`
}

func (s *Server) handleDocumentSymbol(msg *Message) *Message {
	var params DocumentSymbolParams
	json.Unmarshal(msg.Params, &params)

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil || doc.Symbols == nil {
		return &Message{ID: msg.ID, Result: []SymbolInformation{}}
	}

	symbols := []SymbolInformation{}
	for _, sym := range doc.Symbols.AllSymbols {
		// Skip parameters as they clutter the outline
		if sym.Kind == SymbolParameter {
			continue
		}

		symbols = append(symbols, SymbolInformation{
			Name: sym.Name,
			Kind: symbolKindToDocumentSymbolKind(sym.Kind),
			Location: Location{
				URI: params.TextDocument.URI,
				Range: Range{
					Start: Position{
						Line:      sym.Location.Line - 1,
						Character: sym.Location.Column - 1,
					},
					End: Position{
						Line:      sym.Location.Line - 1,
						Character: sym.Location.Column - 1 + len(sym.Name),
					},
				},
			},
		})
	}

	return &Message{ID: msg.ID, Result: symbols}
}

func symbolKindToDocumentSymbolKind(kind SymbolKind) int {
	switch kind {
	case SymbolVariable:
		return 13 // Variable
	case SymbolConstant:
		return 14 // Constant
	case SymbolType:
		return 22 // Struct
	case SymbolFunction, SymbolProcedure:
		return 12 // Function
	case SymbolClass:
		return 5 // Class
	case SymbolInterface:
		return 11 // Interface
	case SymbolMethod:
		return 6 // Method
	case SymbolField:
		return 8 // Field
	case SymbolProperty:
		return 7 // Property
	case SymbolParameter:
		return 13 // Variable
	default:
		return 1 // File
	}
}

// ReferenceParams represents the parameters for a reference request
type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

// ReferenceContext contains context for a reference request
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// ReferenceWalker walks the AST to find all references to a symbol
type ReferenceWalker struct {
	targetName string
	uri        string
	references []Location
}

// Walk walks the AST and collects all references
func (w *ReferenceWalker) Walk(node ast.Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.Program:
		for _, decl := range n.Declarations {
			w.Walk(decl)
		}
		for _, stmt := range n.Statements {
			w.Walk(stmt)
		}

	case *ast.VarDecl:
		for _, name := range n.Names {
			if name == w.targetName {
				w.references = append(w.references, Location{
					URI: w.uri,
					Range: Range{
						Start: Position{Line: n.Token.Line - 1, Character: n.Token.Column - 1},
						End:   Position{Line: n.Token.Line - 1, Character: n.Token.Column - 1 + len(name)},
					},
				})
			}
		}

	case *ast.FunctionDecl:
		if n.Name == w.targetName {
			w.references = append(w.references, Location{
				URI: w.uri,
				Range: Range{
					Start: Position{Line: n.Token.Line - 1, Character: n.Token.Column - 1},
					End:   Position{Line: n.Token.Line - 1, Character: n.Token.Column - 1 + len(n.Name)},
				},
			})
		}
		for _, param := range n.Parameters {
			if param.Name == w.targetName {
				w.references = append(w.references, Location{
					URI: w.uri,
					Range: Range{
						Start: Position{Line: param.Token.Line - 1, Character: param.Token.Column - 1},
						End:   Position{Line: param.Token.Line - 1, Character: param.Token.Column - 1 + len(param.Name)},
					},
				})
			}
		}
		if n.Body != nil {
			w.Walk(n.Body)
		}

	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			w.Walk(stmt)
		}

	case *ast.ExpressionStatement:
		w.Walk(n.Expression)

	case *ast.ReturnStatement:
		if n.Value != nil {
			w.Walk(n.Value)
		}

	case *ast.IfStatement:
		w.Walk(n.Condition)
		w.Walk(n.Consequence)
		if n.Alternative != nil {
			w.Walk(n.Alternative)
		}

	case *ast.WhileStatement:
		w.Walk(n.Condition)
		w.Walk(n.Body)

	case *ast.ForStatement:
		if n.Variable == w.targetName {
			w.references = append(w.references, Location{
				URI: w.uri,
				Range: Range{
					Start: Position{Line: n.Token.Line - 1, Character: n.Token.Column - 1},
					End:   Position{Line: n.Token.Line - 1, Character: n.Token.Column - 1 + len(n.Variable)},
				},
			})
		}
		w.Walk(n.From)
		w.Walk(n.To)
		w.Walk(n.Body)

	case *ast.AssignmentStatement:
		w.Walk(n.Name)
		w.Walk(n.Value)

	case *ast.Identifier:
		if n.Value == w.targetName {
			w.references = append(w.references, Location{
				URI: w.uri,
				Range: Range{
					Start: Position{Line: n.Token.Line - 1, Character: n.Token.Column - 1},
					End:   Position{Line: n.Token.Line - 1, Character: n.Token.Column - 1 + len(n.Value)},
				},
			})
		}

	case *ast.CallExpression:
		w.Walk(n.Function)
		for _, arg := range n.Arguments {
			w.Walk(arg)
		}

	case *ast.InfixExpression:
		w.Walk(n.Left)
		w.Walk(n.Right)

	case *ast.PrefixExpression:
		w.Walk(n.Right)

	case *ast.IndexExpression:
		w.Walk(n.Left)
		w.Walk(n.Index)

	case *ast.MemberExpression:
		w.Walk(n.Object)
		if n.Member == w.targetName {
			// Note: MemberExpression doesn't have position for member name
			// This is a limitation we'll need to fix in the AST
		}
	}
}

func (s *Server) handleReferences(msg *Message) *Message {
	var params ReferenceParams
	json.Unmarshal(msg.Params, &params)

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil || doc.AST == nil {
		return &Message{ID: msg.ID, Result: []Location{}}
	}

	// Get the identifier at the position
	identifier := doc.GetIdentifierAt(params.Position.Line, params.Position.Character)
	if identifier == "" {
		return &Message{ID: msg.ID, Result: []Location{}}
	}

	// Walk the AST to find all references
	walker := &ReferenceWalker{
		targetName: identifier,
		uri:        params.TextDocument.URI,
		references: []Location{},
	}
	walker.Walk(doc.AST)

	return &Message{ID: msg.ID, Result: walker.references}
}

// RenameParams represents the parameters for a rename request
type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}

// WorkspaceEdit represents a workspace edit
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes"`
}

func (s *Server) handleRename(msg *Message) *Message {
	var params RenameParams
	json.Unmarshal(msg.Params, &params)

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil || doc.AST == nil {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Get the identifier at the position
	identifier := doc.GetIdentifierAt(params.Position.Line, params.Position.Character)
	if identifier == "" {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Find all references to the identifier
	walker := &ReferenceWalker{
		targetName: identifier,
		uri:        params.TextDocument.URI,
		references: []Location{},
	}
	walker.Walk(doc.AST)

	// Create text edits for all references
	edits := []TextEdit{}
	for _, ref := range walker.references {
		edits = append(edits, TextEdit{
			Range:   ref.Range,
			NewText: params.NewName,
		})
	}

	workspaceEdit := WorkspaceEdit{
		Changes: map[string][]TextEdit{
			params.TextDocument.URI: edits,
		},
	}

	return &Message{ID: msg.ID, Result: workspaceEdit}
}

// FormattingParams represents the parameters for a formatting request
type FormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions contains formatting options
type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

func (s *Server) handleFormatting(msg *Message) *Message {
	var params FormattingParams
	json.Unmarshal(msg.Params, &params)

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil || doc.AST == nil {
		return &Message{ID: msg.ID, Result: []TextEdit{}}
	}

	// Use the formatter to format the document
	f := formatter.New()
	formattedCode := f.Format(doc.AST)

	// Return a single text edit that replaces the entire document
	lines := strings.Split(doc.Text, "\n")
	lastLine := len(lines) - 1
	lastChar := len(lines[lastLine])

	edit := TextEdit{
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: lastLine, Character: lastChar},
		},
		NewText: formattedCode,
	}

	return &Message{ID: msg.ID, Result: []TextEdit{edit}}
}

// SignatureHelpParams represents the parameters for a signature help request
type SignatureHelpParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// SignatureHelp represents signature help information
type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

// SignatureInformation represents information about a callable signature
type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation string                 `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters"`
}

// ParameterInformation represents information about a parameter
type ParameterInformation struct {
	Label         string `json:"label"`
	Documentation string `json:"documentation,omitempty"`
}

func (s *Server) handleSignatureHelp(msg *Message) *Message {
	var params SignatureHelpParams
	json.Unmarshal(msg.Params, &params)

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil || doc.Symbols == nil {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Get the line text
	if params.Position.Line >= len(doc.Lines) {
		return &Message{ID: msg.ID, Result: nil}
	}
	lineText := doc.Lines[params.Position.Line]

	// Find the function name before the opening parenthesis
	// Simple heuristic: look backwards from cursor position
	charPos := params.Position.Character
	if charPos > len(lineText) {
		charPos = len(lineText)
	}

	// Find the opening parenthesis
	parenPos := -1
	for i := charPos - 1; i >= 0; i-- {
		if lineText[i] == '(' {
			parenPos = i
			break
		}
	}

	if parenPos == -1 {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Extract the function name
	funcName := ""
	for i := parenPos - 1; i >= 0; i-- {
		ch := lineText[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			funcName = string(ch) + funcName
		} else {
			break
		}
	}

	if funcName == "" {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Find the function symbol
	sym := doc.Symbols.FindSymbol(funcName)
	if sym == nil || (sym.Kind != SymbolFunction && sym.Kind != SymbolProcedure && sym.Kind != SymbolMethod) {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Count commas to determine active parameter
	activeParam := 0
	for i := parenPos + 1; i < charPos; i++ {
		if i < len(lineText) && lineText[i] == ',' {
			activeParam++
		}
	}

	// Build signature information
	signature := SignatureInformation{
		Label:      sym.Name + "(" + sym.Type + ")",
		Parameters: []ParameterInformation{},
	}

	// Extract parameters from the function's children
	for _, child := range sym.Children {
		if child.Kind == SymbolParameter {
			signature.Parameters = append(signature.Parameters, ParameterInformation{
				Label: child.Name + ": " + child.Type,
			})
		}
	}

	sigHelp := SignatureHelp{
		Signatures:      []SignatureInformation{signature},
		ActiveSignature: 0,
		ActiveParameter: activeParam,
	}

	return &Message{ID: msg.ID, Result: sigHelp}
}

// CodeActionParams represents the parameters for a code action request
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

// CodeActionContext contains context for code actions
type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// CodeAction represents a code action
type CodeAction struct {
	Title   string      `json:"title"`
	Kind    string      `json:"kind"`
	Edit    interface{} `json:"edit,omitempty"`
	Command interface{} `json:"command,omitempty"`
}

func (s *Server) handleCodeAction(msg *Message) *Message {
	var params CodeActionParams
	json.Unmarshal(msg.Params, &params)

	actions := []CodeAction{}

	// Add organize imports action
	actions = append(actions, CodeAction{
		Title: "Organize Imports",
		Kind:  "source.organizeImports",
	})

	// Add format document action
	actions = append(actions, CodeAction{
		Title: "Format Document",
		Kind:  "source.format",
	})

	return &Message{ID: msg.ID, Result: actions}
}

// WorkspaceSymbolParams represents the parameters for a workspace symbol request
type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

func (s *Server) handleWorkspaceSymbol(msg *Message) *Message {
	var params WorkspaceSymbolParams
	json.Unmarshal(msg.Params, &params)

	symbols := []SymbolInformation{}
	query := strings.ToLower(params.Query)

	// Search through all documents
	for _, doc := range s.docs.GetAll() {
		if doc.Symbols == nil {
			continue
		}

		for _, sym := range doc.Symbols.AllSymbols {
			// Skip parameters
			if sym.Kind == SymbolParameter {
				continue
			}

			// Filter by query
			if query != "" && !strings.Contains(strings.ToLower(sym.Name), query) {
				continue
			}

			symbols = append(symbols, SymbolInformation{
				Name: sym.Name,
				Kind: symbolKindToDocumentSymbolKind(sym.Kind),
				Location: Location{
					URI: doc.URI,
					Range: Range{
						Start: Position{Line: sym.Location.Line - 1, Character: sym.Location.Column - 1},
						End:   Position{Line: sym.Location.Line - 1, Character: sym.Location.Column - 1 + len(sym.Name)},
					},
				},
			})
		}
	}

	return &Message{ID: msg.ID, Result: symbols}
}
