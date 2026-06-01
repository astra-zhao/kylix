package repl

import (
	"fmt"
	"io"
	"kylix/generator"
	"kylix/lexer"
	"kylix/parser"
	"kylix/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"
)

const (
	prompt         = "kylix> "
	continuePrompt = "...    "
	version        = "0.3.0"
)

// Color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

// historyFile is the path to the persistent history file
var historyFile = filepath.Join(os.TempDir(), ".kylix_repl_history")

// REPL state
type REPL struct {
	line         *liner.State
	history      []string
	declarations []string
	out          io.Writer
	errOut       io.Writer
}

// Start launches the interactive REPL
func Start(in io.Reader, out io.Writer) error {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	// Load persistent history
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	repl := &REPL{
		line:         line,
		history:      []string{},
		declarations: []string{},
		out:          out,
		errOut:       os.Stderr,
	}

	fmt.Fprintln(out, colorCyan+"Kylix REPL v"+version+colorReset)
	fmt.Fprintln(out, "Type "+colorYellow+":help"+colorReset+" for help, "+colorYellow+":quit"+colorReset+" to exit")
	fmt.Fprintln(out, "Use "+colorYellow+"↑/↓ arrows"+colorReset+" for history navigation")
	fmt.Fprintln(out)

	var buffer strings.Builder
	multiline := false

	for {
		var p string
		if multiline {
			p = colorYellow + continuePrompt + colorReset
		} else {
			p = colorGreen + prompt + colorReset
		}

		input, err := line.Prompt(p)
		if err != nil {
			if err == liner.ErrPromptAborted {
				// Ctrl-C: cancel current multiline input
				if multiline {
					multiline = false
					buffer.Reset()
					fmt.Fprintln(out, colorYellow+"Multiline input cancelled"+colorReset)
					continue
				}
				// Ctrl-C on empty prompt: exit
				fmt.Fprintln(out, colorCyan+"Goodbye!"+colorReset)
				break
			}
			if err == io.EOF {
				fmt.Fprintln(out, colorCyan+"Goodbye!"+colorReset)
				break
			}
			break
		}

		line.AppendHistory(input)

		// Handle meta-commands (only when not in multiline mode)
		if !multiline && strings.HasPrefix(strings.TrimSpace(input), ":") {
			if repl.handleMetaCommand(strings.TrimSpace(input)) {
				break
			}
			continue
		}

		// Empty line in multiline mode - check if complete
		if multiline && input == "" {
			code := buffer.String()
			depth := repl.countBlockDepth(code)
			if depth == 0 {
				buffer.Reset()
				multiline = false
				repl.addToHistory(code)
				repl.execute(code)
				fmt.Fprintln(out)
			} else {
				buffer.WriteString("\n")
			}
			continue
		}

		// Single-line quick execute
		if !multiline && repl.isCompleteStatement(input) {
			repl.addToHistory(input)
			repl.execute(input)
			fmt.Fprintln(out)
			continue
		}

		// Start or continue multiline
		if !multiline {
			multiline = true
		}
		buffer.WriteString(input)
		buffer.WriteString("\n")

		// Auto-execute if block depth is balanced
		code := buffer.String()
		depth := repl.countBlockDepth(code)
		if depth == 0 && strings.Contains(strings.ToLower(code), "begin") {
			buffer.Reset()
			multiline = false
			repl.addToHistory(code)
			repl.execute(code)
			fmt.Fprintln(out)
		}
	}

	// Save history
	if f, err := os.Create(historyFile); err == nil {
		line.WriteHistory(f)
		f.Close()
	}

	return nil
}

func (r *REPL) addToHistory(code string) {
	code = strings.TrimSpace(code)
	if code != "" && (len(r.history) == 0 || r.history[len(r.history)-1] != code) {
		r.history = append(r.history, code)
	}
}

