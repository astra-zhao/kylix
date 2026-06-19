package lsp

import (
	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/compiler"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Document represents a parsed Kylix source file
type Document struct {
	URI         string
	Text        string
	Version     int // Latest LSP version applied; -1 for unset
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
			line := cd.Line - 1
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
				Severity: 1,
				Message:  cd.Message,
				Source:   "kylix",
			})
		}
	}

	// Collect symbols if parsing succeeded
	if len(doc.ParseErrors) == 0 && doc.AST != nil {
		doc.Symbols = CollectSymbols(doc.AST)
		// Enrich with stdlib declarations for used modules
		loadStdlibSymbols(doc)
	} else {
		doc.Symbols = NewSymbolTable()
	}

	return doc
}

// loadStdlibSymbols finds stdlib/klx/<module>.klx files for each uses clause
// entry and merges their symbols into doc.Symbols.
func loadStdlibSymbols(doc *Document) {
	if doc.AST == nil || len(doc.AST.Uses) == 0 {
		return
	}
	klxDir := findStdlibKlxDir()
	if klxDir == "" {
		return
	}
	for _, mod := range doc.AST.Uses {
		path := filepath.Join(klxDir, mod+".klx")
		if _, err := os.Stat(path); err != nil {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		l := lexer.New(string(src))
		p := parser.New(l)
		prog := p.ParseProgram()
		if len(p.Errors()) > 0 {
			continue
		}
		unitSyms := CollectSymbols(prog)
		// Merge: add both qualified (mod.Name) and unqualified symbols.
		for _, sym := range unitSyms.AllSymbols {
			qualified := &Symbol{
				Name:     mod + "." + sym.Name,
				Kind:     sym.Kind,
				Type:     sym.Type,
				Location: sym.Location,
			}
			doc.Symbols.AllSymbols = append(doc.Symbols.AllSymbols, qualified)
			doc.Symbols.AllSymbols = append(doc.Symbols.AllSymbols, sym)
		}
	}
}

// findStdlibKlxDir locates the stdlib/klx directory by checking:
// 1. $KYLIX_HOME/stdlib/klx
// 2. directory of the running executable ± stdlib/klx
// 3. well-known development path relative to the executable
func findStdlibKlxDir() string {
	if home := os.Getenv("KYLIX_HOME"); home != "" {
		d := filepath.Join(home, "stdlib", "klx")
		if _, err := os.Stat(d); err == nil {
			return d
		}
	}
	exe, err := exec.LookPath(os.Args[0])
	if err == nil {
		// Walk up from exe looking for stdlib/klx
		dir := filepath.Dir(exe)
		for i := 0; i < 5; i++ {
			d := filepath.Join(dir, "stdlib", "klx")
			if _, err := os.Stat(d); err == nil {
				return d
			}
			dir = filepath.Dir(dir)
		}
	}
	return ""
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

// Update replaces a document's text and re-parses it. Returns the new Document.
// If version is provided (>= 0) and is older than the existing version, the
// update is rejected and the existing document is returned unchanged.
func (ds *DocumentStore) Update(uri, text string, version int) *Document {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if existing, ok := ds.docs[uri]; ok && version >= 0 && existing.Version > version {
		// Stale update — keep what we have.
		return existing
	}

	doc := NewDocument(uri, text)
	if version >= 0 {
		doc.Version = version
	}
	ds.docs[uri] = doc
	return doc
}

// ApplyChanges applies a sequence of LSP content changes incrementally.
// Each change may be a full-document replace (Range == nil) or a range edit.
// Returns the new Document. Older-version updates are rejected.
func (ds *DocumentStore) ApplyChanges(uri string, version int, changes []TextDocumentContentChange) *Document {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	existing, ok := ds.docs[uri]
	text := ""
	if ok {
		if version >= 0 && existing.Version > version {
			return existing // stale
		}
		text = existing.Text
	}

	for _, ch := range changes {
		if ch.Range == nil {
			// Full-document replace.
			text = ch.Text
		} else {
			text = applyRangeEdit(text, *ch.Range, ch.Text)
		}
	}

	doc := NewDocument(uri, text)
	if version >= 0 {
		doc.Version = version
	}
	ds.docs[uri] = doc
	return doc
}

// applyRangeEdit replaces the substring of `text` covered by `r` with `newText`.
// LSP positions are 0-based (line, character). Out-of-range edits clamp to bounds.
func applyRangeEdit(text string, r Range, newText string) string {
	startOff := positionToOffset(text, r.Start)
	endOff := positionToOffset(text, r.End)
	if startOff > endOff {
		startOff, endOff = endOff, startOff
	}
	if startOff < 0 {
		startOff = 0
	}
	if endOff > len(text) {
		endOff = len(text)
	}
	return text[:startOff] + newText + text[endOff:]
}

// positionToOffset converts an LSP {line, character} pair to a byte offset
// within text. Lines are separated by '\n'.
func positionToOffset(text string, pos Position) int {
	line, col := 0, 0
	for i := 0; i < len(text); i++ {
		if line == pos.Line && col == pos.Character {
			return i
		}
		if text[i] == '\n' {
			line++
			col = 0
			if line > pos.Line {
				return i // ran past target line — clamp
			}
		} else {
			col++
		}
	}
	return len(text)
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
