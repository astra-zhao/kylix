package lsp

import (
	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/compiler"
	"strings"
	"sync"
)

// Document represents a parsed Kylix source file
type Document struct {
	URI         string
	Text        string
	Lines       []string
	AST         *ast.Program
	Symbols     *SymbolTable
	Diagnostics []Diagnostic
	ParseErrors []string
}

// NewDocument creates a new document from text
func NewDocument(uri, text string) *Document {
	doc := &Document{
		URI:         uri,
		Text:        text,
		Lines:       strings.Split(text, "\n"),
		Diagnostics: []Diagnostic{},
		ParseErrors: []string{},
	}

	// Parse the document
	l := lexer.New(text)
	p := parser.New(l)
	doc.AST = p.ParseProgram()
	doc.ParseErrors = p.Errors()

	// Collect diagnostics from parse errors
	for _, errMsg := range doc.ParseErrors {
		line, col := parseLocation(errMsg)
		doc.Diagnostics = append(doc.Diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: line, Character: col},
				End:   Position{Line: line, Character: col + 1},
			},
			Severity: 1, // Error
			Message:  errMsg,
		})
	}

	// Semantic diagnostics: interface implementation validation
	if len(doc.ParseErrors) == 0 && doc.AST != nil {
		sourcePath := uriToPath(uri)
		for _, cd := range compiler.CheckInterfaces(doc.AST, sourcePath) {
			line := cd.Line - 1   // LSP is 0-based
			col := cd.Column - 1
			if line < 0 {
				line = 0
			}
			if col < 0 {
				col = 0
			}
			doc.Diagnostics = append(doc.Diagnostics, Diagnostic{
				Range: Range{
					Start: Position{Line: line, Character: col},
					End:   Position{Line: line, Character: col + 1},
				},
				Severity: 1, // Error
				Message:  cd.Message,
				Source:   "kylix",
			})
		}
	}

	// Collect symbols if parsing succeeded
	if len(doc.ParseErrors) == 0 && doc.AST != nil {
		doc.Symbols = CollectSymbols(doc.AST)
	} else {
		doc.Symbols = NewSymbolTable()
	}

	return doc
}

// uriToPath converts a file:// URI to a local file path.
func uriToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		return uri[7:]
	}
	return uri
}

// GetLine returns the text of a specific line (0-indexed)
func (d *Document) GetLine(line int) string {
	if line < 0 || line >= len(d.Lines) {
		return ""
	}
	return d.Lines[line]
}

// GetWordAt returns the word at the given position
func (d *Document) GetWordAt(line, col int) string {
	lineText := d.GetLine(line)
	if lineText == "" || col >= len(lineText) {
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

// GetIdentifierAt returns the identifier at the given position
func (d *Document) GetIdentifierAt(line, col int) string {
	return d.GetWordAt(line, col)
}

// DocumentStore manages multiple documents
type DocumentStore struct {
	docs map[string]*Document
	mu   sync.RWMutex
}

// NewDocumentStore creates a new document store
func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		docs: make(map[string]*Document),
	}
}

// Update updates a document and re-parses it
func (ds *DocumentStore) Update(uri, text string) *Document {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	doc := NewDocument(uri, text)
	ds.docs[uri] = doc
	return doc
}

// Get retrieves a document
func (ds *DocumentStore) Get(uri string) *Document {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.docs[uri]
}

// Delete removes a document
func (ds *DocumentStore) Delete(uri string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.docs, uri)
}

// GetAll returns all documents
func (ds *DocumentStore) GetAll() []*Document {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	result := make([]*Document, 0, len(ds.docs))
	for _, doc := range ds.docs {
		result = append(result, doc)
	}
	return result
}
