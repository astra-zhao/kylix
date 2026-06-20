// docgen.go — Extract doc comments from Kylix source files and generate Markdown.
//
// Doc comment convention:
//
//	// Any comment immediately preceding a declaration is its documentation.
//	// Multi-line comments are concatenated.
//
// Output format: one Markdown file per unit.
package docgen

import (
	"fmt"
	"kylix/ast"
	"kylix/lexer"
	"kylix/parser"
	"os"
	"path/filepath"
	"strings"
)

// DocEntry holds documentation for a single symbol.
type DocEntry struct {
	Kind      string // "function", "procedure", "class", "interface", "type", "const", "var"
	Name      string
	Signature string
	Comment   string
}

// UnitDoc holds all documentation for one unit/program.
type UnitDoc struct {
	Name    string
	Comment string
	Entries []DocEntry
}

// GenerateFile parses a .klx file and returns a UnitDoc.
func GenerateFile(path string) (*UnitDoc, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	comments := extractComments(string(src))

	l := lexer.New(string(src))
	p := parser.New(l)
	prog := p.ParseProgram()

	doc := &UnitDoc{
		Name:    unitName(prog, path),
		Comment: unitComment(comments),
	}

	doc.Entries = extractEntries(prog, comments)
	return doc, nil
}

// RenderMarkdown converts a UnitDoc to a Markdown string.
func RenderMarkdown(doc *UnitDoc) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", doc.Name))
	if doc.Comment != "" {
		sb.WriteString(doc.Comment + "\n\n")
	}

	// Group entries by kind
	kinds := []string{"type", "const", "var", "function", "procedure", "class", "interface"}
	headers := map[string]string{
		"type": "Types", "const": "Constants", "var": "Variables",
		"function": "Functions", "procedure": "Procedures",
		"class": "Classes", "interface": "Interfaces",
	}

	for _, kind := range kinds {
		var section []DocEntry
		for _, e := range doc.Entries {
			if e.Kind == kind {
				section = append(section, e)
			}
		}
		if len(section) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("## %s\n\n", headers[kind]))
		for _, e := range section {
			sb.WriteString(fmt.Sprintf("### %s\n\n", e.Name))
			sb.WriteString("```pascal\n")
			sb.WriteString(e.Signature + "\n")
			sb.WriteString("```\n\n")
			if e.Comment != "" {
				sb.WriteString(e.Comment + "\n\n")
			}
		}
	}

	return sb.String()
}

// ── internal helpers ──────────────────────────────────────────────────────────

// extractComments builds a map from line number (1-based) → comment text.
// Only single-line comments (//) are supported; blank lines break comment blocks.
func extractComments(src string) map[int]string {
	result := make(map[int]string)
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			text := strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
			result[i+1] = text // 1-based line number
		}
	}
	return result
}

// commentBefore returns the block of comment lines immediately preceding line.
// Lines are joined with newlines (not spaces) so fenced code blocks are preserved.
func commentBefore(comments map[int]string, line int) string {
	var parts []string
	for l := line - 1; l >= 1; l-- {
		if c, ok := comments[l]; ok {
			parts = append([]string{c}, parts...)
		} else {
			break
		}
	}
	return strings.Join(parts, "\n")
}

// unitName returns the name of the unit or the file basename.
func unitName(prog *ast.Program, path string) string {
	if prog.UnitName != "" {
		return prog.UnitName
	}
	if prog.Name != "" {
		return prog.Name
	}
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// unitComment extracts the file-level comment (lines 1–N before the first token).
// Lines are joined with newlines to preserve fenced code blocks.
func unitComment(comments map[int]string) string {
	var parts []string
	for l := 1; l <= 10; l++ {
		if c, ok := comments[l]; ok {
			parts = append(parts, c)
		} else if len(parts) > 0 {
			break
		}
	}
	return strings.Join(parts, "\n")
}

// extractEntries walks the AST and generates DocEntry for each declaration.
func extractEntries(prog *ast.Program, comments map[int]string) []DocEntry {
	var entries []DocEntry

	for _, decl := range prog.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			kind := "function"
			if d.ReturnType == nil && len(d.ReturnTypes) == 0 {
				kind = "procedure"
			}
			entries = append(entries, DocEntry{
				Kind:      kind,
				Name:      d.Name,
				Signature: funcSignature(d),
				Comment:   commentBefore(comments, d.Token.Line),
			})

		case *ast.TypeDecl:
			kind := "type"
			if _, ok := d.Type.(*ast.ClassDecl); ok {
				kind = "class"
			} else if _, ok := d.Type.(*ast.InterfaceDecl); ok {
				kind = "interface"
			}
			entries = append(entries, DocEntry{
				Kind:      kind,
				Name:      d.Name,
				Signature: fmt.Sprintf("type %s = ...", d.Name),
				Comment:   commentBefore(comments, d.Token.Line),
			})

		case *ast.VarDecl:
			if len(d.Names) > 0 {
				typeStr := ""
				if d.Type != nil {
					typeStr = typeExprStr(d.Type)
				}
				for _, name := range d.Names {
					entries = append(entries, DocEntry{
						Kind:      "var",
						Name:      name,
						Signature: fmt.Sprintf("var %s: %s", name, typeStr),
						Comment:   commentBefore(comments, d.Token.Line),
					})
				}
			}

		case *ast.ConstDecl:
			entries = append(entries, DocEntry{
				Kind:      "const",
				Name:      d.Name,
				Signature: fmt.Sprintf("const %s = ...", d.Name),
				Comment:   commentBefore(comments, d.Token.Line),
			})
		}
	}

	return entries
}

// funcSignature builds a Pascal-style signature string.
func funcSignature(fd *ast.FunctionDecl) string {
	var sb strings.Builder
	if fd.ReturnType != nil || len(fd.ReturnTypes) > 0 {
		sb.WriteString("function ")
	} else {
		sb.WriteString("procedure ")
	}
	sb.WriteString(fd.Name)

	if len(fd.Parameters) > 0 {
		sb.WriteString("(")
		for i, p := range fd.Parameters {
			if i > 0 {
				sb.WriteString("; ")
			}
			sb.WriteString(p.Name)
			if p.Type != nil {
				sb.WriteString(": ")
				sb.WriteString(typeExprStr(p.Type))
			}
		}
		sb.WriteString(")")
	}

	if fd.ReturnType != nil {
		sb.WriteString(": ")
		sb.WriteString(typeExprStr(fd.ReturnType))
	} else if len(fd.ReturnTypes) > 1 {
		sb.WriteString(": (")
		for i, rt := range fd.ReturnTypes {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(typeExprStr(rt))
		}
		sb.WriteString(")")
	}
	return sb.String()
}

// typeExprStr converts a type expression to a readable string.
func typeExprStr(expr ast.Expression) string {
	if expr == nil {
		return "?"
	}
	switch t := expr.(type) {
	case *ast.Identifier:
		return t.Value
	case *ast.ArrayType:
		elem := typeExprStr(t.ElementType)
		if t.Dynamic {
			return "array of " + elem
		}
		return "array of " + elem
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", typeExprStr(t.KeyType), typeExprStr(t.ValueType))
	case *ast.GenericType:
		return t.Base + "<...>"
	default:
		return fmt.Sprintf("%T", expr)
	}
}
