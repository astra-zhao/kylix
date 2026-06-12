// handler_navigation.go — LSP navigation and refactoring handlers:
// definition, document symbols, references, rename, formatting,
// signature help, code actions, and workspace symbols.
package lsp

import (
	"encoding/json"
	"kylix/ast"
	"kylix/pkg/formatter"
	"strings"
)

// ── Definition ───────────────────────────────────────────────────────────────

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

	word := doc.GetWordAt(params.Position.Line, params.Position.Character)
	if word == "" {
		return &Message{ID: msg.ID, Result: nil}
	}

	sym := doc.Symbols.FindSymbol(word)
	if sym == nil {
		return &Message{ID: msg.ID, Result: nil}
	}

	return &Message{
		ID: msg.ID,
		Result: Location{
			URI: params.TextDocument.URI,
			Range: Range{
				Start: Position{Line: sym.Location.Line - 1, Character: sym.Location.Column - 1},
				End:   Position{Line: sym.Location.Line - 1, Character: sym.Location.Column - 1 + len(sym.Name)},
			},
		},
	}
}

// ── Document Symbols ─────────────────────────────────────────────────────────

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SymbolInformation represents a symbol in the document outline.
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
		if sym.Kind == SymbolParameter {
			continue // parameters clutter the outline
		}
		symbols = append(symbols, SymbolInformation{
			Name: sym.Name,
			Kind: symbolKindToDocumentSymbolKind(sym.Kind),
			Location: Location{
				URI: params.TextDocument.URI,
				Range: Range{
					Start: Position{Line: sym.Location.Line - 1, Character: sym.Location.Column - 1},
					End:   Position{Line: sym.Location.Line - 1, Character: sym.Location.Column - 1 + len(sym.Name)},
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

// ── References ───────────────────────────────────────────────────────────────

type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// ReferenceWalker walks the AST collecting all locations where targetName appears.
type ReferenceWalker struct {
	targetName string
	uri        string
	references []Location
}

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
		// Note: member name position is not tracked in the AST yet.
	}
}

func (s *Server) handleReferences(msg *Message) *Message {
	var params ReferenceParams
	json.Unmarshal(msg.Params, &params)

	doc := s.docs.Get(params.TextDocument.URI)
	if doc == nil || doc.AST == nil {
		return &Message{ID: msg.ID, Result: []Location{}}
	}

	identifier := doc.GetIdentifierAt(params.Position.Line, params.Position.Character)
	if identifier == "" {
		return &Message{ID: msg.ID, Result: []Location{}}
	}

	walker := &ReferenceWalker{
		targetName: identifier,
		uri:        params.TextDocument.URI,
		references: []Location{},
	}
	walker.Walk(doc.AST)
	return &Message{ID: msg.ID, Result: walker.references}
}

// ── Rename ───────────────────────────────────────────────────────────────────

type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}

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

	identifier := doc.GetIdentifierAt(params.Position.Line, params.Position.Character)
	if identifier == "" {
		return &Message{ID: msg.ID, Result: nil}
	}

	walker := &ReferenceWalker{
		targetName: identifier,
		uri:        params.TextDocument.URI,
		references: []Location{},
	}
	walker.Walk(doc.AST)

	edits := []TextEdit{}
	for _, ref := range walker.references {
		edits = append(edits, TextEdit{Range: ref.Range, NewText: params.NewName})
	}
	return &Message{
		ID:     msg.ID,
		Result: WorkspaceEdit{Changes: map[string][]TextEdit{params.TextDocument.URI: edits}},
	}
}

// ── Formatting ───────────────────────────────────────────────────────────────

type FormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

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

	f := formatter.New()
	formattedCode := f.Format(doc.AST)

	lines := strings.Split(doc.Text, "\n")
	lastLine := len(lines) - 1
	edit := TextEdit{
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: lastLine, Character: len(lines[lastLine])},
		},
		NewText: formattedCode,
	}
	return &Message{ID: msg.ID, Result: []TextEdit{edit}}
}

// ── Signature Help ───────────────────────────────────────────────────────────

type SignatureHelpParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation string                 `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters"`
}

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

	if params.Position.Line >= len(doc.Lines) {
		return &Message{ID: msg.ID, Result: nil}
	}
	lineText := doc.Lines[params.Position.Line]

	charPos := params.Position.Character
	if charPos > len(lineText) {
		charPos = len(lineText)
	}

	// Find the opening parenthesis before the cursor.
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

	// Extract function name immediately before '('.
	funcName := ""
	for i := parenPos - 1; i >= 0; i-- {
		ch := lineText[i]
		if isIdentChar(ch) {
			funcName = string(ch) + funcName
		} else {
			break
		}
	}
	if funcName == "" {
		return &Message{ID: msg.ID, Result: nil}
	}

	sym := doc.Symbols.FindSymbol(funcName)
	if sym == nil || (sym.Kind != SymbolFunction && sym.Kind != SymbolProcedure && sym.Kind != SymbolMethod) {
		return &Message{ID: msg.ID, Result: nil}
	}

	// Count commas to determine the active parameter index.
	activeParam := 0
	for i := parenPos + 1; i < charPos && i < len(lineText); i++ {
		if lineText[i] == ',' {
			activeParam++
		}
	}

	signature := SignatureInformation{
		Label:      sym.Name + "(" + sym.Type + ")",
		Parameters: []ParameterInformation{},
	}
	for _, child := range sym.Children {
		if child.Kind == SymbolParameter {
			signature.Parameters = append(signature.Parameters, ParameterInformation{
				Label: child.Name + ": " + child.Type,
			})
		}
	}

	return &Message{
		ID: msg.ID,
		Result: SignatureHelp{
			Signatures:      []SignatureInformation{signature},
			ActiveSignature: 0,
			ActiveParameter: activeParam,
		},
	}
}

// ── Code Actions ─────────────────────────────────────────────────────────────

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type CodeAction struct {
	Title   string      `json:"title"`
	Kind    string      `json:"kind"`
	Edit    interface{} `json:"edit,omitempty"`
	Command interface{} `json:"command,omitempty"`
}

func (s *Server) handleCodeAction(msg *Message) *Message {
	actions := []CodeAction{
		{Title: "Organize Imports", Kind: "source.organizeImports"},
		{Title: "Format Document", Kind: "source.format"},
	}
	return &Message{ID: msg.ID, Result: actions}
}

// ── Workspace Symbols ────────────────────────────────────────────────────────

type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

func (s *Server) handleWorkspaceSymbol(msg *Message) *Message {
	var params WorkspaceSymbolParams
	json.Unmarshal(msg.Params, &params)

	query := strings.ToLower(params.Query)
	symbols := []SymbolInformation{}

	for _, doc := range s.docs.GetAll() {
		if doc.Symbols == nil {
			continue
		}
		for _, sym := range doc.Symbols.AllSymbols {
			if sym.Kind == SymbolParameter {
				continue
			}
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
