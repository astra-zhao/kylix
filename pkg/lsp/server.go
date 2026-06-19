// server.go — LSP server core: JSON-RPC transport, message dispatch, document sync.
package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Server implements a minimal LSP server over stdin/stdout.
type Server struct {
	in  io.Reader
	out io.Writer

	// document store: URI → Document (AST, symbols, diagnostics)
	docs *DocumentStore
}

// New creates an LSP server that reads from in and writes to out.
func New(in io.Reader, out io.Writer) *Server {
	return &Server{in: in, out: out, docs: NewDocumentStore()}
}

// Run is the main loop: reads JSON-RPC messages and dispatches them.
func (s *Server) Run() error {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if response := s.handleMessage(msg); response != nil {
			if err := s.writeMessage(response); err != nil {
				return err
			}
		}
	}
}

// ── JSON-RPC types ────────────────────────────────────────────────────────────

// Message represents a JSON-RPC 2.0 message.
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError is the JSON-RPC error object.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ── LSP common types ──────────────────────────────────────────────────────────

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeParams struct {
	TextDocument   VersionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []TextDocumentContentChange     `json:"contentChanges"`
}

type TextDocumentContentChange struct {
	Range       *Range `json:"range,omitempty"`       // optional: nil means full-document replace
	RangeLength int    `json:"rangeLength,omitempty"` // optional: deprecated by LSP but still seen
	Text        string `json:"text"`
}

type DidCloseParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
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
	Source   string `json:"source,omitempty"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// ── Transport ─────────────────────────────────────────────────────────────────

func (s *Server) readMessage() (*Message, error) {
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

// ── Dispatch ──────────────────────────────────────────────────────────────────

func (s *Server) handleMessage(msg *Message) *Message {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return nil
	case "shutdown":
		return &Message{ID: msg.ID, Result: nil}
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
				ID:    msg.ID,
				Error: &ResponseError{Code: -32601, Message: fmt.Sprintf("method not found: %s", msg.Method)},
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
				"textDocumentSync": 2, // 2 = Incremental sync (was 1 = Full)
				"completionProvider": map[string]interface{}{
					"triggerCharacters": []string{".", ":"},
					"resolveProvider":   false,
				},
				"hoverProvider":              true,
				"definitionProvider":         true,
				"documentSymbolProvider":     true,
				"referencesProvider":         true,
				"renameProvider":             true,
				"documentFormattingProvider": true,
				"signatureHelpProvider": map[string]interface{}{
					"triggerCharacters": []string{"(", ","},
				},
				"codeActionProvider":      true,
				"workspaceSymbolProvider": true,
			},
			"serverInfo": map[string]interface{}{
				"name":    "kylix-lsp",
				"version": "1.2.2",
			},
		},
	}
}

// ── Document sync ─────────────────────────────────────────────────────────────

func (s *Server) handleDidOpen(msg *Message) {
	var params DidOpenParams
	json.Unmarshal(msg.Params, &params)
	doc := s.docs.Update(params.TextDocument.URI, params.TextDocument.Text, params.TextDocument.Version)
	s.publishDiagnostics(doc)
}

func (s *Server) handleDidChange(msg *Message) {
	var params DidChangeParams
	json.Unmarshal(msg.Params, &params)
	// Apply all changes incrementally in a single pass — avoids multiple parses
	// per didChange and ensures atomic version bookkeeping.
	doc := s.docs.ApplyChanges(params.TextDocument.URI, params.TextDocument.Version, params.ContentChanges)
	if doc != nil {
		s.publishDiagnostics(doc)
	}
}

func (s *Server) handleDidClose(msg *Message) {
	var params DidCloseParams
	json.Unmarshal(msg.Params, &params)
	s.docs.Delete(params.TextDocument.URI)
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

// ── Helpers ───────────────────────────────────────────────────────────────────

// parseLocation extracts (line, col) from an error message of the form
// "... (line N, column M)". Returns 0-based indices for LSP.
func parseLocation(msg string) (int, int) {
	var line, col int
	if idx := strings.Index(msg, "(line "); idx >= 0 {
		fmt.Sscanf(msg[idx:], "(line %d, column %d)", &line, &col)
	}
	return line - 1, col - 1
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