// isCompleteStatement uses lexer-based analysis to determine if input is complete
func (r *REPL) isCompleteStatement(line string) bool {
	line = strings.TrimSpace(line)

	// Meta-commands
	if strings.HasPrefix(line, ":") {
		return true
	}

	// Empty line
	if line == "" {
		return false
	}

	// Lexer-based analysis: check if the line forms a complete statement
	l := lexer.New(line)
	tokens := []token.Token{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	// Check for begin/end blocks
	depth := 0
	for _, tok := range tokens {
		if tok.Type == token.BEGIN {
			depth++
		} else if tok.Type == token.END {
			depth--
		}
	}
	// If has begin/end and balanced, it's complete
	if depth == 0 && len(tokens) > 2 {
		hasBegin := false
		for _, tok := range tokens {
			if tok.Type == token.BEGIN {
				hasBegin = true
				break
			}
		}
		if hasBegin {
			return true
		}
	}

	// Try to parse as a complete program
	testCode := "program test;\nbegin\n" + line + "\nend."
	pl := lexer.New(testCode)
	p := parser.New(pl)
	program := p.ParseProgram()

	if len(p.Errors()) == 0 && len(program.Statements) > 0 {
		return true
	}

	// Lexer-based keyword detection for single-line statements
	if len(tokens) >= 2 {
		lastNonEOF := tokens[len(tokens)-1]
		if lastNonEOF.Type == token.EOF && len(tokens) >= 3 {
			lastNonEOF = tokens[len(tokens)-2]
		}

		first := tokens[0]

		// Statement ends with semicolon and starts with a known keyword
		if lastNonEOF.Type == token.SEMICOLON {
			switch first.Type {
			case token.VAR, token.CONST, token.TYPE,
				token.BREAK, token.CONTINUE, token.RETURN,
				token.IDENT:
				return true
			}
			// Also complete if starts with a known function call pattern
			if first.Type == token.IDENT && len(tokens) >= 4 {
				return true
			}
		}

		// Ends with "end."
		if lastNonEOF.Type == token.DOT && len(tokens) >= 3 {
			prev := tokens[len(tokens)-2]
			if prev.Type == token.EOF {
				prev = tokens[len(tokens)-3]
			}
			if prev.Type == token.END {
				return true
			}
		}
	}

	return false
}

// countBlockDepth uses lexer to count begin/end tokens (AST-based)
func (r *REPL) countBlockDepth(code string) int {
	l := lexer.New(code)
	depth := 0

	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.BEGIN {
			depth++
		} else if tok.Type == token.END {
			depth--
		}
	}

	return depth
}

func (r *REPL) handleMetaCommand(cmd string) bool {
	parts := strings.Fields(cmd)

	switch parts[0] {
	case ":help", ":h", ":?":
		r.printHelp()
		return false

	case ":quit", ":q", ":exit":
		fmt.Fprintln(r.out, colorCyan+"Goodbye!"+colorReset)
		return true

	case ":clear", ":c":
		fmt.Fprint(r.out, "\033[2J\033[H")
		return false

	case ":history":
		r.printHistory()
		return false

	case ":decls", ":declarations":
		r.printDeclarations()
		return false

	case ":reset":
		r.declarations = []string{}
		fmt.Fprintln(r.out, colorGreen+"✓ Declarations reset"+colorReset)
		return false

	case ":save":
		if len(parts) < 2 {
			fmt.Fprintln(r.out, colorYellow+"Usage: :save <filename>"+colorReset)
		} else {
			r.saveSession(parts[1])
		}
		return false

	case ":version", ":v":
		fmt.Fprintf(r.out, "Kylix REPL v%s\n", version)
		return false

	default:
		fmt.Fprintf(r.errOut, colorRed+"Unknown command: %s"+colorReset+"\n", parts[0])
		fmt.Fprintln(r.out, "Type "+colorYellow+":help"+colorReset+" for available commands")
		return false
	}
}

func (r *REPL) printHelp() {
	fmt.Fprintln(r.out, colorCyan+"\nAvailable Commands:"+colorReset)
	fmt.Fprintln(r.out)
	fmt.Fprintln(r.out, "  "+colorYellow+":help, :h, :?"+colorReset+"     Show this help message")
	fmt.Fprintln(r.out, "  "+colorYellow+":quit, :q, :exit"+colorReset+"  Exit the REPL")
	fmt.Fprintln(r.out, "  "+colorYellow+":clear, :c"+colorReset+"        Clear the screen")
	fmt.Fprintln(r.out, "  "+colorYellow+":history"+colorReset+"          Show command history")
	fmt.Fprintln(r.out, "  "+colorYellow+":decls"+colorReset+"            Show accumulated declarations")
	fmt.Fprintln(r.out, "  "+colorYellow+":reset"+colorReset+"            Reset all declarations")
	fmt.Fprintln(r.out, "  "+colorYellow+":save <file>"+colorReset+"       Save session to file")
	fmt.Fprintln(r.out, "  "+colorYellow+":version, :v"+colorReset+"      Show version information")
	fmt.Fprintln(r.out)
	fmt.Fprintln(r.out, colorCyan+"Usage:"+colorReset)
	fmt.Fprintln(r.out, "  • Type Kylix code directly")
	fmt.Fprintln(r.out, "  • Use "+colorYellow+"↑/↓ arrows"+colorReset+" to navigate history")
	fmt.Fprintln(r.out, "  • Press Enter twice to execute multiline code")
	fmt.Fprintln(r.out, "  • Press "+colorYellow+"Ctrl-C"+colorReset+" to cancel multiline input")
	fmt.Fprintln(r.out, "  • Declarations (var, function, type) accumulate across sessions")
	fmt.Fprintln(r.out, "  • Use "+colorYellow+":decls"+colorReset+" to see what's been declared")
	fmt.Fprintln(r.out)
}

