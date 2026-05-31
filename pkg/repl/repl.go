package repl

import (
	"bufio"
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

// REPL state
type REPL struct {
	history     []string
	declarations []string
	in          io.Reader
	out         io.Writer
}

// Start launches the interactive REPL
func Start(in io.Reader, out io.Writer) error {
	repl := &REPL{
		history:     []string{},
		declarations: []string{},
		in:          in,
		out:         out,
	}

	fmt.Fprintln(out, colorCyan+"Kylix REPL v"+version+colorReset)
	fmt.Fprintln(out, "Type "+colorYellow+":help"+colorReset+" for help, "+colorYellow+":quit"+colorReset+" to exit")
	fmt.Fprintln(out)

	scanner := bufio.NewScanner(in)
	var buffer strings.Builder
	multiline := false

	for {
		if multiline {
			fmt.Fprint(out, colorYellow+continuePrompt+colorReset)
		} else {
			fmt.Fprint(out, colorGreen+prompt+colorReset)
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		// Handle meta-commands (only when not in multiline mode)
		if !multiline && strings.HasPrefix(line, ":") {
			if repl.handleMetaCommand(line) {
				return nil // :quit command
			}
			continue
		}

		// Empty line in multiline mode - check if complete
		if multiline && line == "" {
			code := buffer.String()
			depth := repl.countBlockDepth(code)
			if depth == 0 {
				buffer.Reset()
				multiline = false
				repl.addToHistory(code)
				repl.execute(code, false)
				fmt.Fprintln(out)
			} else {
				buffer.WriteString("\n")
			}
			continue
		}

		// Single-line quick execute
		if !multiline && repl.isCompleteStatement(line) {
			repl.addToHistory(line)
			repl.execute(line, false)
			fmt.Fprintln(out)
			continue
		}

		// Start or continue multiline
		if !multiline {
			multiline = true
		}
		buffer.WriteString(line)
		buffer.WriteString("\n")

		// Auto-execute if block depth is balanced
		code := buffer.String()
		depth := repl.countBlockDepth(code)
		if depth == 0 && strings.Contains(strings.ToLower(code), "begin") {
			buffer.Reset()
			multiline = false
			repl.addToHistory(code)
			repl.execute(code, false)
			fmt.Fprintln(out)
		}
	}

	fmt.Fprintln(out)
	return scanner.Err()
}

func (r *REPL) addToHistory(code string) {
	code = strings.TrimSpace(code)
	if code != "" && (len(r.history) == 0 || r.history[len(r.history)-1] != code) {
		r.history = append(r.history, code)
	}
}

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

	// Try to parse as a complete program to check completeness
	testCode := "program test;\nbegin\n" + line + "\nend."
	l := lexer.New(testCode)
	p := parser.New(l)
	program := p.ParseProgram()

	// If no errors, it's likely complete
	if len(p.Errors()) == 0 && len(program.Statements) > 0 {
		return true
	}

	// Check for common complete patterns
	lower := strings.ToLower(line)

	// Single-line statements that are complete
	if strings.HasSuffix(lower, ";") {
		keywords := []string{"writeln", "write", "var ", "const ", "type ", "break", "continue", "return"}
		for _, kw := range keywords {
			if strings.HasPrefix(lower, kw) {
				return true
			}
		}
	}

	// Complete if it ends with end.
	if strings.HasSuffix(lower, "end.") {
		return true
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
	cmd = strings.TrimSpace(cmd)
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
		fmt.Fprintf(r.out, colorRed+"Unknown command: %s"+colorReset+"\n", parts[0])
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
	fmt.Fprintln(r.out, "  • Press Enter twice to execute multiline code")
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

	// Write declarations
	if len(r.declarations) > 0 {
		content.WriteString("// Declarations\n")
		for _, decl := range r.declarations {
			content.WriteString(decl)
			content.WriteString("\n\n")
		}
	}

	// Write history as comments
	if len(r.history) > 0 {
		content.WriteString("\n// Command History\n")
		for i, cmd := range r.history {
			content.WriteString(fmt.Sprintf("// %d: %s\n", i+1, strings.ReplaceAll(cmd, "\n", "\\n")))
		}
	}

	err := os.WriteFile(filename, []byte(content.String()), 0644)
	if err != nil {
		fmt.Fprintf(r.out, colorRed+"Error saving session: %v"+colorReset+"\n", err)
	} else {
		fmt.Fprintf(r.out, colorGreen+"✓ Session saved to %s"+colorReset+"\n", filename)
	}
}

func (r *REPL) execute(code string, isDecl bool) {
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
		fmt.Fprintln(r.out, colorRed+"Errors:"+colorReset)
		for _, err := range p.Errors() {
			fmt.Fprintf(r.out, "  "+colorRed+"✗ %s"+colorReset+"\n", err)
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
		fmt.Fprintf(r.out, colorRed+"Error writing temp file: %v"+colorReset+"\n", err)
		return
	}
	defer os.Remove(tmpFile)

	// Run with go run
	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = r.out
	cmd.Stderr = r.out
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(r.out, colorRed+"Runtime error: %v"+colorReset+"\n", err)
	}
}