func (r *REPL) printHistory() {
	if len(r.history) == 0 {
		fmt.Fprintln(r.out, colorYellow+"No history yet"+colorReset)
		return
	}

	fmt.Fprintln(r.out, colorCyan+"\nCommand History:"+colorReset)
	for i, cmd := range r.history {
		fmt.Fprintf(r.out, "  "+colorBlue+"%3d:"+colorReset+" %s\n", i+1, strings.ReplaceAll(cmd, "\n", " "))
	}
	fmt.Fprintln(r.out)
}

func (r *REPL) printDeclarations() {
	if len(r.declarations) == 0 {
		fmt.Fprintln(r.out, colorYellow+"No declarations yet"+colorReset)
		return
	}

	fmt.Fprintln(r.out, colorCyan+"\nAccumulated Declarations:"+colorReset)
	fmt.Fprintln(r.out)
	for _, decl := range r.declarations {
		fmt.Fprintln(r.out, colorGreen+decl+colorReset)
	}
	fmt.Fprintln(r.out)
}

func (r *REPL) saveSession(filename string) {
	var content strings.Builder

	// Write as a proper Kylix program
	content.WriteString("program SavedSession;\n\n")

	if len(r.declarations) > 0 {
		for _, decl := range r.declarations {
			content.WriteString(decl)
			content.WriteString("\n\n")
		}
	}

	// Write history as comments
	if len(r.history) > 0 {
		content.WriteString("// Command History\n")
		for i, cmd := range r.history {
			content.WriteString(fmt.Sprintf("// %d: %s\n", i+1, strings.ReplaceAll(cmd, "\n", "\\n")))
		}
	}

	err := os.WriteFile(filename, []byte(content.String()), 0644)
	if err != nil {
		fmt.Fprintf(r.errOut, colorRed+"Error saving session: %v"+colorReset+"\n", err)
	} else {
		fmt.Fprintf(r.out, colorGreen+"✓ Session saved to %s"+colorReset+"\n", filename)
	}
}

func (r *REPL) execute(code string) {
	// Build complete program with accumulated declarations
	var fullCode strings.Builder

	fullCode.WriteString("program repl;\n")

	// Add accumulated declarations
	for _, decl := range r.declarations {
		fullCode.WriteString(decl)
		fullCode.WriteString("\n")
	}

	// Check if this is a declaration (var, function, procedure, type, const)
	lower := strings.ToLower(strings.TrimSpace(code))
	isDeclaration := strings.HasPrefix(lower, "var ") ||
		strings.HasPrefix(lower, "function ") ||
		strings.HasPrefix(lower, "procedure ") ||
		strings.HasPrefix(lower, "type ") ||
		strings.HasPrefix(lower, "const ")

	if isDeclaration {
		// Add to declarations and confirm
		r.declarations = append(r.declarations, code)
		fmt.Fprintf(r.out, colorGreen+"✓ Declaration added"+colorReset+"\n")
		return
	}

	// Otherwise, treat as statement in begin/end block
	if !strings.Contains(code, "begin") && !strings.Contains(code, "program") {
		fullCode.WriteString("begin\n")
		fullCode.WriteString("  ")
		fullCode.WriteString(strings.ReplaceAll(code, "\n", "\n  "))
		fullCode.WriteString("\nend.")
	} else {
		fullCode.WriteString(code)
	}

	// Lex and parse
	l := lexer.New(fullCode.String())
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(r.errOut, colorRed+"Errors:"+colorReset)
		for _, err := range p.Errors() {
			fmt.Fprintf(r.errOut, "  "+colorRed+"✗ %s"+colorReset+"\n", err)
		}
		return
	}

	// Generate Go code
	gen := generator.New()
	goCode := gen.Generate(program)

	// Write to temp file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "kylix_repl.go")
	if err := os.WriteFile(tmpFile, []byte(goCode), 0644); err != nil {
		fmt.Fprintf(r.errOut, colorRed+"Error writing temp file: %v"+colorReset+"\n", err)
		return
	}
	defer os.Remove(tmpFile)

	// Run with go run — stdout goes to r.out, stderr goes to r.errOut
	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = r.out
	cmd.Stderr = r.errOut
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(r.errOut, colorRed+"Runtime error: %v"+colorReset+"\n", err)
	}
}
